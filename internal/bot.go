package bot

import (
	"fmt"
	"log"
	"safeboxtgbot/internal/core/config"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/feat/items"
	"safeboxtgbot/internal/feat/user"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/text"
	"time"

	"gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/react"
)

type Bot struct {
	*telebot.Bot
	Fsm          *fsmManager.FSMState
	UserService  *user.Service
	ItemsService *items.Service
	Config       *config.AppConfig
	Replies      *text.Replies
	Logger       logger.AppLogger
}

func (bot *Bot) MustSend(userID int64, what interface{}, opts ...interface{}) *telebot.Message {
	msg, err := bot.Send(&telebot.User{ID: userID}, what, opts...)
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error sending message to %v: %v", userID, err.Error()))
	}
	return msg
}

func (bot *Bot) MustSendAlbum(userID int64, album telebot.Album) []telebot.Message {
	msg, err := bot.Bot.SendAlbum(&telebot.Chat{ID: userID}, album)
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error sending album to %v: %v", userID, err.Error()))
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

func (bot *Bot) MustEdit(msg telebot.Editable, what interface{}, opts ...interface{}) *telebot.Message {
	m, err := bot.Edit(msg, what, opts...)
	if err != nil {
		bot.Logger.Error(fmt.Sprintf("Error editing message: %v", err.Error()))
	}
	return m
}

func (bot *Bot) RespondSilently(ctx telebot.Context) {
	_ = ctx.Respond()
}

func MustBot(
	config *config.AppConfig,
	fsm *fsmManager.FSMState,
	userService *user.Service,
	itemsService *items.Service,
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
		Bot:          bot,
		Fsm:          fsm,
		UserService:  userService,
		ItemsService: itemsService,
		Config:       config,
		Replies:      replies,
		Logger:       logger,
	}
}
