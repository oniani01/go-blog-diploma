package logger

import (
	"os"
	"sync"
	"time"
)

var (
	LogChan chan string
	wg      sync.WaitGroup
)

func Init() {
	LogChan = make(chan string, 100)
	wg.Add(1)
	go worker()
}

func worker() {
	defer wg.Done()
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	for msg := range LogChan {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		f.WriteString(timestamp + " " + msg + "\n")
		time.Sleep(1 * time.Second)
	}
}

func LogAction(msg string) {
	select {
	case LogChan <- msg:
	default:
	}
}

func Close() {
	close(LogChan)
	wg.Wait()
}