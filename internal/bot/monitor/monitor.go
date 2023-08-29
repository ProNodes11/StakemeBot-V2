package monitor

import (
	"StakemeBotV2/internal/bot/users"
	"StakemeBotV2/internal/config"
	"StakemeBotV2/internal/db"
	"StakemeBotV2/internal/services/api/rest"
	"database/sql"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

func RunMonitor(cnfg config.Config, dbcon *sql.DB, bot *tgbotapi.BotAPI, logger *logrus.Logger) {
	for {	
		dbUsers, err := db.GetUsers(dbcon) 
		if err != nil {
			logger.WithFields(logrus.Fields{
				"module": "uptime",
			}).Errorf("Failed to get users %s", err)
		}
		for _, user := range dbUsers {
			for i, chain := range user.UptimeAlerts {
				address := rest.GetHexAddress(cnfg.GetAPIUrlsForChain(chain.ChainName), chain.Address)
				blockCount, err := db.CountSignedBlocks(dbcon, chain.ChainName)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"module": "uptime",
					}).Errorf("Failed to get count signed blocks %s", err)
				}
				signedblockCount, err := db.CountSignedBlocksByAddress(dbcon, chain.ChainName, address)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"module": "uptime",
					}).Errorf("Failed to get count signed blocks by address %s %s", address, err)
				}
				if blockCount == 0 {
					continue
				}
				uptime := (signedblockCount * 100) / blockCount
				logger.WithFields(logrus.Fields{
					"module": "uptime",
					"chain": chain.ChainName,
					"address": chain.Address,
				}).Infof("Uptime = %d %%", uptime)
				
				if jailStatus, _ := rest.GetBoundStateByAddress(cnfg.GetAPIUrlsForChain(chain.ChainName), chain.Address); jailStatus {
					explorer := fmt.Sprintf("https://www.mintscan.io/%s/validators/%s", strings.ToLower(chain.ChainName) , chain.Address)
					message := fmt.Sprintf("*Message:* your validator jailed \n*Chain:* %s\n*VallAddress* %s\n*Uptime* %d\n*Explorer:* %s", chain.ChainName, chain.Address, uptime, explorer)
					user.UptimeAlerts[i].SendAlert(bot, message, users.JailAlert, user.ChatID, logger)
					chain.SentAlerts[users.JailAlert] = true
					db.UpdateUser(dbcon, &user, logger)
					continue
				} else {
					chain.SentAlerts[users.JailAlert] = false
					db.UpdateUser(dbcon, &user, logger)
				}
				
				if uptime < 30 {
					message := fmt.Sprintf("*Chain:* %s\n*ValAddress:* %s\n*Uptime* %d", chain.ChainName,chain.Address, uptime)
					user.UptimeAlerts[i].SendAlert(bot, message, users.CriticalAlert, user.ChatID, logger)
					chain.SentAlerts[users.CriticalAlert] = true
				} else if uptime < 60 {
					message := fmt.Sprintf("Chain: %s\nValAddress: %s\nUptime %d", chain.ChainName,chain.Address, uptime)
					user.UptimeAlerts[i].SendAlert(bot, message, users.SeriousAlert, user.ChatID, logger)
					chain.SentAlerts[users.SeriousAlert] = true
				} else if uptime < 101 {
					message := fmt.Sprintf("Chain: %s\nValAddress: %s\nUptime %d", chain.ChainName,chain.Address, uptime)
					user.UptimeAlerts[i].SendAlert(bot, message, users.MinorAlert, user.ChatID, logger)
					chain.SentAlerts[users.MinorAlert] = true
				} else {
					message := fmt.Sprintf("Validator uptime returned to normal state\nChain: %s\nUptime %d", chain.ChainName, uptime)
					user.UptimeAlerts[i].SendAlert(bot, message, users.MinorAlert, user.ChatID, logger)
					chain.ResetSentAlerts()
				}
				db.UpdateUser(dbcon, &user, logger)
			}
		}
		time.Sleep(time.Second * 10)
	}
}