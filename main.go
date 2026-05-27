package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	tele "gopkg.in/telebot.v3"
	_ "modernc.org/sqlite"
)

var garnishes = []string{
	"🍝 макароны с подливкой",
	"🍚 рис с подливкой",
	"🥔 пюре с подливкой",
	"🌾 гречку с подливкой",
}

var rarePhrases = []string{
	"💩 _%s только что навернул говнеца! сегодня без зраз!_\n🍽 Голоден? /zraza",
}

func getDBPath() string {
	dataDir := "/app/data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Println("Warning: cannot create data dir:", err)
	}
	return filepath.Join(dataDir, "zrazy.db")
}

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
		userName := c.Sender().FirstName

		lastUsed := getLastUsed(userID)
		now := time.Now().Unix()
		if now-lastUsed < 3600 && lastUsed != 0 {
			return c.Send(fmt.Sprintf("⏰ _*%s*, сначала нагуляй аппетyeat!!!_\n_Попробуй еще раз примерно через час_\n\n🍽 /zraza", userName), tele.ModeMarkdown)
		}

		if rand.Intn(10) == 0 {
			phrase := fmt.Sprintf(rarePhrases[0], "*"+userName+"*")
			updateLastUsed(userID, now)
			return c.Send(phrase, tele.ModeMarkdown)
		}

		eaten := rand.Intn(10) + 1
		garnish := garnishes[rand.Intn(len(garnishes))]
		addZrazy(userID, eaten)
		total := getTotal(userID)
		updateLastUsed(userID, now)

		message := fmt.Sprintf(
			"_*%s* только что сожрал %d зраз и %s!!!_\n📊 _А всего им уничтожено - %d зраз!_\n\n🍽 _Голоден? /zraza_",
			userName, eaten, garnish, total,
		)

		return c.Send(message, tele.ModeMarkdown)
	})

	log.Println("Бот запущен! Напиши /zraza в Telegram")
	b.Start()
}

func initDB() {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			total INTEGER DEFAULT 0,
			last_used INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func addZrazy(userID int64, amount int) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
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
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
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

func getLastUsed(userID int64) int64 {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return 0
	}
	defer db.Close()

	var lastUsed int64
	err = db.QueryRow("SELECT last_used FROM users WHERE user_id = ?", userID).Scan(&lastUsed)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		log.Println("DB error:", err)
		return 0
	}
	return lastUsed
}

func updateLastUsed(userID int64, timestamp int64) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, last_used) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET last_used = ?
	`, userID, timestamp, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}
