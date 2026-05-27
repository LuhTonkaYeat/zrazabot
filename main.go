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

var customSuffixes = map[int64]string{
	1137760134: " (уплетал за обе щеки и тяжку сделал)",
	1005685864: " (и ягером запил все нах)",
	2035294142: " (и балтосом 1 запил все нax)",
	1966955912: " (по трезвяку спокойно наслаждался...)",
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
			return c.Send(fmt.Sprintf("⏰ _%s, сначала нагуляй аппетyeat!!!_\n_Попробуй еще раз примерно через час_\n\n🍽 /zraza", userName), tele.ModeMarkdown)
		}

		if rand.Intn(10) == 0 {
			phrase := fmt.Sprintf(rarePhrases[0], userName)
			addShit(userID, userName, 1)
			updateLastUsed(userID, now)
			shitTotal := getShitTotal(userID)
			phrase += fmt.Sprintf("\n\n💩 _Всего говна навернуто: %d_", shitTotal)
			return c.Send(phrase, tele.ModeMarkdown)
		}

		eaten := rand.Intn(10) + 1
		garnish := garnishes[rand.Intn(len(garnishes))]
		addZrazy(userID, userName, eaten)
		total := getTotal(userID)
		updateLastUsed(userID, now)

		var zrazaWord string
		if total%10 == 1 && total%100 != 11 {
			zrazaWord = "зраза"
		} else if (total%10 >= 2 && total%10 <= 4) && (total%100 < 10 || total%100 >= 20) {
			zrazaWord = "зразы"
		} else {
			zrazaWord = "зраз"
		}

		message := fmt.Sprintf(
			"_%s только что сожрал %d зраз и %s!!!_\n📊 _А всего им уничтожено - %d %s!_\n\n🍽 _Голоден? /zraza_",
			userName, eaten, garnish, total, zrazaWord,
		)

		return c.Send(message, tele.ModeMarkdown)
	})

	b.Handle("/zrazastat", func(c tele.Context) error {
		users := getLeaderboard(5)
		if len(users) == 0 {
			return c.Send("_Пока никто не ел зразы... Напиши /zraza_", tele.ModeMarkdown)
		}

		message := "_🏆 Легенды столешницы СОШ №1 по финансовым махинациям со зразами:_\n\n"
		for i, u := range users {
			message += fmt.Sprintf("%d. _%s_ - _%d_\n", i+1, u.name, u.total)
		}

		shitLeaders := getShitLeaderboard(3)
		if len(shitLeaders) > 0 {
			message += "\n_💩 Антигерои (говноеды) за всё время:_\n"
			for i, s := range shitLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d раз(а)_\n", i+1, s.name, s.total)
			}
		}

		return c.Send(message, tele.ModeMarkdown)
	})

	log.Println("Бот запущен! Напиши /zraza в Telegram")
	b.Start()
}

type userStats struct {
	name  string
	total int
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
			user_name TEXT DEFAULT '',
			total INTEGER DEFAULT 0,
			last_used INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`ALTER TABLE users ADD COLUMN shit_total INTEGER DEFAULT 0`)
	if err != nil {
		log.Println("Column shit_total already exists or migration skipped:", err)
	}
}

func addZrazy(userID int64, userName string, amount int) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, user_name, total) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET 
			total = total + ?,
			user_name = EXCLUDED.user_name
	`, userID, userName, amount, amount)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func addShit(userID int64, userName string, amount int) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, user_name, shit_total) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET 
			shit_total = shit_total + ?,
			user_name = EXCLUDED.user_name
	`, userID, userName, amount, amount)
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

func getShitTotal(userID int64) int {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return 0
	}
	defer db.Close()

	var shitTotal int
	err = db.QueryRow("SELECT shit_total FROM users WHERE user_id = ?", userID).Scan(&shitTotal)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		log.Println("DB error:", err)
		return 0
	}
	return shitTotal
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
		ON CONFLICT(user_id) DO UPDATE SET last_used = EXCLUDED.last_used
	`, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getLeaderboard(limit int) []userStats {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT user_name, total FROM users 
		WHERE total > 0 
		ORDER BY total DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()

	var users []userStats
	for rows.Next() {
		var u userStats
		if err := rows.Scan(&u.name, &u.total); err != nil {
			log.Println("DB error:", err)
			continue
		}
		users = append(users, u)
	}

	return users
}

func getShitLeaderboard(limit int) []userStats {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT user_id, user_name, shit_total FROM users 
		WHERE shit_total > 0 
		ORDER BY shit_total DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()

	var users []userStats
	for rows.Next() {
		var u userStats
		var userID int64
		if err := rows.Scan(&userID, &u.name, &u.total); err != nil {
			log.Println("DB error:", err)
			continue
		}
		if suffix, ok := customSuffixes[userID]; ok {
			u.name = u.name + suffix
		}
		users = append(users, u)
	}

	return users
}
