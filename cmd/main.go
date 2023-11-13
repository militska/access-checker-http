package main

import (
	"bufio"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strconv"
	"sync"
	"time"
)

//Написать сервис проверяющий доступность ресурсов. Сервис работает в фоне 24/7.
//1. Читает ссылки из файла. Ссылки могут добавляться в файл в любой момент времени работы сервиса.
//2. Проверяет доступность ресурсов и получает следующие данные:
//- ip адрес
//- код ответа
//- заголовок из html (в случае если ресурс вернул код 200).
//3. Все результаты работы складываются для отчетности в выходной файл. В конце дня файл с результатами будет изыматься с сервера для последующей обработки.
//4. Если ресурс не ответил за 5 секунд, то считается, что ресурс не работает (фиксируем ошибку 500).
//5. Если ресурс выполнил перенаправление (редирект), то зафиксировать путь перенаправления и код.
//6. Примерная нагрузка 70 000 ссылок. Время разбора: "нужно было сделать уже вчера".
//
//Желаемый формат вывода данных в файл:
//ya.ru | 87.250.250.242 | 200 | Яндекс
//www.ya.ru | ya.ru/redirect | 301 | -

func main() {
	filename := "result/result_" + strconv.Itoa(int(time.Now().UnixMicro()))
	f_result, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer f_result.Close()

	client := &http.Client{}

	addrCh := make(chan string, 4)

	var wg sync.WaitGroup
	wg.Add(1)
	go initCh(addrCh, &wg)

	go func() {
		for elem := range addrCh {
			go exec(elem, f_result, client, &wg)
		}
	}()
	wg.Wait()

	log.Println("Shutting down server...")
}

func initCh(addrCh chan string, wg *sync.WaitGroup) {
	file, err := os.Open("source")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	defer wg.Done()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addrCh <- scanner.Text()
		wg.Add(1)
		if err != nil {
			log.Println(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func exec(addr string, resFile *os.File, client *http.Client, wg *sync.WaitGroup) {
	text, err := sendRequest(addr, client)
	if err != nil {
		log.Println(err)
	}

	_, err = resFile.Write([]byte(text))
	if err != nil {
		log.Println(err)
	}
	wg.Done()
}

func sendRequest(addr string, client *http.Client) (string, error) {
	var rip string

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return "", err
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			rip = connInfo.Conn.RemoteAddr().String()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	return addr + " " + strconv.Itoa(res.StatusCode) + " " + rip + "\n", nil
}
