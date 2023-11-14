package main

import (
	checkerPkg "access-checker-http/internal/checker"
	"flag"
	"log"
	"sync"
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

const DefaultSourcePath = "source"

var sourceFilePath string

func init() {
	flag.StringVar(&sourceFilePath, "sourceFilePath", DefaultSourcePath, "Source file path")
}

func main() {
	checker := checkerPkg.NewHttpChecker()

	var wg sync.WaitGroup
	wg.Add(1)

	go checker.SetData(sourceFilePath, &wg)
	go checker.Exec(&wg)

	wg.Wait()

	log.Println("Shutting down app...")
}
