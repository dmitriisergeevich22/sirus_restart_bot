package repository

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type Repository struct {
	DB *sqlx.DB
}

// Конфигурация.
type MySQLConfig struct {
	Host     string
	Port     string
	Username string
	DBName   string
	SSLMode  string
	Password string
}

// Подключение к БД.
func NewMySqlDB() (*sqlx.DB, error) {
	cfg := MySQLConfig{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		SSLMode:  viper.GetString("db.ssl_mode"),
		Username: viper.GetString("db.user_name"),
		DBName:   viper.GetString("db.db_name"),
		Password: viper.GetString("db.password"),
	}

	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName))

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func SaveChatID(db *sqlx.DB, chatId int64) error {
	query := fmt.Sprintf("INSERT INTO sirus5.chat_bot (chat_id) VALUES (%d);", chatId)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error save chatID: %w", err)
	}

	return nil
}

func GetListChatID(db *sqlx.DB) ([]int64, error) {
	rows, err := db.Query("SELECT chat_id FROM sirus5.chat_bot;")
	if err != nil {
		return nil, fmt.Errorf("error save chatID: %w", err)
	}
	defer rows.Close()

	chatIDs := make([]int64, 0, 50)
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("error scan chatID: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}
	fmt.Println(chatIDs)

	return chatIDs, nil
}
