package app

import (
	"StakemeBotV2/internal/bot"
	"StakemeBotV2/internal/bot/monitor"
	"StakemeBotV2/internal/config"
	"StakemeBotV2/internal/db"
	"StakemeBotV2/internal/logger"
	"StakemeBotV2/internal/services/ws"
	"database/sql"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config  config.Config
	Logger  *logrus.Logger
	DB      *sql.DB
	WG      *sync.WaitGroup
	Bot     *tgbotapi.BotAPI
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init() error {
	// Инициализация конфигурации
	config, err := config.LoadConfig()
	if err != nil {
		return err
	}
	a.Config = config

	// Инициализация логгера
	logger := logger.ConfigLogger(config.App.AppLogLevel)
	a.Logger = logger
	a.Logger.WithField("module", "app").Debug("Succesfully load config")
	a.Logger.WithField("module", "app").Debugf("Succesfully set log level: %s", a.Config.App.AppLogLevel)
	// Инициализация базы данных
	dbcon := db.RunAndConfigDatabase(config, logger)
	a.DB = dbcon

	// Инициализация WaitGroup
	a.WG = &sync.WaitGroup{}

	return nil
}

func (a *App) StartWsCon() {
	a.WG.Add(1)
	go ws.RunWebsocketConnection(a.Config, a.Logger, a.DB)
}

func (a *App) StartTgBot() {
	a.WG.Add(1)
	botChan := make(chan *tgbotapi.BotAPI) // Создаем канал для передачи экземпляра бота
	go bot.RunBot(a.Config, a.DB, a.Logger, botChan)
	a.Bot = <-botChan
}

func (a *App) StartMntr() {
	a.WG.Add(1)
	go monitor.RunMonitor(a.Config, a.DB, a.Bot, a.Logger)
}

func (a *App) Wait() {
	a.WG.Wait()
}

func (a *App) Close() {
	a.DB.Close()
}

func (a *App) Stop() {
	a.DB.Close()
	a.WG.Wait()
}
