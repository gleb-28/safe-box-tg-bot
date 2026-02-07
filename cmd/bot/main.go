package main

import (
	"context"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/config"
	l "safeboxtgbot/internal/core/logger"
	d "safeboxtgbot/internal/db"
	"safeboxtgbot/internal/feat/items"
	"safeboxtgbot/internal/feat/notify"
	"safeboxtgbot/internal/feat/prompt"
	"safeboxtgbot/internal/feat/user"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/handlers/commands"
	"safeboxtgbot/internal/handlers/keyboard"
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
	itemsService := items.NewService(itemRepo, sessionStore, logger)

	replies := text.NewReplies()
	bot := b.MustBot(cfg, fsm, userService, itemsService, replies, logger)

	commands.MustInitCommandsHandler(bot)
	keyboard.MustInitKeyboardHandler(bot)
	message.MustInitMessagesHandler(bot)

	llmClient := prompt.MustNewOpenRouterClient(cfg.ModelApiKey, logger)
	llmService := prompt.MustNewLLMService(llmClient, cfg.ModelName, logger)
	promptBuilder := prompt.MustNewPromptBuilder(cfg.PromptPath, logger)
	messageGenerator := prompt.MustNewMessageGenerator(promptBuilder, llmService, logger)

	notifyWorker := notify.NewWorker(userService, itemsService, messageLogRepo, messageGenerator, bot, logger)
	go notifyWorker.Start(context.Background())

	logger.Info("Bot successfully started!")
	bot.Start()
}
