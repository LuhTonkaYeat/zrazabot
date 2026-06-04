package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
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

func formatEatenZrazy(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "зразу"
	}
	return "зраз"
}

func formatLuckyCount(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "раз"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "раза"
	}
	return "раз"
}

func formatShitCount(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "раз"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "раза"
	}
	return "раз"
}

func formatStealCount(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "ограбление"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "ограбления"
	}
	return "ограблений"
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

		if rarity < 5 {
			addZrazy(userID, userName, 67)
			incrementLuckyCount(userID)
			total := getTotal(userID)
			updateLastUsed(userID, now)
			garnish := garnishes[rand.Intn(len(garnishes))]
			message := fmt.Sprintf(
				"_✨✨✨ ЧУДО! ЧЗХХХ!!! ✨✨✨_\n_%s нашел заначку и сожрал 67 зраз с %s!!!_\n📊 _А всего он схавал - %d %s!_\n\n🍽 _Голоден? /zraza_",
				userName, garnish, total, formatZrazyCount(total),
			)
			return sendToTopic(b, c, message)
		}

		if rarity < 15 {
			resetZrazy(userID)
			addShit(userID, userName, 1)
			updateLastUsed(userID, now)
			shitTotal := getShitTotal(userID)
			phrase := fmt.Sprintf("💩 ХЕХЕХЕХЕ _%s навернул тарелку говнеца и обнулил свой счётчик зраз!_\n🍽 Голоден? /zraza", userName)
			phrase += fmt.Sprintf("\n\n💩 _Всего тарелок говна им навернуто: %d %s_", shitTotal, formatShitCount(shitTotal))
			return sendToTopic(b, c, phrase)
		}

		eaten := rand.Intn(10) + 1
		garnish := garnishes[rand.Intn(len(garnishes))]
		addZrazy(userID, userName, eaten)
		total := getTotal(userID)
		updateLastUsed(userID, now)

		message := fmt.Sprintf(
			"_%s ток что сожрал %d %s и %s!!!_\n📊 _А всего он схавал - %d %s!_\n\n🍽 _Голоден? /zraza_",
			userName, eaten, formatEatenZrazy(eaten), garnish, total, formatZrazyCount(total),
		)

		return sendToTopic(b, c, message)
	})

	b.Handle("/stat", func(c tele.Context) error {
		users := getLeaderboard(5)
		if len(users) == 0 {
			return sendToTopic(b, c, "_Пока никто не ел зразы... Напиши /zraza_")
		}

		message := "🥣 *Актуальный еврейтинг СОШ №1 по поеданию зраз:*\n\n"
		for i, u := range users {
			message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, u.name, u.total, formatZrazyCount(u.total))
		}

		return sendToTopic(b, c, message)
	})

	b.Handle("/top", func(c tele.Context) error {
		message := "_📊 Наша стата:_\n\n"

		shitLeaders := getShitLeaderboard(5)
		if len(shitLeaders) > 0 {
			message += "_💩 Топ говноедов:_\n"
			for i, s := range shitLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, s.name, s.total, formatShitCount(s.total))
			}
			message += "\n"
		}

		luckyLeaders := getLuckyLeaderboard(5)
		if len(luckyLeaders) > 0 {
			message += "_✨ Топ лакеров (67 зраз):_\n"
			for i, l := range luckyLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, l.name, l.count, formatLuckyCount(l.count))
			}
			message += "\n"
		}

		stealLeaders := getStealLeaderboard(5)
		if len(stealLeaders) > 0 {
			message += "_🦝 Топ воров (успешные кражи):_\n"
			for i, st := range stealLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, st.name, st.count, formatStealCount(st.count))
			}
			message += "\n"
		}

		giveLeaders := getGiveLeaderboard(5)
		if len(giveLeaders) > 0 {
			message += "_🎁 Топ донатеров (подаренные зразы):_\n"
			for i, g := range giveLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, g.name, g.total, formatZrazyCount(g.total))
			}
		}

		if len(message) == len("_📊 Наша стата:_\n\n") {
			message += "_Пока стата пустая..._"
		}

		return sendToTopic(b, c, message)
	})

	b.Handle("/kisel", func(c tele.Context) error {
		userName := c.Sender().FirstName
		message := fmt.Sprintf("*%s* _сказал, что КИСЕЛЬ ДАУН_", userName)
		return sendToTopic(b, c, message)
	})

	b.Handle("/steal", func(c tele.Context) error {
		userID := c.Sender().ID
		userName := c.Sender().FirstName
		args := c.Args()

		if len(args) < 1 {
			return sendToTopic(b, c, "🎰 *ДА БУДЕТ ГЕМБЛИНГ*\n\n🎭 *Кража зраз - как юзать:*\n`/steal @username`\n\nПример: `/steal @derden993`")
		}

		targetMention := strings.TrimPrefix(args[0], "@")

		targetUserID, targetName, err := getUserByUsername(targetMention)
		if err != nil {
			return sendToTopic(b, c, fmt.Sprintf("❌ *Ошибка:* %s", err.Error()))
		}

		if targetUserID == userID {
			return sendToTopic(b, c, "_Слышь, умник, я щас мамке твоей пожалуюсь, что ты абузить пытаешься, усёк? 😉_")
		}

		stealCooldown := getStealCooldown(userID)
		now := time.Now().Unix()
		if now-stealCooldown < 3600*5 && stealCooldown != 0 {
			secondsLeft := 3600*5 - (now - stealCooldown)
			timeLeft := formatCooldown(secondsLeft)
			return sendToTopic(b, c, fmt.Sprintf("⏰ Не, %s, додепа не будет!_\n_Попробуй тырнуть через: %s_\n\n🍽 А лучше сам похавай /zraza", userName, timeLeft))
		}

		total := getTotal(userID)
		if total < 5 {
			return sendToTopic(b, c, fmt.Sprintf("⚠️ _Эйоу, обнаружен нищук! (%s) У тя меньше 5 зраз, накопи сначала, а потом уже воруй dayum_\n\n🍽 Cам пожри /zraza", userName))
		}

		targetTotal := getTotal(targetUserID)
		if targetTotal < 3 {
			return sendToTopic(b, c, fmt.Sprintf("😢 _У %s всего %d %s, с него нех взять..._", targetName, targetTotal, formatZrazyCount(targetTotal)))
		}

		stealAmount := rand.Intn(5) + 1
		if stealAmount > targetTotal {
			stealAmount = targetTotal
		}
		if stealAmount > total {
			stealAmount = total
		}

		success := rand.Intn(100) < 50

		if success {
			addZrazy(userID, userName, stealAmount)
			addZrazy(targetUserID, targetName, -stealAmount)
			incrementStealSuccess(userID)
			updateStealCooldown(userID, now)
			message := fmt.Sprintf(
				"🦝 БЕБЕБЕ, *%s* тихонечко тырнул %d %s у *%s*!\n📊 Теперь у *%s* на счету %d %s,\nа у *%s* на балике %d %s",
				userName, stealAmount, formatZrazyCount(stealAmount), targetName,
				userName, getTotal(userID), formatZrazyCount(getTotal(userID)),
				targetName, getTotal(targetUserID), formatZrazyCount(getTotal(targetUserID)),
			)
			return sendToTopic(b, c, message)
		} else {
			addZrazy(userID, userName, -stealAmount)
			addZrazy(targetUserID, targetName, stealAmount)
			incrementStealFail(userID)
			updateStealCooldown(userID, now)

			message := fmt.Sprintf(
				"💥 ХИХИХИ, *%s* попытался украсть %d %s у *%s*, но ставка не зашла (ЛОШАРА) и он лузнул свои %d %s!\n📊 Терь у *%s* на балике %d %s,\nа у *%s* на счету %d %s",
				userName, stealAmount, formatZrazyCount(stealAmount), targetName,
				stealAmount, formatZrazyCount(stealAmount),
				userName, getTotal(userID), formatZrazyCount(getTotal(userID)),
				targetName, getTotal(targetUserID), formatZrazyCount(getTotal(targetUserID)),
			)
			return sendToTopic(b, c, message)
		}
	})

	b.Handle("/give", func(c tele.Context) error {
		userID := c.Sender().ID
		userName := c.Sender().FirstName
		args := c.Args()

		if len(args) < 2 {
			return sendToTopic(b, c, "🎁 *С ДНЕМ ЗРАЗЫ*\n\n🎭 *Меценатство - как юзать:*\n`/give @username [количество]`\n\nПример: `/give @derden993 5`")
		}

		amount, err := parseAmount(args[len(args)-1])
		if err != nil {
			return sendToTopic(b, c, "❌ *Ошибка отправки:* Надо целое положительное число зраз для подарка")
		}

		targetMention := strings.TrimPrefix(args[0], "@")

		targetUserID, targetName, err := getUserByUsername(targetMention)
		if err != nil {
			return sendToTopic(b, c, fmt.Sprintf("❌ *Ошибка:* %s", err.Error()))
		}

		if targetUserID == userID {
			return sendToTopic(b, c, "_Лучше подари зразы моему хозяину @theG82_")
		}

		total := getTotal(userID)
		if total < amount {
			return sendToTopic(b, c, fmt.Sprintf("⚠️ _У тя только %d %s, для подарка не хватает. Совсем туго с матешей?_", total, formatZrazyCount(total)))
		}

		addZrazy(userID, userName, -amount)
		addZrazy(targetUserID, targetName, amount)
		incrementGive(userID, amount)

		message := fmt.Sprintf(
			"🎁 юхууу, *%s* подарил %d %s пользователю *%s*!\n📊 Теперь у *%s* на балике - %d %s, а у *%s* - %d %s",
			userName, amount, formatZrazyCount(amount), targetName,
			userName, getTotal(userID), formatZrazyCount(getTotal(userID)),
			targetName, getTotal(targetUserID), formatZrazyCount(getTotal(targetUserID)),
		)
		return sendToTopic(b, c, message)
	})

	log.Println("Бот запущен! Напиши /zraza в Telegram")
	b.Start()
}

