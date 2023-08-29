package handlers

import (
	keyboard "StakemeBotV2/internal/bot/keyboards"
	"StakemeBotV2/internal/bot/users"
	"StakemeBotV2/internal/config"
	"StakemeBotV2/internal/db"
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

const (
	StepMainMenu = iota // Главная менюшка
	StepSubMenu // Проекты, FAQ
	StepProjectInfo // Цена за проект + инфа
	StepPay // Покупка
)


var main = `👋 Привет, я бот-проводник по софту для мультиаккинга дропов

✅ Софт прогоняет кошельки по топовым актуальным ретродропам (список дальше) в сжатые сроки, на низком газе, без бритья 🚀`
type Handlers struct {
	Callbacks map[string]CallbackHandler
	Text      map[string]TextHandler
}

type CallbackHandler func(bot *tgbotapi.BotAPI, update tgbotapi.Update, callbackParametr string, userData *users.User, projects config.Config, dbcon *sql.DB)
type TextHandler     func(bot *tgbotapi.BotAPI, update tgbotapi.Update, userData *users.User)

type CallbackHandlers map[string]CallbackHandler
type TextHandlers     map[string]TextHandler

func LoadCallbackHandlers() CallbackHandlers {
	return CallbackHandlers{
		// Список проектов
		"chains": func(bot *tgbotapi.BotAPI, update tgbotapi.Update, callbackParametr string, userData  *users.User, projects config.Config, dbcon *sql.DB) {
			dmsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, userData.State.MessageID)
			bot.DeleteMessage(dmsg)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Список актуальных проектов")
			msg.ReplyMarkup = keyboard.ProjectsMenuKeyboard(projects)
			smsg, _ := bot.Send(msg)
			userData.State.MessageID = smsg.MessageID
		},
		// Добавить проект в корзину
		"uptime": func(bot *tgbotapi.BotAPI, update tgbotapi.Update, callbackParametr string, userData  *users.User, projects config.Config, dbcon *sql.DB) {
			dmsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, userData.State.MessageID)
			bot.DeleteMessage(dmsg)
			// project := projects[callbackParametr]
			// userData.Items = append(userData.Items, project)
			log.Println(userData)
			
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Отправьте адрес валидатора")
			// msg.ReplyMarkup = keyboard.OrderMenuKeyboard()
			smsg, _ := bot.Send(msg)
			userData.State.MessageID = smsg.MessageID
			userData.State.WaitForAddress = true
			userData.State.CurrentChain = callbackParametr
			// db.UpdateUser(dbcon, userData, )
			log.Println(userData)
		},
		// Принять хеш транзакции
		"sendhash": func(bot *tgbotapi.BotAPI, update tgbotapi.Update, callbackParametr string, userData *users.User, projects config.Config, dbcon *sql.DB) {
			// dmsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, userData.MessageID)
			// bot.DeleteMessage(dmsg)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Отправьте хеш транзакции или ссылку на нее")
			smsg, _ := bot.Send(msg)
			userData.State.MessageID = smsg.MessageID
			// userData.Step = StepSubMenu
			// userData.WaitForHash = true
		},
	}
}

func LoadTextHandlers() TextHandlers {
	return TextHandlers{
		"/start": func(bot *tgbotapi.BotAPI, update tgbotapi.Update, userData *users.User) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, main)
			msg.ReplyMarkup = keyboard.MainMenuKeyboard()
			smsg, _ := bot.Send(msg)
			userData.State.MessageID = smsg.MessageID
		},	
		"default_message": func(bot *tgbotapi.BotAPI, update tgbotapi.Update, userData *users.User) {
		
		},
	}
}

// func TextMessageHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, userStates map[int64]*users.User) {
// 	if update.Message == nil || update.Message.Text == "" {
// 		return
// 	}

// 	userID := update.Message.From.ID
// 	text := update.Message.Text

// 	if userState, exists := userStates[int64(userID)]; exists {
// 		userState.SendedHash = text
// 	}
// }

func DefaultMessageHandler(bot *tgbotapi.BotAPI, dbcon *sql.DB, update tgbotapi.Update, userData *users.User, logger *logrus.Logger) {
	log.Println(userData.State.WaitForAddress)
	if userData.State.WaitForAddress {
		// dmsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, userData.MessageID)
		// bot.DeleteMessage(dmsg)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Получен адрес")
		// msg.ReplyMarkup = keyboard.SendNotification()
		// bot.Send(msg)
		userData.AddUptimeAlert(userData.State.CurrentChain, update.Message.Text)
		log.Println(userData)
		err := db.UpdateUser(dbcon, userData, logger)
		if err != nil {
			log.Println(err)
		}
		// userData.Step = StepSubMenu
		// userData.WaitForHash = false
		// msg = tgbotapi.NewMessage(384103097, fmt.Sprintf("Юзер @%s купил что-то", update.Message.From.UserName))
		bot.Send(msg)
	}
}