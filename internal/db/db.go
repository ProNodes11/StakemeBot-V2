package db

import (
	"StakemeBotV2/internal/bot/users"
	"StakemeBotV2/internal/config"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type TableData struct {
	BlockNumber     StringInt64
	SignedAddresses []string
}

type StringInt64 string

// helper to make the "everything is a string" issue less painful.
func (si StringInt64) val() int64 {
	i, _ := strconv.ParseInt(string(si), 10, 64)
	return i
}


const tableExistsQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_name = $1
		)
	`

func RunAndConfigDatabase(cnfg config.Config, logger *logrus.Logger) *sql.DB {
	
	db := ConnectDatabase(cnfg.Db, logger)
	
	logger.WithFields(logrus.Fields{
		"module":    "db",
	}).Info("Connected to database")
	
	// Цикл для проверки каждой сетки в бд
	for _, chain := range cnfg.Chains {
		tableName := strings.ToLower(chain.Name)
		logger.WithFields(logrus.Fields{
			"module":    "db",
		}).Debugf("Start checking table for chain %s",tableName)

		// Проверка на существование таблицы
		tableExists, err := CheckTableExists(db, tableName)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"module":    "db",
			}).Fatalf("Failed to check table for chain %s %s",tableName, err)
		}

		// Создание таблицы, если она не существует
		if !tableExists {

			logger.WithFields(logrus.Fields{
				"module":    "db",
			}).Debugf("Start create table for chain %s",tableName)

			err = CreateTable(db, tableName)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"module":    "db",
				}).Fatalf("Failed to create table for chain %s %s", tableName, err)
			}
		}
	}
	CreateUserTable(db, logger)

	logger.WithFields(logrus.Fields{
		"module":    "db",
	}).Debug("Finish to configure database")

	return db
}

func ConnectDatabase(dbd config.DbConfigData, logger *logrus.Logger) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	dbd.Host, dbd.Port, dbd.User, dbd.Pass, "postgres")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"module":    "db",
		}).Fatalf("Failed to connect to database %d", err)
	}
	// defer db.Close()
	return db
}

func CheckTableExists(db *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := db.QueryRow(tableExistsQuery, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateTable(db *sql.DB, tableName string) error {
	query := fmt.Sprintf("CREATE TABLE %s (block_number INT PRIMARY KEY, signed_addresses TEXT[]);", tableName)
	_, err := db.Exec(query)
	return err
}

func InsertData(db *sql.DB, tableName string, data *TableData) error {
	query := fmt.Sprintf("INSERT INTO %s (block_number, signed_addresses) VALUES ($1, $2);", tableName)
	_, err := db.Exec(query, data.BlockNumber, pq.StringArray(data.SignedAddresses))
	return err
}


func ReadData(db *sql.DB, tableName string) error {
	rows, err := db.Query("SELECT * FROM kichain;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log.Println(rows)
	// Проходим по всем строкам результатов и выводим их
	for rows.Next() {
		var blockNumber int
		var signedAddresses pq.StringArray  // Используем pq.StringArray
		if err := rows.Scan(&blockNumber, &signedAddresses); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("BlockNumber: %d, SignedAddresses: %v\n", blockNumber, signedAddresses)
	}

	// Обработка ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return nil
}


func CountSignedBlocksByAddress(db *sql.DB, tableName string, address string) (int, error) {
	query := `
		SELECT SUM(CASE WHEN $1 = ANY(signed_addresses) THEN 1 ELSE 0 END) AS total_count
		FROM ` + tableName + `;
	`
	var totalCount int
	err := db.QueryRow(query, address).Scan(&totalCount)
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}

func CountSignedBlocks(db *sql.DB, tableName string) (int, error) {
	query := `
		SELECT COUNT(*) FROM ` + tableName +`;
	`
	var totalCount int
	err := db.QueryRow(query).Scan(&totalCount)
	if err != nil {
		return 0, err
	}

	// fmt.Printf("Общее количество блоков: %d\n", totalCount)
	return totalCount, nil
}

// func UpsertUser(db *sql.DB, user *users.User) error {
//     selectQuery := "SELECT COUNT(*) FROM users WHERE chat_id = $1"
//     var count int
//     err := db.QueryRow(selectQuery, user.ChatID).Scan(&count)
//     if err != nil {
//         return err
//     }

//     if count == 0 {
//         // Пользователь не существует, добавляем его
//         return InsertUser(db, user)
//     }

//     // Пользователь существует, обновляем его данные
//     // return UpdateUser(db, user)
// 	return nil
// }


// func InsertUser(db *sql.DB, user *users.User) error {
//     query := `
//         INSERT INTO users (chat_id, uptime_alerts, proposal_alerts)
//         VALUES ($1, $2, $3)
//         ON CONFLICT (chat_id) DO UPDATE
//         SET uptime_alerts = EXCLUDED.uptime_alerts,
//             proposal_alerts = EXCLUDED.proposal_alerts;
//     `

