package main

import (
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/config"
	l "safeboxtgbot/internal/core/logger"
	d "safeboxtgbot/internal/db"
	"safeboxtgbot/internal/feat/user"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/handlers/commands"
	"safeboxtgbot/internal/handlers/message"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/internal/text"
)

func main() {
	cfg := config.MustConfig()
	logger := l.MustLogger(cfg.IsDebug, l.MustLoggerBot(cfg.LoggerBotToken), cfg.AdminID)

	db := d.MustDB(cfg, logger)

	fsm := fsmManager.New(logger)
	sessionStore := session.NewStore(logger)

	userRepo := repo.NewUserRepo(db)
	itemRepo := repo.NewItemRepo(db)
	messageLogRepo := repo.NewMessageLogRepo(db)

	userService := user.NewUserService(userRepo, itemRepo, messageLogRepo, sessionStore, logger)

	replies := text.NewReplies()
	bot := b.MustBot(cfg, fsm, userService, replies, logger)

	commands.MustInitCommandsHandler(bot)
	message.MustInitMessagesHandler(bot)

	logger.Info("Bot successfully started!")
	bot.Start()
}
