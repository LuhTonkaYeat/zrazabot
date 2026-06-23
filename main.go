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

var db *sql.DB

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

func formatZrazyNominative(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "зраза"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "зразы"
	}
	return "зраз"
}

func formatZrazyAccusative(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "зразу"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "зразы"
	}
	return "зраз"
}

func formatZrazyGenitive(count int) string {
	if count%10 == 1 && count%100 != 11 {
		return "зразы"
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return "зраз"
	}
	return "зраз"
}

func formatGenericCount(count int, one string, few string, many string) string {
	if count%10 == 1 && count%100 != 11 {
		return one
	} else if (count%10 >= 2 && count%10 <= 4) && (count%100 < 10 || count%100 >= 20) {
		return few
	}
	return many
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

func animateSlots(b *tele.Bot, c tele.Context, initialText string, slotSymbols []string) (*tele.Message, []string, error) {
	msg := c.Message()
	opt := &tele.SendOptions{
		ParseMode: tele.ModeMarkdown,
		ThreadID:  msg.ThreadID,
	}

	startMsg, err := b.Send(msg.Chat, initialText, opt)
	if err != nil {
		return nil, nil, err
	}

	var lastFrameResults []string

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)

		animResults := []string{
			slotSymbols[rand.Intn(len(slotSymbols))],
			slotSymbols[rand.Intn(len(slotSymbols))],
			slotSymbols[rand.Intn(len(slotSymbols))],
		}

		if i == 2 {
			lastFrameResults = animResults
		}

		displayText := fmt.Sprintf("%s | %s | %s", animResults[0], animResults[1], animResults[2])
		_, _ = b.Edit(startMsg, displayText, opt)
	}

	time.Sleep(900 * time.Millisecond)

	return startMsg, lastFrameResults, nil
}

