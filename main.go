package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	tele "gopkg.in/telebot.v3"
)

var b *tele.Bot

func main() {
	rand.Seed(time.Now().UnixNano())

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN not set")
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	var err error
	b, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	initDB()
	registerHandlers(b)

	log.Println("Бот запущен! Напиши /zraza в Telegram")

	// Graceful shutdown
	go func() {
		b.Start()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Получен сигнал завершения. Остановка...")
	b.Stop()
	db.Close()
	log.Println("Бот остановлен.")
}
