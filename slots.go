package main

import (
	"fmt"
	"math/rand"
	"time"

	tele "gopkg.in/telebot.v3"
)

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
