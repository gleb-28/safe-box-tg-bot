package bot

import (
	"fmt"
	"log"
	"safeboxtgbot/internal/core/config"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/feat/user"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/text"
	"time"

	"gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/react"
)

type Bot struct {
	*telebot.Bot
	Fsm         *fsmManager.FSMState
	UserService *user.Service
	Config      *config.AppConfig
	Replies     *text.Replies
	Logger      logger.AppLogger
}

func (bot *Bot) MustSend(chatId int64, what interface{}, opts ...interface{}) *telebot.Message {
	msg, err := bot.Send(&telebot.User{ID: chatId}, what, opts...)
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error sending message to %v: %v", chatId, err.Error()))
	}
	return msg
}

func (bot *Bot) MustSendAlbum(chatID int64, album telebot.Album) []telebot.Message {
	msg, err := bot.Bot.SendAlbum(&telebot.Chat{ID: chatID}, album)
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error sending album to %v: %v", chatID, err.Error()))
	}
	return msg
}

func (bot *Bot) MustDelete(msg *telebot.Message) {
	err := bot.Delete(msg)
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error deleting message to %v: %v", msg.Chat.ID, err.Error()))
	}
}

func (bot *Bot) MustReact(msg *telebot.Message, reaction telebot.Reaction) {
	err := bot.React(&telebot.User{ID: msg.Chat.ID}, msg, react.React(reaction))
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error reacting message to %v: %v", msg.Chat.ID, err.Error()))
	}
}

// TODO
func MustEdit() {}

func MustBot(
	config *config.AppConfig,
	fsm *fsmManager.FSMState,
	userService *user.Service,
	replies *text.Replies,
	logger logger.AppLogger) *Bot {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:     config.TgBotToken,
		Poller:    &telebot.LongPoller{Timeout: 10 * time.Second},
		ParseMode: telebot.ModeHTML,
		OnError: func(err error, context telebot.Context) {
			logger.Error(err.Error())
		},
	})
	if err != nil {
		log.Fatal("Error creating bot:", err.Error())
	}

	return &Bot{
		bot,
		fsm,
		userService,
		config,
		replies,
		logger,
	}
}
