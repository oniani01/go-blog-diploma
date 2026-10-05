package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogChan - канал для отправки сообщений на логирование
var LogChan chan string

// Init запускает фоновую горутину-воркер для записи логов
func Init() {
	// Создаем буферизированный канал (вмещает до 100 сообщений, чтобы не блокировать сервер)
	LogChan = make(chan string, 100)

	// Запускаем горутину (фоновый процесс)
	go func() {
		// Открываем файл log.txt для дозаписи (O_APPEND). Создаем его, если нет (O_CREATE).
		f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println("Ошибка открытия файла логов:", err)
			return
		}
		defer f.Close()

		// Воркер бесконечно читает сообщения из канала, пока канал не закроют
		for msg := range LogChan {
			// Обязательная задержка 1-2 секунды по заданию
			time.Sleep(1 * time.Second)

			// Записываем сообщение в файл
			_, err := f.WriteString(msg + "\n")
			if err != nil {
				log.Println("Ошибка записи в лог:", err)
			}
			f.Sync() // Принудительно сбрасываем буфер на диск
		}
	}()
}

// Close корректно завершает работу логгера (Graceful shutdown)
func Close() {
	fmt.Println("Завершение работы логгера...")
	close(LogChan)              // Закрываем канал, чтобы цикл for-range в горутине завершился
	time.Sleep(2 * time.Second) // Даем воркеру время дописать последние сообщения
}

// LogAction отправляет сообщение в канал
func LogAction(msg string) {
	select {
	case LogChan <- msg:
		// Сообщение успешно отправлено в канал
	default:
		log.Println("Канал логирования переполнен, сообщение пропущено")
	}
}
