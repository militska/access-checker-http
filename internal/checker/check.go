package checker

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strconv"
	"sync"
	"time"
)

type HttpChecker struct {
	addrCh     chan string
	client     *http.Client
	fileResult *os.File
}

func NewHttpChecker() HttpChecker {
	filename := "result/result_" + strconv.Itoa(int(time.Now().UnixMicro()))
	resultFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	return HttpChecker{
		addrCh:     make(chan string, 20),
		client:     &http.Client{Timeout: 5 * time.Second},
		fileResult: resultFile,
	}
}

func (c *HttpChecker) CloseResultFile() {
	if err := c.fileResult.Close(); err != nil {
		log.Println(err)
	}
}

func (c *HttpChecker) SetData(sourceFilePath string, wg *sync.WaitGroup) {
	file, err := os.Open(sourceFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	defer wg.Done()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c.addrCh <- scanner.Text()
		wg.Add(1)
		if err != nil {
			log.Println(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func (c *HttpChecker) Exec(wg *sync.WaitGroup) {
	for elem := range c.addrCh {
		go c.execInternal(elem, wg)
	}

}

func (c *HttpChecker) execInternal(addr string, wg *sync.WaitGroup) {
	text, err := sendRequest(addr, c.client)
	if err != nil {
		log.Println(err)
	}

	_, err = c.fileResult.Write([]byte(text))
	if err != nil {
		log.Println(err)
	}
	wg.Done()
}

func sendRequest(addr string, client *http.Client) (string, error) {
	var realIp string

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return "", err
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			realIp = connInfo.Conn.RemoteAddr().String()
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %d %s\n", addr, res.StatusCode, realIp), nil
}