//     uptimeAlertsJSON, err := json.Marshal(user.UptimeAlerts)
//     if err != nil {
//         return err
//     }

//     proposalAlertsJSON, err := json.Marshal(user.ProposalAlert)
//     if err != nil {
//         return err
//     }

//     _, err = db.Exec(query, user.ChatID, uptimeAlertsJSON, proposalAlertsJSON)
//     return err
// }

// func UpdateUser(db *sql.DB, user *users.User) error {
//     query := `
//         UPDATE users
//         SET uptime_alerts = $2, proposal_alerts = $3
//         WHERE chat_id = $1;
//     `

//     uptimeAlertsJSON, err := json.Marshal(user.UptimeAlerts)
//     if err != nil {
//         return err
//     }

//     proposalAlertsJSON, err := json.Marshal(user.ProposalAlert)
//     if err != nil {
//         return err
//     }

//     _, err = db.Exec(query, user.ChatID, uptimeAlertsJSON, proposalAlertsJSON)
//     return err
// }


func CreateUserTable(db *sql.DB, logger *logrus.Logger) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			chat_id BIGINT PRIMARY KEY,
			state JSONB,
			uptime_alerts JSONB,
			proposal_alerts JSONB
		);
	`)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"module": "db",
		}).Errorf("Failed to create user table: %v", err)
		return
	}
	logger.WithFields(logrus.Fields{
		"module": "db",
	}).Debug("Successfully created user table")
}

func GetUsers(db *sql.DB) ([]users.User, error) {
	query := "SELECT chat_id, state, uptime_alerts, proposal_alerts FROM users"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbUsers []users.User
	for rows.Next() {
		var user users.User
		var stateJSON, uptimeAlertsJSON, proposalAlertsJSON []byte
		err := rows.Scan(&user.ChatID, &stateJSON, &uptimeAlertsJSON, &proposalAlertsJSON)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(stateJSON, &user.State)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(uptimeAlertsJSON, &user.UptimeAlerts)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(proposalAlertsJSON, &user.ProposalAlert)
		if err != nil {
			return nil, err
		}

		dbUsers = append(dbUsers, user)
	}

	return dbUsers, nil
}


func GetUserByChatID(db *sql.DB, chatID int64, logger *logrus.Logger) (*users.User, error) {
	// Проверяем наличие пользователя в базе данных
	user, err := getUserByChatID(db, chatID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.WithFields(logrus.Fields{
			"module": "db",
		}).Errorf("Failed to get user: %v", err)
		return nil, err
	}

	// Если пользователя нет, создаем и возвращаем нового пользователя
	if user == nil {
		newUser := users.NewUser(chatID)
		if err := createUser(db, chatID, logger); err != nil {
			return nil, err
		}
		return newUser, nil
	}

	// Пользователь найден, возвращаем его
	return user, nil
}

func getUserByChatID(db *sql.DB, chatID int64) (*users.User, error) {
	var user users.User
	var stateJSON, uptimeAlertsJSON, proposalAlertsJSON []byte 
	err := db.QueryRow(`
		SELECT chat_id, state, uptime_alerts, proposal_alerts
		FROM users WHERE chat_id = $1
	`, chatID).Scan(
		&user.ChatID, &stateJSON, &uptimeAlertsJSON, &proposalAlertsJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(stateJSON, &user.State)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(uptimeAlertsJSON, &user.UptimeAlerts)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(proposalAlertsJSON, &user.ProposalAlert)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func createUser(db *sql.DB, chatID int64, logger *logrus.Logger) error {
	_, err := db.Exec(`
		INSERT INTO users (chat_id, state, uptime_alerts, proposal_alerts)
		VALUES ($1, $2, $3, $4)
	`, chatID, "{}", "[]", "[]")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"module": "db",
		}).Errorf("Failed to create user: %v", err)
		return err
	}
	logger.WithFields(logrus.Fields{
		"module": "db",
	}).Debug("Successfully created user")
	return nil
}

func UpdateUser(db *sql.DB, user *users.User, logger *logrus.Logger) error {
	stateJSON, err := json.Marshal(user.State)
	if err != nil {
		return err
	}

	uptimeAlertsJSON, err := json.Marshal(user.UptimeAlerts)
	if err != nil {
		return err
	}

	proposalAlertsJSON, err := json.Marshal(user.ProposalAlert)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE users
		SET state = $1, uptime_alerts = $2, proposal_alerts = $3
		WHERE chat_id = $4
	`, stateJSON, uptimeAlertsJSON, proposalAlertsJSON, user.ChatID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"module": "db",
		}).Errorf("Failed to update user: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"module": "db",
	}).Debug("Successfully updated user")

	return nil
}