func calculateWin(results []string, betAmount int) (int, string) {
	r1, r2, r3 := results[0], results[1], results[2]

	if r1 == r2 && r2 == r3 {
		switch r1 {
		case "💎":
			return betAmount * 10, " (x10)"
		case "7️⃣":
			return betAmount * 7, " (x7)"
		case "🍒":
			return betAmount * 3, " (x3)"
		case "🍋", "🍉":
			return betAmount * 2, " (x2)"
		}
	}

	if r1 == r2 || r2 == r3 || r1 == r3 {
		win := int(float64(betAmount) * 1.5)
		return win, " (x1.5)"
	}

	return 0, ""
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

	dbPath := getDBPath()
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

	b.Handle("/zraza", func(c tele.Context) error {
		userID := c.Sender().ID
		userName := c.Sender().FirstName

		today := time.Now().Format("2006-01-02")
		lastReward := getLastDailyReward(userID)

		if lastReward != today {
			rewardMsg := ""
			nowTime := time.Now()

			if hasPerk(userID, "has_prime_perk") && nowTime.Hour() >= 9 {
				addZrazy(userID, userName, 15)
				rewardMsg += "_Доброго времени суток, повелитель! Остальные бичи не такие пиздaтыe, как ты!!!\nЛови перки:_\n\n*+15 🥣 (Prime)*\n\n"
			}

			if hasPerk(userID, "has_plus_perk") && (nowTime.Hour() > 9 || (nowTime.Hour() == 9 && nowTime.Minute() >= 30)) {
				addZrazy(userID, userName, 10)
				rewardMsg += "+10 🥣 (Plus)\n\n"
			}

			if hasPerk(userID, "has_base_perk") && nowTime.Hour() >= 10 {
				addZrazy(userID, userName, 3)
				rewardMsg += "+3 🥣 (Base)\n"
			}

			if rewardMsg != "" {
				updateDailyReward(userID, today)
				go sendToTopic(b, c, fmt.Sprintf("🌅 *Раздача!*\n\n%s", rewardMsg))
			}
		}

		lastUsed := getLastUsed(userID)
		now := time.Now().Unix()
		if now-lastUsed < 1800 && lastUsed != 0 {
			secondsLeft := 1800 - (now - lastUsed)
			timeLeft := formatCooldown(secondsLeft)
			return sendToTopic(b, c, fmt.Sprintf("⏰ _%s, сначала нагуляй аппетyeat!!!_\n_Осталось ждать: %s_\n\n🍽_ /zraza_",
				userName, timeLeft))
		}

		rarity := rand.Intn(100)

		if rarity < 5 {
			addZrazy(userID, userName, 67)
			incrementLuckyCount(userID)
			updateLastUsed(userID, now)
			garnish := garnishes[rand.Intn(len(garnishes))]
			return sendToTopic(b, c, fmt.Sprintf("_✨✨✨ ЧУДО! ЧЗХХХ!!!_\n*%s* _нашел заначку и сожрал 67 eбaныx зраз и %s!!!_\n\n🍽_ /zraza_",
				userName, garnish))
		}

		if rarity < 10 {
			hasAntiGovno := hasPerk(userID, "has_antigovno")

			if hasAntiGovno && rand.Intn(100) >= 2 {
				return sendToTopic(b, c, "_🛡️ Antigovno спас твою жопу! Обнуления не будэ_\n\n🍽_ /zraza_")
			}

			resetZrazy(userID)
			addShit(userID, userName, 1)
			updateLastUsed(userID, now)

			return sendToTopic(b, c, fmt.Sprintf("_💩💩💩 ХЕХЕХЕХЕ, %s навернул тарелку говнеца и обнулил свой счётчик зраз!_\n\n🍽_ /zraza_",
				userName))
		}

		eaten := rand.Intn(20) + 1
		garnish := garnishes[rand.Intn(len(garnishes))]
		addZrazy(userID, userName, eaten)
		updateLastUsed(userID, now)
		total := getTotal(userID)

		return sendToTopic(b, c, fmt.Sprintf("_%s ток что сожрал %d eбaныx %s и %s!!!_\n\n📊 *Балик: %d %s*\n\n🍽 _/zraza_",
			userName, eaten, formatZrazyAccusative(eaten), garnish, total, formatZrazyAccusative(total)))
	})

	b.Handle("/stat", func(c tele.Context) error {
		users := getLeaderboard(5)
		if len(users) == 0 {
			return sendToTopic(b, c, "_Пока никто не ел зразы... Напиши /zraza_")
		}

		message := "🥣 *Актуальный еврейтинг СОШ №1 по поеданию зраз:*\n\n"
		for i, u := range users {
			message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, u.name, u.total, formatZrazyNominative(u.total))
		}

		return sendToTopic(b, c, message)
	})

	b.Handle("/top", func(c tele.Context) error {
		message := ""

		maxTotalLeaders := getMaxTotalLeaderboard(5)
		if len(maxTotalLeaders) > 0 {
			message += "🏆🏆🏆 *АБСОЛЮТНЫЕ ЛИДЕРЫ ПО ЗРАЗОПОЕДАНИЮ:*\n\n"
			for i, l := range maxTotalLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, l.name, l.maxTotal, formatZrazyNominative(l.maxTotal))
			}
			message += "\n"
		}

		shitLeaders := getShitLeaderboard(5)
		if len(shitLeaders) > 0 {
			message += "_💩 Топ говноедов:_\n"
			for i, s := range shitLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, s.name, s.total, formatGenericCount(s.total, "раз", "раза", "раз"))
			}
			message += "\n"
		}

		luckyLeaders := getLuckyLeaderboard(5)
		if len(luckyLeaders) > 0 {
			message += "_✨ Топ лакеров (67 зраз):_\n"
			for i, l := range luckyLeaders {
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, l.name, l.count, formatGenericCount(l.count, "раз", "раза", "раз"))
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
				message += fmt.Sprintf("%d. _%s_ - _%d %s_\n", i+1, g.name, g.total, formatZrazyGenitive(g.total))
			}
			message += "\n"
		}

		if len(message) == 0 {
			message = "_Либо бд пpoeбaлacь, либо еще нет актива..._"
		}

		return sendToTopic(b, c, message)
	})

	b.Handle("/steal", func(c tele.Context) error {
		userID := c.Sender().ID
		userName := c.Sender().FirstName
		msg := c.Message()

		if msg.ReplyTo == nil {
			return sendToTopic(b, c, "🥷🏻 *Ском*\n\n🎭 *Кража зраз - как юзать:*\n_Ответь на соо цели:\n'/steal'_")
		}

		targetUserID := msg.ReplyTo.Sender.ID
		targetName := msg.ReplyTo.Sender.FirstName

		if targetUserID == userID {
			return sendToTopic(b, c, "_Слышь, умник, я щас мамке твоей пожалуюсь, что ты абузить пытаешься, усёк? 😉_")
		}

		stealCooldown := getStealCooldown(userID)
		now := time.Now().Unix()
		if now-stealCooldown < 3600*5 && stealCooldown != 0 {
			secondsLeft := 3600*5 - (now - stealCooldown)
			timeLeft := formatCooldown(secondsLeft)
			return sendToTopic(b, c, fmt.Sprintf("⏰ Не, %s, додепа не будет!\n_Попробуй тырнуть через: %s_\n\n🍽 А лучше сам похавай /zraza",
				userName, timeLeft))
		}

		total := getTotal(userID)
		if total < 5 {
			return sendToTopic(b, c, fmt.Sprintf("⚠️ _Обнаружен нищук! (%s) У тя меньше 5 зраз. Накопи сначала, а потом уже воруй dayum_\n\n🍽 Cам пожри /zraza",
				userName))
		}

		targetTotal := getTotal(targetUserID)
		if targetTotal < 5 {
			return sendToTopic(b, c, fmt.Sprintf("😢 _У %s всего %d %s, с него нех взять..._",
				targetName, targetTotal, formatZrazyGenitive(targetTotal)))
		}

		stealAmount := rand.Intn(5) + 1
		success := rand.Intn(100) < 50

		if success {
			addZrazy(userID, userName, stealAmount)
			addZrazy(targetUserID, targetName, -stealAmount)
			incrementStealSuccess(userID)
			updateStealCooldown(userID, now)
			return sendToTopic(b, c, fmt.Sprintf("🦝 БЕБЕБЕ, *%s* тихонечко тырнул %d %s у *%s*!",
				userName, stealAmount, formatZrazyAccusative(stealAmount), targetName))
		} else {
			addZrazy(userID, userName, -stealAmount)
			addZrazy(targetUserID, targetName, stealAmount)
			incrementStealFail(userID)
			updateStealCooldown(userID, now)
			return sendToTopic(b, c, fmt.Sprintf("💥 ХИХИХИ, *%s* попытался украсть %d %s у *%s*, но ставка не зашла (ЛОШАРА) и он лузнул свои %d %s!",
				userName, stealAmount, formatZrazyAccusative(stealAmount), targetName,
				stealAmount, formatZrazyAccusative(stealAmount)))
		}
	})

	b.Handle("/give", func(c tele.Context) error {
		userID := c.Sender().ID
		userName := c.Sender().FirstName
		args := c.Args()
		msg := c.Message()

		if len(args) < 1 || msg.ReplyTo == nil {
			return sendToTopic(b, c, "🎁 *С днем зразы*\n\n🎭 *Подарки - как юзать:*\n_Ответь на соо получателя:\n'/give [количество]'_")
		}

		amount, err := parseAmount(args[0])
		if err != nil {
			return sendToTopic(b, c, "❌ *Ошибка:* Надо целое положительное число зраз для подарка")
		}

		targetUserID := msg.ReplyTo.Sender.ID
		targetName := msg.ReplyTo.Sender.FirstName

		if targetUserID == userID {
			return sendToTopic(b, c, "_Лучше подари зразы моему хозяину @theG82_")
		}

		total := getTotal(userID)
		if total < amount {
			return sendToTopic(b, c, fmt.Sprintf("⚠️ _У тя только %d %s, для подарка не хватает_", total, formatZrazyGenitive(total)))
		}

		addZrazy(userID, userName, -amount)
		addZrazy(targetUserID, targetName, amount)
		incrementGive(userID, amount)
		targetTotal := getTotal(targetUserID)

		return sendToTopic(b, c, fmt.Sprintf("🎁🎁🎁 ЮХУУУ, *%s* подарил %d %s пользователю *%s*!\n\n📊 Балик: %d %s",
			userName, amount, formatZrazyAccusative(amount), targetName, targetTotal, formatZrazyAccusative(targetTotal)))
	})

	b.Handle("/slot", func(c tele.Context) error {
		userID := c.Sender().ID
		userFirstName := c.Sender().FirstName
		args := c.Args()

		if len(args) < 1 {
			return sendToTopic(b, c, "🎰 *Гемблинг*\n\n💎 *Слоты - как юзать?*\n_Просто напиши:\n'/slot [ставка (≤50)]'_")
		}

		amount, err := parseAmount(args[0])
		if err != nil || amount > 50 {
			return sendToTopic(b, c, "_❌ *Ошибка:* Надо целое положительное число зраз, <= 50 для ставки. Иначе ты слудоманишься хах_")
		}

		total := getTotal(userID)
		if total == 0 {
			return sendToTopic(b, c, fmt.Sprintf("*%s*, ты серьезно?) У тя на балике 0... Можт похаваешь сначала?\n\n🍽_ /zraza_", userFirstName))
		}
		if total < amount {
			diff := amount - total
			return sendToTopic(b, c, fmt.Sprintf("*%s*, Понизь ставку на %d %s, у тя не хватает балика", userFirstName, diff, formatZrazyAccusative(diff)))
		}

		lastSlot := getSlotCooldown(userID)
		now := time.Now().Unix()
		cooldownTime := int64(180)

		if hasBMW := hasPerk(userID, "has_bmw"); hasBMW {
			cooldownTime = int64(float64(cooldownTime) * 0.75)
		}

		if now-lastSlot < cooldownTime && lastSlot != 0 {
			secondsLeft := cooldownTime - (now - lastSlot)
			timeLeft := formatCooldown(secondsLeft)
			return sendToTopic(b, c, fmt.Sprintf("⏰ _%s, погоди %s_", userFirstName, timeLeft))
		}

		slots := []string{"🍒", "🍋", "🍉", "💎", "7️⃣"}

		startMsg, finalResults, err := animateSlots(b, c, "🎰 *Поехали...* 🎰", slots)
		if err != nil {
			return err
		}

		winAmount, multiplierText := calculateWin(finalResults, amount)
		slotDisplay := fmt.Sprintf("%s | %s | %s", finalResults[0], finalResults[1], finalResults[2])

		updateSlotCooldown(userID, now)
		if winAmount > 0 {
			addZrazy(userID, userFirstName, winAmount)
		} else {
			addZrazy(userID, userFirstName, -amount)
			distributeJewTax(userID, amount)
		}

		newTotal := getTotal(userID)

		var finalMessage string
		if winAmount > 0 {
			finalMessage = fmt.Sprintf("%s\n\n*%s* выиграл %d %s!%s\n\n📊 *Балик:* %d %s",
				slotDisplay, userFirstName, winAmount, formatZrazyAccusative(winAmount), multiplierText,
				newTotal, formatZrazyNominative(newTotal))
		} else {
			finalMessage = fmt.Sprintf("%s\n\n*%s* проебал %d %s...\n\n📊 *Балик:* %d %s",
				slotDisplay, userFirstName, amount, formatZrazyAccusative(amount),
				newTotal, formatZrazyNominative(newTotal))
		}

		opt := &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
			ThreadID:  c.Message().ThreadID,
		}
		_, err = b.Edit(startMsg, finalMessage, opt)
		return err
	})

	b.Handle("/all", func(c tele.Context) error {
		userID := c.Sender().ID
		userFirstName := c.Sender().FirstName

		lastAll := getAllCooldown(userID)
		now := time.Now().Unix()
		cooldownAll := int64(10800)

		if hasBMW := hasPerk(userID, "has_bmw"); hasBMW {
			cooldownAll = int64(float64(cooldownAll) * 0.75)
		}

		if now-lastAll < cooldownAll && lastAll != 0 {
			secondsLeft := cooldownAll - (now - lastAll)
			timeLeft := formatCooldown(secondsLeft)
			return sendToTopic(b, c, fmt.Sprintf("⏰ _%s, all in можно делать только раз в 3 часа_\n_Осталось ждать: %s_",
				userFirstName, timeLeft))
		}

		total := getTotal(userID)
		if total < 5 {
			return sendToTopic(b, c, fmt.Sprintf("⚠️ _%s, для аллына нужно хотя бы 5 зраз... Накопи сначала_", userFirstName))
		}

		slots := []string{"🍒", "🍋", "🍉", "💎", "7️⃣"}

		startMsg, finalResults, err := animateSlots(b, c, "🔥 *ALL IN! ПОШЛО-ПОЕХАЛО...* 🔥", slots)
		if err != nil {
			return err
		}

		winAmount, multiplierText := calculateWin(finalResults, total)
		slotDisplay := fmt.Sprintf("%s | %s | %s", finalResults[0], finalResults[1], finalResults[2])

		updateAllCooldown(userID, now)
		if winAmount > 0 {
			addZrazy(userID, userFirstName, winAmount)
		} else {
			resetZrazy(userID)
			distributeJewTax(userID, total)
		}

		newTotal := getTotal(userID)

		var finalMessage string
		if winAmount > 0 {
			finalMessage = fmt.Sprintf("%s\n\n🔥 *ALL IN!* 🔥\n*%s* поставил всё (%d %s) и выиграл %d %s!%s\n\n📊 *Балик:* %d %s",
				slotDisplay, userFirstName, total, formatZrazyAccusative(total),
				winAmount, formatZrazyAccusative(winAmount), multiplierText,
				newTotal, formatZrazyNominative(newTotal))
		} else {
			finalMessage = fmt.Sprintf("%s\n\n💀 *ALL IN* 💀\n*%s* поставил всё (%d %s) и проебал всё!\n\nЛУДИК EБAHЫЙ!!!\n\n📊 *Балик:* 000000 зраз",
				slotDisplay, userFirstName, total, formatZrazyGenitive(total))
		}

		opt := &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
			ThreadID:  c.Message().ThreadID,
		}
		_, err = b.Edit(startMsg, finalMessage, opt)
		return err
	})

	b.Handle("/shop", func(c tele.Context) error {
		userID := c.Sender().ID

		hasBase := hasPerk(userID, "has_base_perk")
		hasPlus := hasPerk(userID, "has_plus_perk")
		hasPrime := hasPerk(userID, "has_prime_perk")
		hasBalabol := hasPerk(userID, "has_balabol")
		hasGiga := hasPerk(userID, "has_gigabalabol")
		hasJew := hasPerk(userID, "has_jew")
		hasBMW := hasPerk(userID, "has_bmw")
		hasAntiGovno := hasPerk(userID, "has_antigovno")

		total := getTotal(userID)

		message := "🛒 *МАГАЗИН ПЕРКОВ*\n\n"

		message += "🥄 *Base Perk* — 150 зраз\n— Ежедневно +3 зразы после 10:00\n"
		if hasBase {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "🥄 *Plus Perk* — 500 зраз\n— Ежедневно +10 зраз после 9:30\n"
		if hasPlus {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "👑 *Prime Perk* — 1000 зраз\n— Ежедневно +15 зраз после 9:00 + статус в чате\n"
		if hasPrime {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "💬 *Balabol* — 800 зраз\n— +10 зраз за каждые 50 сообщений\n"
		if hasBalabol {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "💬 *Gigabalabol* — 2500 зраз\n— +30 зраз за каждые 50 сообщений\n"
		if hasGiga {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "💰 *Jew* — 5000 зраз\n— 3%% от проигрышей других в казино\n"
		if hasJew {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "🚗 *BMW M4 Owner* — 480 зраз\n— -25%% ко всем кулдаунам\n"
		if hasBMW {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += "🛡️ *Antigovno* — 1500 зраз\n— Шанс обнуления ↓ до 2%%\n"
		if hasAntiGovno {
			message += "   ✅ *КУПЛЕНО*\n"
		}
		message += "\n"

		message += fmt.Sprintf("\n💰 *Твой балик:* %d %s", total, formatZrazyNominative(total))

		var buttons []tele.InlineButton
		if !hasBase {
			buttons = append(buttons, tele.InlineButton{Text: "🥄 Base perk", Data: "buy_base"})
		}
		if !hasPlus {
			buttons = append(buttons, tele.InlineButton{Text: "🥄 Plus perk", Data: "buy_plus"})
		}
		if !hasPrime {
			buttons = append(buttons, tele.InlineButton{Text: "👑 Prime perk", Data: "buy_prime"})
		}
		if !hasBalabol {
			buttons = append(buttons, tele.InlineButton{Text: "💬 Balabol", Data: "buy_balabol"})
		}
		if !hasGiga {
			buttons = append(buttons, tele.InlineButton{Text: "💬 Gigabalabol", Data: "buy_giga"})
		}
		if !hasJew {
			buttons = append(buttons, tele.InlineButton{Text: "💰 Jew", Data: "buy_jew"})
		}
		if !hasBMW {
			buttons = append(buttons, tele.InlineButton{Text: "🚗 BMW M4", Data: "buy_bmw"})
		}
		if !hasAntiGovno {
			buttons = append(buttons, tele.InlineButton{Text: "🛡️ Antigovno", Data: "buy_antigovno"})
		}

		var rows [][]tele.InlineButton
		for i := 0; i < len(buttons); i += 2 {
			end := i + 2
			if end > len(buttons) {
				end = len(buttons)
			}
			rows = append(rows, buttons[i:end])
		}

		if len(buttons) == 0 {
			message += "\n\n✨ *Все перки уже куплены! Eбaный богач...*"
			return sendToTopic(b, c, message)
		}

		msg := c.Message()
		chat := msg.Chat
		topicID := msg.ThreadID

		_, err := b.Send(chat, message, &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
			ThreadID:  topicID,
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: rows,
			},
		})
		return err
	})

	b.Handle(tele.OnCallback, func(c tele.Context) error {
		data := c.Data()
		userID := c.Sender().ID

		if !strings.HasPrefix(data, "buy_") {
			return nil
		}

		var cost int
		var perkName string
		var perkColumn string

		switch data {
		case "buy_base":
			cost = 150
			perkName = "Base Perk"
			perkColumn = "has_base_perk"
		case "buy_plus":
			cost = 500
			perkName = "Plus Perk"
			perkColumn = "has_plus_perk"
		case "buy_prime":
			cost = 1000
			perkName = "Prime Perk"
			perkColumn = "has_prime_perk"
		case "buy_balabol":
			cost = 800
			perkName = "Balabol"
			perkColumn = "has_balabol"
		case "buy_giga":
			cost = 2500
			perkName = "Gigabalabol"
			perkColumn = "has_gigabalabol"
		case "buy_jew":
			cost = 5000
			perkName = "Jew"
			perkColumn = "has_jew"
		case "buy_bmw":
			cost = 480
			perkName = "BMW M4 Owner"
			perkColumn = "has_bmw"
		case "buy_antigovno":
			cost = 1500
			perkName = "Antigovno"
			perkColumn = "has_antigovno"
		default:
			return nil
		}

		if hasPerk(userID, perkColumn) {
			c.Respond(&tele.CallbackResponse{Text: "Уже куплено! ✅"})
			return nil
		}

		total := getTotal(userID)
		if total < cost {
			c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("Не хватает! Нужно: %d, есть: %d", cost, total)})
			return nil
		}

		addZrazy(userID, c.Sender().FirstName, -cost)
		buyPerk(userID, perkColumn)
		c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf("🎉 Куплено! %s", perkName)})

		return c.Edit(c.Message().Text, &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{},
			},
		})
	})

	b.Handle("/kisel", func(c tele.Context) error {
		userName := c.Sender().FirstName
		message := fmt.Sprintf("*%s* _сказал, что КИСЕЛЬ ДАУН_", userName)
		return sendToTopic(b, c, message)
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		userID := c.Sender().ID

		if strings.HasPrefix(c.Text(), "/") {
			return nil
		}

		db.Exec("INSERT INTO users (user_id, message_count) VALUES (?, 1) ON CONFLICT(user_id) DO UPDATE SET message_count = message_count + 1", userID)

		var msgCount int
		err = db.QueryRow("SELECT message_count FROM users WHERE user_id = ?", userID).Scan(&msgCount)
		if err != nil {
			return nil
		}

		if msgCount > 0 && msgCount%50 == 0 {
			reward := 0
			perkName := ""
			if hasPerk(userID, "has_gigabalabol") {
				reward = 30
				perkName = "Gigabalabol"
			} else if hasPerk(userID, "has_balabol") {
				reward = 10
				perkName = "Balabol"
			}

			if reward > 0 {
				addZrazy(userID, c.Sender().FirstName, reward)
				go sendToTopic(b, c, fmt.Sprintf("💬 *%s* активирован! +%d зраз за %d сообщений.", perkName, reward, msgCount))
			}
		}
		return nil
	})

	log.Println("Бот запущен! Напиши /zraza в Telegram")
	b.Start()
}