type userStats struct {
	name  string
	total int
}

type userCountStats struct {
	name  string
	count int
}

type userGiveStats struct {
	name  string
	total int
}

func getUserByUsername(username string) (int64, string, error) {
	var userID int64
	var userName string

	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, "", fmt.Errorf("ошибка базы данных")
	}
	defer db.Close()

	err = db.QueryRow("SELECT user_id, user_name FROM users WHERE user_name LIKE ?", "%"+username+"%").Scan(&userID, &userName)
	if err != nil {
		return 0, "", fmt.Errorf("пользователь @%s не найден в базе", username)
	}

	return userID, userName, nil
}

func parseAmount(s string) (int, error) {
	var amount int
	_, err := fmt.Sscanf(s, "%d", &amount)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("неверное количество")
	}
	return amount, nil
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
			max_total INTEGER DEFAULT 0,
			last_used INTEGER DEFAULT 0,
			shit_total INTEGER DEFAULT 0,
			steal_cooldown INTEGER DEFAULT 0,
			lucky_count INTEGER DEFAULT 0,
			steal_success INTEGER DEFAULT 0,
			steal_fail INTEGER DEFAULT 0,
			give_total INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatal(err)
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
			user_name = EXCLUDED.user_name,
			max_total = CASE 
				WHEN total + ? > max_total THEN total + ? 
				ELSE max_total 
			END
	`, userID, userName, amount, amount, amount, amount)
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

func incrementLuckyCount(userID int64) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, lucky_count) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET lucky_count = lucky_count + 1
	`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func incrementStealSuccess(userID int64) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, steal_success) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET steal_success = steal_success + 1
	`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func incrementStealFail(userID int64) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, steal_fail) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET steal_fail = steal_fail + 1
	`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func incrementGive(userID int64, amount int) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, give_total) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET give_total = give_total + ?
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

