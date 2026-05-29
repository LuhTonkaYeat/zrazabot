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

var customSuffixes = map[int64]string{
	1137760134: " (уплетал за обе щеки и тяжку сделал)",
	1005685864: " (ягером запил все нах)",
	2035294142: " (балтосом 1 запил все нax)",
	1966955912: " (по трезвяку спокойно наслаждался...)",
}

func getDBPath() string {
	dataDir := "/app/data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Println("Warning: cannot create data dir:", err)
	}
	return filepath.Join(dataDir, "zrazy.db")
}

func formatCooldown(secondsLeft int64) string {
	hours := secondsLeft / 3600
	minutes := (secondsLeft % 3600) / 60
	secs := secondsLeft % 60

	if hours > 0 {
		return fmt.Sprintf("%dч %dмин %dс", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dмин %dс", minutes, secs)
	}
	return fmt.Sprintf("%dс", secs)
}

func formatZrazyCount(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "зраза"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "зразы"
	}
	return "зраз"
}

func sendToTopic(b *tele.Bot, c tele.Context, text string) error {
	msg := c.Message()
	chat := msg.Chat
	topicID := msg.ThreadID

	opt := &tele.SendOptions{
		ParseMode:             tele.ModeMarkdown,
		ReplyTo:               msg,
		ThreadID:              topicID,
		DisableWebPagePreview: true,
	}

	_, err := b.Send(chat, text, opt)
	return err
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
			secondsLeft := 3600 - (now - lastUsed)
			timeLeft := formatCooldown(secondsLeft)
			return sendToTopic(b, c, fmt.Sprintf("⏰ _%s, сначала нагуляй аппетyeat!!!_\n_Осталось ждать: %s_\n\n🍽 /zraza", userName, timeLeft))
		}

		rarity := rand.Intn(100)

		if rarity < 3 {
			addZrazy(userID, userName, 67)
			total := getTotal(userID)
			updateLastUsed(userID, now)
			garnish := garnishes[rand.Intn(len(garnishes))]
			message := fmt.Sprintf(
				"_✨✨✨ ЧУДО! ЧЗХХХ!!! ✨✨✨_\n_%s нашел заначку и сожрал 67 зраз с %s!!!_\n📊 _А всего им уничтожено - %d %s!_\n\n🍽 _Голоден? /zraza_",
				userName, garnish, total, formatZrazyCount(total),
			)
			return sendToTopic(b, c, message)
		}

		if rarity < 10 {
			resetZrazy(userID)
			addShit(userID, userName, 1)
			updateLastUsed(userID, now)
			shitTotal := getShitTotal(userID)
			phrase := fmt.Sprintf("💩 _%s навернул говнеца и обнулил свой счётчик зраз!_\n🍽 Голоден? /zraza", userName)
			phrase += fmt.Sprintf("\n\n💩 _Всего говна навернуто: %d_", shitTotal)
			return sendToTopic(b, c, phrase)
		}

		eaten := rand.Intn(10) + 1
		garnish := garnishes[rand.Intn(len(garnishes))]
		addZrazy(userID, userName, eaten)
		total := getTotal(userID)
		updateLastUsed(userID, now)

		message := fmt.Sprintf(
			"_%s только что сожрал %d зраз и %s!!!_\n📊 _А всего им уничтожено - %d %s!_\n\n🍽 _Голоден? /zraza_",
			userName, eaten, garnish, total, formatZrazyCount(total),
		)

		return sendToTopic(b, c, message)
	})

	b.Handle("/zrazastat", func(c tele.Context) error {
		users := getLeaderboard(5)
		if len(users) == 0 {
			return sendToTopic(b, c, "_Пока никто не ел зразы... Напиши /zraza_")
		}

		message := "_🏆 Легенды столешницы СОШ №1 по финансовым махинациям со зразами:_\n\n"
		for i, u := range users {
			message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, u.name, u.total, formatZrazyCount(u.total))
		}

		shitLeaders := getShitLeaderboard(5)
		if len(shitLeaders) > 0 {
			message += "\n_💩 Топ говноедов за всё время:_\n"
			for i, s := range shitLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d раз(а)_%s\n", i+1, s.name, s.total, s.suffix)
			}
		}

		return sendToTopic(b, c, message)
	})

	b.Handle("/kisel", func(c tele.Context) error {
		userName := c.Sender().FirstName
		message := fmt.Sprintf("*%s* _сказал, что КИСЕЛЬ ДАУН_", userName)
		return sendToTopic(b, c, message)
	})

	log.Println("Бот запущен! Напиши /zraza в Telegram")
	b.Start()
}

type userStats struct {
	name   string
	total  int
	suffix string
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

func resetZrazy(userID int64) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		UPDATE users SET total = 0 WHERE user_id = ?
	`, userID)
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
			u.suffix = suffix
		}
		users = append(users, u)
	}

	return users
}
