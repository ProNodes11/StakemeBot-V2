package users

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

type AlertStatus int

const (
    MinorAlert     AlertStatus = iota // Незначительный алерт
    SeriousAlert                       // Серьезный алерт
    CriticalAlert                      // Критический алерт
    JailAlert                          // Алерт про тюрьму
)

type User struct {
	ChatID        int64
	State         State
	UptimeAlerts  []UptimeAlert
	ProposalAlert []ProposalAlert  
}

type State struct {
	MessageID          int
	WaitForAddress     bool
	CurrentChain       string
}

type UptimeAlert struct {
	ChainName  string 
	Address    string 
	SentAlerts map[AlertStatus]bool
}

type ProposalAlert struct {
	ChainName   string 
	ProposalID  int 
	SentAlerts  bool
}

func (ua *UptimeAlert) SendAlert(bot *tgbotapi.BotAPI, message string, status AlertStatus, toUser int64, logger *logrus.Logger) error {
	if ua.SentAlerts[status] {
        return nil // Алерт уже был отправлен
    }

    var priority string
    switch status {
    case MinorAlert:
        priority = "Minor Alert"
    case SeriousAlert:
        priority = "Serious Alert"
    case CriticalAlert:
        priority = "Critical Alert"
    case JailAlert:
        priority = "Jail"
    default:
        priority = "Unknown Alert"
    }

	text := fmt.Sprintf("*New alert!* \n*Priority:* %s \n\n%s", priority, message)
    msg := tgbotapi.NewMessage(toUser, text)
	msg.ParseMode = "markdown"

    _, err := bot.Send(msg)
	if err == nil {
        ua.SentAlerts[status] = true // Отметить алерт как отправленный
		logger.WithFields(logrus.Fields{
			"module": "alert",
			"chain":  ua.ChainName,
			"alert_level": priority,
		}).Info("Successfully send alert to user")
    }
    return err
}

func (ua *UptimeAlert) ResetSentAlerts() {
	for status := range ua.SentAlerts {
		ua.SentAlerts[status] = false
	}
}

func NewUser(chatID int64) *User {
	return &User{
		ChatID:        chatID,
		State:         State{},
		UptimeAlerts:  []UptimeAlert{},
		ProposalAlert: []ProposalAlert{},
	}
}

func NewUptimeAlert(chainName, address string) UptimeAlert {
	return UptimeAlert{
		ChainName:  chainName,
		Address:    address,
		SentAlerts: map[AlertStatus]bool{
			MinorAlert:   false,
			SeriousAlert: false,
			CriticalAlert: false,
			JailAlert:    false,
		},
	}
}

func (user *User) AddUptimeAlert(chainName, address string) {
	uptimeAlert := NewUptimeAlert(chainName, address)
	user.UptimeAlerts = append(user.UptimeAlerts, uptimeAlert)
}