func getStealCooldown(userID int64) int64 {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return 0
	}
	defer db.Close()

	var stealCooldown int64
	err = db.QueryRow("SELECT steal_cooldown FROM users WHERE user_id = ?", userID).Scan(&stealCooldown)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		log.Println("DB error:", err)
		return 0
	}
	return stealCooldown
}

func updateStealCooldown(userID int64, timestamp int64) {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO users (user_id, steal_cooldown) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET steal_cooldown = EXCLUDED.steal_cooldown
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
		SELECT user_name, shit_total FROM users 
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
		if err := rows.Scan(&u.name, &u.total); err != nil {
			log.Println("DB error:", err)
			continue
		}
		users = append(users, u)
	}

	return users
}

func getLuckyLeaderboard(limit int) []userCountStats {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT user_name, lucky_count FROM users 
		WHERE lucky_count > 0 
		ORDER BY lucky_count DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()

	var users []userCountStats
	for rows.Next() {
		var u userCountStats
		if err := rows.Scan(&u.name, &u.count); err != nil {
			log.Println("DB error:", err)
			continue
		}
		users = append(users, u)
	}

	return users
}

func getStealLeaderboard(limit int) []userCountStats {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT user_name, steal_success FROM users 
		WHERE steal_success > 0 
		ORDER BY steal_success DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()

	var users []userCountStats
	for rows.Next() {
		var u userCountStats
		if err := rows.Scan(&u.name, &u.count); err != nil {
			log.Println("DB error:", err)
			continue
		}
		users = append(users, u)
	}

	return users
}

func getGiveLeaderboard(limit int) []userGiveStats {
	dbPath := getDBPath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT user_name, give_total FROM users 
		WHERE give_total > 0 
		ORDER BY give_total DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()

	var users []userGiveStats
	for rows.Next() {
		var u userGiveStats
		if err := rows.Scan(&u.name, &u.total); err != nil {
			log.Println("DB error:", err)
			continue
		}
		users = append(users, u)
	}

	return users
}
