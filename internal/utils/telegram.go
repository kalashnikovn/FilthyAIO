package utils

import (
	"filthy/internal/constants"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

func SendTelegramMessage(args ...interface{}) {
	bot := constants.TELEGRAM_BOT
	if bot == nil {
		//fmt.Println("бот не был инициализирован, скорее всего не задан botToken")
		return
		//return errors.New("бот не был инициализирован, скорее всего не задан botToken")
	}

	chatID := constants.SETTINGS.Telegram.ChatId
	if chatID == 0 {
		//fmt.Println("chatId=0, введи правильный айди чата")
		return
		//return errors.New("chatId=0, введи правильный айди чата")
	}

	var textBuilder strings.Builder
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			textBuilder.WriteString(v)
		case int:
			textBuilder.WriteString(strconv.Itoa(v))
		case float64:
			textBuilder.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		default:
			textBuilder.WriteString(fmt.Sprintf("%v", v))
		}
		//textBuilder.WriteString(" ")
	}

	message := tgbotapi.NewMessage(chatID, textBuilder.String())

	// Отправляем сообщение с помощью бота
	_, err := bot.Send(message)
	if err != nil {
		//fmt.Println("Ошибка при отправке сообщения: ", err)
		return
		//return errors.New("Ошибка при отправке сообщения: " + err.Error())
	}

	//return nil
}
