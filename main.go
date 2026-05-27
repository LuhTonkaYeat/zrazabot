package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
	_ "modernc.org/sqlite"
)

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

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	initDB()

	b.Handle("/zraza", func(c tele.Context) error {
		userID := c.Sender().ID
		eaten := rand.Intn(10) + 1

		addZrazy(userID, eaten)
		total := getTotal(userID)

		return c.Send(
			fmt.Sprintf("🍽 %s съел %d зразу(ы)!\n📊 *Всего за историю:* %d зраз.",
				c.Sender().FirstName, eaten, total),
			tele.ModeMarkdown,
		)
	})

	log.Println("Бот запущен! Напиши /zraza в Telegram")
	b.Start()
}

func initDB() {
	db, err := sql.Open("sqlite", "zrazy.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			total INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func addZrazy(userID int64, amount int) {
	db, err := sql.Open("sqlite", "zrazy.db")
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, total) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET total = total + ?
	`, userID, amount, amount)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getTotal(userID int64) int {
	db, err := sql.Open("sqlite", "zrazy.db")
	if err != nil {
		log.Println("DB error:", err)
		return 0
	}
	defer db.Close()

	var total int
	err = db.QueryRow("SELECT total FROM users WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		log.Println("DB error:", err)
		return 0
	}
	return total
}
