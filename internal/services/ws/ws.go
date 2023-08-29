package ws

import (
	"StakemeBotV2/internal/config"
	"StakemeBotV2/internal/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	pbtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	QueryNewBlock string = `tm.event='NewBlock'`
	QueryVote     string = `tm.event='Vote'`
)

type Block struct {
	Height int64  `json:"height"`
	Hash   string `json:"hash"`
}

type WsReply struct {
	Id     int64 `json:"id"`
	Result struct {
		Query string `json:"query"`
		Data  struct {
			Type  string          `json:"type"`
			Value json.RawMessage `json:"value"`
		} `json:"data"`
	} `json:"result"`
}

type rawVote struct {
	Vote struct {
		Type             pbtypes.SignedMsgType `json:"type"`
		Height           db.StringInt64                 `json:"height"`
		ValidatorAddress string                `json:"validator_address"`
	} `json:"Vote"`
}

func RunWebsocketConnection(cnfg config.Config, logger *logrus.Logger, dbcon *sql.DB) {
	var wg sync.WaitGroup
	for _, chain := range cnfg.Chains {
		wg.Add(1)
		go InitWebsocketConnection(chain.Name, chain.RPCUrls, logger, &wg, dbcon)
		// for _, rpcURL := range chain.RPCUrls {
		// 	wg.Add(1)
		// 	go SubscribeToWebSocket(chain.Name, rpcURL, logger, &wg, dbcon)
		// }
	}

	logger.WithField("module", "websocket").Debug("Start load module wallet")
	wg.Wait()
}

func connectWebSocket(chainName, wsUrl string, logger *logrus.Logger, wg *sync.WaitGroup, dbcon *sql.DB) (*websocket.Conn, error) {
	// Проверяем юпл
	u, err := url.Parse(wsUrl + "/websocket")
	if err != nil {
		logger.WithField("module", "websocket").WithField("chain", chainName).Errorf("Error parsing URL: %v", err)
		return nil, err
	}
	
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func SubscribeWebSocket(chainName, url string, conn *websocket.Conn, logger *logrus.Logger) {
	for _, subscribe := range []string{QueryNewBlock, QueryVote} {
		q := fmt.Sprintf(`{"jsonrpc":"2.0","method":"subscribe","id":1,"params":{"query":"%s"}}`, subscribe)
		err := conn.WriteMessage(websocket.TextMessage, []byte(q))
		if err != nil {
			logger.WithFields(logrus.Fields{
				"module": "websocket",
				"chain":  chainName,
				"websocket": url,
			}).Debugf("Failed subscribe %d", err)
		}
	}
}

func InitWebsocketConnection(chainName string, urls []string, logger *logrus.Logger, wg *sync.WaitGroup, dbcon *sql.DB) {
	// defer wg.Done()

	for {
		for _, url := range urls {
			wsURL := strings.Replace(url, "https://", "wss://", 1)
			conn, err := connectWebSocket(chainName, wsURL, logger, wg, dbcon)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"module": "websocket",
					"chain":  chainName,
					"websocket": url,
				}).Error("Failed connected to websocket")
				continue
			}

			logger.WithFields(logrus.Fields{
				"module": "websocket",
				"chain":  chainName,
				"websocket": url,
			}).Info("Succesfully connected to websocket")

			logger.WithFields(logrus.Fields{
				"module": "websocket",
				"chain":  chainName,
				"websocket": url,
			}).Debug("Start subscribe")

			// Создаем подписку на данные
			SubscribeWebSocket(chainName, url, conn, logger)

			logger.WithFields(logrus.Fields{
				"module": "websocket",
				"chain":  chainName,
				"websocket": url,
			}).Debug("Start read message")

			// Получаем сообщение от вебсокета
			ReadWebSocket(conn, logger, chainName, url, dbcon)
			conn.Close()
		}
		logger.WithFields(logrus.Fields{
			"module": "websocket",
			"chain":  chainName,
		}).Warnf("All connection for chain %s closed, reconecting", chainName)
		time.Sleep(time.Minute)
	}
}



func ReadWebSocket(conn *websocket.Conn, logger *logrus.Logger, chainName, rpcURL string, dbcon *sql.DB) {
	blockData := db.TableData{
		BlockNumber: "10",
	}
	// Вечный цикл для получения сообщений
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.WithField("module", "websocket").WithField("chain", chainName).Errorf("Error reading message: %v", err)
			break
		}

		var errorCounter int
		// Читаем сообщение
		if err := handleWebSocketMessage(message, logger, chainName, rpcURL, &blockData, dbcon); err != nil {
			errorCounter ++
		}
		if errorCounter > 2 {
			// break
		}
	}
}


func handleWebSocketMessage(message []byte, logger *logrus.Logger, chainName, rpcURL string, blockData *db.TableData, dbcon *sql.DB) error {
	// Анмаршалим сообщения с вебсокета в структуру
	var response WsReply
	if err := json.Unmarshal(message, &response); err != nil {
		logger.WithFields(logrus.Fields{
			"module":    "websocket",
			"chain":     chainName,
			"websocket": rpcURL,
		}).Errorf("Failed read message %d", err)
		return err
	}

	vote := &rawVote{}
	if err := json.Unmarshal([]byte(response.Result.Data.Value), vote); err != nil {
		logger.WithFields(logrus.Fields{
			"module":    "websocket",
			"chain":     chainName,
			"websocket": rpcURL,
		}).Error(err)
		return err
	}
	
	// Фильтрация типов подписи, если блок один тот же собираем всех кто его подписал
	// Если появился новый блок, тогда сохраняем старые данные, очищаем переменную и заполняем по новой
	if vote.Vote.Type == pbtypes.PrevoteType {
		if blockData.BlockNumber != vote.Vote.Height {
			if blockData.BlockNumber != "10" {
				if err := db.InsertData(dbcon, chainName, blockData); err != nil {
					logger.WithFields(logrus.Fields{
						"module":    "websocket",
						"chain":     chainName,
						"websocket": rpcURL,
					}).Errorf("Failed to save data %s", err)
				}
			}
			
			logger.WithFields(logrus.Fields{
				"module":    "websocket",
				"chain":     chainName,
				"websocket": rpcURL,
			}).Infof("Succesfully save data for block %s", blockData.BlockNumber)

			blockData.BlockNumber = vote.Vote.Height
			blockData.SignedAddresses = []string{}
		} else {
			blockData.SignedAddresses = append(blockData.SignedAddresses, vote.Vote.ValidatorAddress)
		}
	}
	return nil
}
