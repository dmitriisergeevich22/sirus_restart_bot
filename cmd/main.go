package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sirus_restart_bot/models"
	"sirus_restart_bot/repository"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

func main() {
	// Инициализация конфигурации
	if err := initConfig(); err != nil {
		log.Println(err)
		return
	}

	// Подключение к базе данных с ресурсами
	db, err := repository.NewMySqlDB()
	if err != nil {
		log.Println(err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(viper.GetString("bot_key"))
	if err != nil {
		log.Fatal(err)
	}

	// bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(fmt.Errorf("error get updates chanel: %w", err))
		return
	}

	// Оповещалка
	go checkRealms(db, bot)

	for update := range updates {
		var message *tgbotapi.Message
		if update.Message != nil {
			message = update.Message
		}
		if update.CallbackQuery != nil {
			message = update.CallbackQuery.Message
		}

		// Сохранение в подписке.
		repository.SaveChatID(db, message.Chat.ID)

		// Запрос на получение информации о серверах.
		realmInfo, err := getRealmInfo()
		if err != nil {
			log.Println(fmt.Errorf("error get realm info: %w", err))
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println(realmInfo)

		text := "Привет! Буду оповещать тебя об перезагрузке сервера =) Текущее состояние серверов:\n"
		for _, r := range realmInfo.Realms {
			var status string
			if r.IsOnline {
				status = "online"
			} else {
				status = "ofline"
			}
			text = text + fmt.Sprintf("Cервер %s. Статус сервера: %s. Онлайн: %d.\n", r.Name, status, r.Online)
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		bot.Send(msg)
	}
}

func checkRealms(db *sqlx.DB, bot *tgbotapi.BotAPI) {
	fmt.Println("Start check servers...")

	for {
		time.Sleep(60 * time.Second)

		// Запрос на получение информации о серверах.
		realmInfo, err := getRealmInfo()
		if err != nil {
			log.Println(fmt.Errorf("error get realm info: %w", err))
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println(realmInfo)

		// проверка реалмов.
		for _, r := range realmInfo.Realms {
			if r.IsOnline {
				continue
			}

			// получить список подписчиков
			listChatID, err := repository.GetListChatID(db)
			if err != nil {
				log.Println("error get list chatID: %w", err)
				return
			}

			// отправить сообщения
			for _, chatID := range listChatID {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Сервер %s в перезагрузе.", r.Name))
				bot.Send(msg)
			}
		}
	}
}

// Инициализация конфигурации.
func initConfig() error {
	viper.AddConfigPath("configs")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error read configuration: %w", err)
	}

	if err := validationConfig(); err != nil {
		return err
	}

	return nil
}

// Валидация конфигурации.
func validationConfig() error {
	keys := []string{
		"db.host",
		"db.port",
		"db.ssl_mode",
		"db.user_name",
		"db.db_name",
		"db.password",
	}

	// Валидация пустой конфигурации.
	for _, key := range keys {
		if viper.GetString(key) == "" {
			return fmt.Errorf("epmty conifgs %s", key)
		}
	}

	return nil
}

func getRealmInfo() (*models.RealmInfo, error) {

	resp, err := http.Get("https://sirus.su/api/statistic/tooltip")
	if err != nil {
		return nil, fmt.Errorf("error get info: %w", resp)
	}

	var realmInfo models.RealmInfo
	if err := json.NewDecoder(resp.Body).Decode(&realmInfo); err != nil {
		return nil, fmt.Errorf("error decode response: %w", err)
	}

	return &realmInfo, nil
}
