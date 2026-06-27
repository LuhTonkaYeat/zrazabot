package main

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

var garnishes = []string{
	"🍝 макароны с подливкой",
	"🍚 рис с подливкой",
	"🥔 пюре с подливкой",
	"🌾 гречку с подливкой",
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

func parseAmount(s string) (int, error) {
	var amount int
	_, err := fmt.Sscanf(s, "%d", &amount)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("неверное количество")
	}
	return amount, nil
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
