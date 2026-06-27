package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func getDBPath() string {
	dataDir := "/app/data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Println("Warning: cannot create data dir:", err)
	}
	return filepath.Join(dataDir, "zrazy.db")
}

func initDB() {
	dbPath := getDBPath()
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=WAL;")

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            user_id INTEGER PRIMARY KEY,
            user_name TEXT DEFAULT '',
            total INTEGER DEFAULT 0,
            max_total INTEGER DEFAULT 0,
            last_used INTEGER DEFAULT 0,
            shit_total INTEGER DEFAULT 0,
            steal_cooldown INTEGER DEFAULT 0,
            slot_cooldown INTEGER DEFAULT 0,
            all_cooldown INTEGER DEFAULT 0,
            lucky_count INTEGER DEFAULT 0,
            steal_success INTEGER DEFAULT 0,
            steal_fail INTEGER DEFAULT 0,
            give_total INTEGER DEFAULT 0,
            has_base_perk INTEGER DEFAULT 0,
            has_plus_perk INTEGER DEFAULT 0,
            has_prime_perk INTEGER DEFAULT 0,
            has_balabol INTEGER DEFAULT 0,
            has_gigabalabol INTEGER DEFAULT 0,
            has_jew INTEGER DEFAULT 0,
            has_bmw INTEGER DEFAULT 0,
            has_antigovno INTEGER DEFAULT 0,
            message_count INTEGER DEFAULT 0,
            last_daily_reward TEXT DEFAULT ''
        )
    `)
	if err != nil {
		log.Fatal(err)
	}
}

func hasPerk(userID int64, column string) bool {
	var val int
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id = ?", column)
	err := db.QueryRow(query, userID).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Println("DB error:", err)
		return false
	}
	return val == 1
}

func buyPerk(userID int64, column string) {
	query := fmt.Sprintf("UPDATE users SET %s = 1 WHERE user_id = ?", column)
	_, err := db.Exec(query, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func addZrazy(userID int64, userName string, amount int) {
	_, err := db.Exec(`
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
	_, err := db.Exec(`UPDATE users SET total = 0 WHERE user_id = ?`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func addShit(userID int64, userName string, amount int) {
	_, err := db.Exec(`
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
	var total int
	err := db.QueryRow("SELECT total FROM users WHERE user_id = ?", userID).Scan(&total)
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
	var lastUsed int64
	err := db.QueryRow("SELECT last_used FROM users WHERE user_id = ?", userID).Scan(&lastUsed)
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
	_, err := db.Exec(`
		INSERT INTO users (user_id, last_used) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET last_used = EXCLUDED.last_used
	`, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getLastDailyReward(userID int64) string {
	var date string
	db.QueryRow("SELECT last_daily_reward FROM users WHERE user_id = ?", userID).Scan(&date)
	return date
}

func updateDailyReward(userID int64, date string) {
	db.Exec("UPDATE users SET last_daily_reward = ? WHERE user_id = ?", date, userID)
}

func incrementLuckyCount(userID int64) {
	_, err := db.Exec(`
		INSERT INTO users (user_id, lucky_count) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET lucky_count = lucky_count + 1
	`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getStealCooldown(userID int64) int64 {
	var stealCooldown int64
	err := db.QueryRow("SELECT steal_cooldown FROM users WHERE user_id = ?", userID).Scan(&stealCooldown)
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
	_, err := db.Exec(`
		INSERT INTO users (user_id, steal_cooldown) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET steal_cooldown = EXCLUDED.steal_cooldown
	`, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func incrementStealSuccess(userID int64) {
	_, err := db.Exec(`
		INSERT INTO users (user_id, steal_success) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET steal_success = steal_success + 1
	`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func incrementStealFail(userID int64) {
	_, err := db.Exec(`
		INSERT INTO users (user_id, steal_fail) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET steal_fail = steal_fail + 1
	`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func incrementGive(userID int64, amount int) {
	_, err := db.Exec(`
		INSERT INTO users (user_id, give_total) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET give_total = give_total + ?
	`, userID, amount, amount)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getSlotCooldown(userID int64) int64 {
	var slotCooldown int64
	err := db.QueryRow("SELECT slot_cooldown FROM users WHERE user_id = ?", userID).Scan(&slotCooldown)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		log.Println("DB error:", err)
		return 0
	}
	return slotCooldown
}

func updateSlotCooldown(userID int64, timestamp int64) {
	_, err := db.Exec(`
        INSERT INTO users (user_id, slot_cooldown) VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET slot_cooldown = EXCLUDED.slot_cooldown
    `, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getAllCooldown(userID int64) int64 {
	var allCooldown int64
	err := db.QueryRow("SELECT all_cooldown FROM users WHERE user_id = ?", userID).Scan(&allCooldown)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		log.Println("DB error:", err)
		return 0
	}
	return allCooldown
}

func updateAllCooldown(userID int64, timestamp int64) {
	_, err := db.Exec(`
        INSERT INTO users (user_id, all_cooldown) VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET all_cooldown = EXCLUDED.all_cooldown
    `, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func resetSlotCooldown(userID int64) {
	_, err := db.Exec(`UPDATE users SET slot_cooldown = 0 WHERE user_id = ?`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func resetAllCooldown(userID int64) {
	_, err := db.Exec(`UPDATE users SET all_cooldown = 0 WHERE user_id = ?`, userID)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getLeaderboard(limit int) []userStats {
	rows, err := db.Query(`SELECT user_name, total FROM users WHERE total > 0 ORDER BY total DESC LIMIT ?`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()
	var users []userStats
	for rows.Next() {
		var u userStats
		if err := rows.Scan(&u.name, &u.total); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

func getMaxTotalLeaderboard(limit int) []userMaxTotalStats {
	rows, err := db.Query(`SELECT user_name, max_total FROM users WHERE max_total > 0 ORDER BY max_total DESC LIMIT ?`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()
	var users []userMaxTotalStats
	for rows.Next() {
		var u userMaxTotalStats
		if err := rows.Scan(&u.name, &u.maxTotal); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

func getShitLeaderboard(limit int) []userStats {
	rows, err := db.Query(`SELECT user_name, shit_total FROM users WHERE shit_total > 0 ORDER BY shit_total DESC LIMIT ?`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()
	var users []userStats
	for rows.Next() {
		var u userStats
		if err := rows.Scan(&u.name, &u.total); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

func getLuckyLeaderboard(limit int) []userCountStats {
	rows, err := db.Query(`SELECT user_name, lucky_count FROM users WHERE lucky_count > 0 ORDER BY lucky_count DESC LIMIT ?`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()
	var users []userCountStats
	for rows.Next() {
		var u userCountStats
		if err := rows.Scan(&u.name, &u.count); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

func getStealLeaderboard(limit int) []userCountStats {
	rows, err := db.Query(`SELECT user_name, steal_success FROM users WHERE steal_success > 0 ORDER BY steal_success DESC LIMIT ?`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()
	var users []userCountStats
	for rows.Next() {
		var u userCountStats
		if err := rows.Scan(&u.name, &u.count); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

func getGiveLeaderboard(limit int) []userGiveStats {
	rows, err := db.Query(`SELECT user_name, give_total FROM users WHERE give_total > 0 ORDER BY give_total DESC LIMIT ?`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()
	var users []userGiveStats
	for rows.Next() {
		var u userGiveStats
		if err := rows.Scan(&u.name, &u.total); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users
}

func distributeJewTax(loserID int64, lostAmount int) {
	if lostAmount <= 0 {
		return
	}
	rows, err := db.Query("SELECT user_id, user_name FROM users WHERE has_jew = 1 AND user_id != ?", loserID)
	if err != nil {
		return
	}
	defer rows.Close()
	tax := int(float64(lostAmount) * 0.03)
	if tax < 1 {
		return
	}
	for rows.Next() {
		var jewID int64
		var jewName string
		if err := rows.Scan(&jewID, &jewName); err == nil {
			addZrazy(jewID, jewName, tax)
		}
	}
}
