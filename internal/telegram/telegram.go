package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewBot(botToken string) *tgbotapi.BotAPI {

	if botToken == "" {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		fmt.Println("Ошибка при создании бота: ", err)
		return nil
	}

	return bot
}
