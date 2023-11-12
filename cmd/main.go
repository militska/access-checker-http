package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
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
	//
	//f_result, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	//if err != nil {
	//	panic(err)
	//}

	defer f_result.Close()

	file, err := os.Open("source")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		res, err := http.Get("http://" + scanner.Text())
		if err != nil {
			log.Println(err)
		}

		_, err = f_result.Write([]byte(scanner.Text() + "  " + strconv.Itoa(res.StatusCode) + res.Request.RemoteAddr + "\n"))
		if err != nil {
			panic(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
