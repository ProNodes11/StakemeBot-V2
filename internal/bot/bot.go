package bot

import (
	"StakemeBotV2/internal/bot/handlers"
	"StakemeBotV2/internal/config"
	"StakemeBotV2/internal/db"
	"database/sql"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

const (
	StepMainMenu = iota // Главная менюшка
	StepSubMenu // Проекты, FAQ
	StepSubMenu2 // Цена за проект + инфа
	StepSubMenu3 // Покупка
)

func RunBot(cnfg config.Config, dbcon *sql.DB, logger *logrus.Logger, botChan chan *tgbotapi.BotAPI) {
	// Запускаем бота
	bot, err := tgbotapi.NewBotAPI(cnfg.Bot.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	// Передаем бота в канал нужный
	botChan <- bot
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	
	// Загружаем хендлеры
	textHandlers     := handlers.LoadTextHandlers()
	callbackHandlers := handlers.LoadCallbackHandlers()

	// Обрабатываем полученые сообщения
	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}
		// Логика текстовых сообщений
		if update.Message != nil {
			// db.UpsertUser(dbcon, users.NewUser(update.Message.Chat.ID))
			// Обработчик текстовых хендлеров
			user, err := db.GetUserByChatID(dbcon, update.Message.Chat.ID, logger)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"module": "bot",
				}).Errorf("Failed to get user from db %s", err)
			}
			handler, exists := textHandlers[update.Message.Text]
			if exists {
				handler(bot, update, user)
			} else {
				handlers.DefaultMessageHandler(bot, dbcon, update, user, logger)
			}
			db.UpdateUser(dbcon, user, logger)
		// Логика кнопок
		} else if update.CallbackQuery != nil {
			// db.UpsertUser(dbcon, users.NewUser(int64(update.CallbackQuery.From.ID)))
			user, err := db.GetUserByChatID(dbcon, int64(update.CallbackQuery.From.ID), logger)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"module": "bot",
				}).Errorf("Failed to get user from db %s", err)
			}
			// Обработчик кнопок
			callbackData := update.CallbackQuery.Data

			parts := strings.SplitN(callbackData, "_", 2)
			command := parts[0]
			data := ""
			if len(parts) > 1 {
				data = parts[1]
			}
			log.Println(command, data)
			handler, exists := callbackHandlers[command]
			if exists {
				handler(bot, update, data, user, cnfg, dbcon)
			}
			db.UpdateUser(dbcon, user, logger)
		}
		
	}
}

