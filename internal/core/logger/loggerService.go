package logger

import (
	"log"
	"time"

	"gopkg.in/telebot.v4"
)

type Logger struct {
	isDebug      bool
	loggerBot    *telebot.Bot
	adminID      int64
	hasLoggerBot bool
}

func (logger *Logger) Debug(message string) {
	if !logger.isDebug {
		return
	}
	log.Println("[DEBUG]: " + message)
}

func (logger *Logger) Info(message string) {
	log.Println("[INFO]: " + message)
}

func (logger *Logger) Error(message string) {
	err := "[ERROR]: " + message

	log.Println(err)

	if logger.hasLoggerBot {
		_, err := logger.loggerBot.Send(&telebot.User{ID: logger.adminID}, err)
		if err != nil {
			log.Println("[ERROR]: ERROR WHILE SENDING TO LOGGER BOT: " + err.Error())
		}
	}
}

func MustLogger(isDebug bool, loggerBot *telebot.Bot, adminID int64) *Logger {
	hasLoggerBot := loggerBot != nil && adminID != 0
	return &Logger{isDebug: isDebug, loggerBot: loggerBot, adminID: adminID, hasLoggerBot: hasLoggerBot}
}

func MustLoggerBot(token string) *telebot.Bot {
	if token == "" {
		return nil
	}

	loggerBotSettings := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second}, // TODO
	}

	bot, err := telebot.NewBot(loggerBotSettings)
	if err != nil {
		log.Fatal("Error creating logger bot:", err.Error())
	}

	return bot
}