type userStats struct {
	name  string
	total int
}

type userMaxTotalStats struct {
	name     string
	maxTotal int
}

type userCountStats struct {
	name  string
	count int
}

type userGiveStats struct {
	name  string
	total int
}

func parseAmount(s string) (int, error) {
	var amount int
	_, err := fmt.Sscanf(s, "%d", &amount)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("неверное количество")
	}
	return amount, nil
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
	_, err := db.Exec(`
		UPDATE users SET total = 0 WHERE user_id = ?
	`, userID)
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

func incrementLuckyCount(userID int64) {
	_, err := db.Exec(`
		INSERT INTO users (user_id, lucky_count) VALUES (?, 1)
		ON CONFLICT(user_id) DO UPDATE SET lucky_count = lucky_count + 1
	`, userID)
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

func getLastDailyReward(userID int64) string {
	var date string
	db.QueryRow("SELECT last_daily_reward FROM users WHERE user_id = ?", userID).Scan(&date)
	return date
}

func updateDailyReward(userID int64, date string) {
	db.Exec("UPDATE users SET last_daily_reward = ? WHERE user_id = ?", date, userID)
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

func updateStealCooldown(userID int64, timestamp int64) {
	_, err := db.Exec(`
		INSERT INTO users (user_id, steal_cooldown) VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET steal_cooldown = EXCLUDED.steal_cooldown
	`, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
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

func updateAllCooldown(userID int64, timestamp int64) {
	_, err := db.Exec(`
        INSERT INTO users (user_id, all_cooldown) VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET all_cooldown = EXCLUDED.all_cooldown
    `, userID, timestamp)
	if err != nil {
		log.Println("DB error:", err)
	}
}

func getLeaderboard(limit int) []userStats {
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

func getMaxTotalLeaderboard(limit int) []userMaxTotalStats {
	rows, err := db.Query(`
		SELECT user_name, max_total FROM users 
		WHERE max_total > 0 
		ORDER BY max_total DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		log.Println("DB error:", err)
		return nil
	}
	defer rows.Close()

	var users []userMaxTotalStats
	for rows.Next() {
		var u userMaxTotalStats
		if err := rows.Scan(&u.name, &u.maxTotal); err != nil {
			log.Println("DB error:", err)
			continue
		}
		users = append(users, u)
	}

	return users
}

func getShitLeaderboard(limit int) []userStats {
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

func getLuckyLeaderboard(limit int) []userCountStats {
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
