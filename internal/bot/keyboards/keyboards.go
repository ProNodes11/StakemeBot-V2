package keyboard

import (
	"StakemeBotV2/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func MainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
    buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Chains", "chains"),
		tgbotapi.NewInlineKeyboardButtonData("Uptime", "uptime"),
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, btn := range buttons {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func ReturnMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Назад↩️", "return"),
	}

	rows := [][]tgbotapi.InlineKeyboardButton{
		buttons,
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func ProjectsMenuKeyboard(projects config.Config) tgbotapi.InlineKeyboardMarkup {
	buttons := []tgbotapi.InlineKeyboardButton{}
	for _, chain := range projects.Chains {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(chain.Name, "uptime_"+chain.Name))
	}
	buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Назад↩️", "return"))
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, btn := range buttons {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuyMenuKeyboard(projectName string) tgbotapi.InlineKeyboardMarkup {
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Добавить в корзину 🛒", "buy_"+projectName),
		tgbotapi.NewInlineKeyboardButtonData("Назад↩️", "return"),
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, btn := range buttons {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
