package main

import (
	"context"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/config"
	"safeboxtgbot/internal/core/constants"
	l "safeboxtgbot/internal/core/logger"
	d "safeboxtgbot/internal/db"
	"safeboxtgbot/internal/feat/items"
	"safeboxtgbot/internal/feat/notify"
	"safeboxtgbot/internal/feat/prompt"
	"safeboxtgbot/internal/feat/reminder"
	"safeboxtgbot/internal/feat/user"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/handler/commands"
	"safeboxtgbot/internal/handler/keyboard"
	"safeboxtgbot/internal/handler/message"
	"safeboxtgbot/internal/repo"
	"safeboxtgbot/internal/session"
	"safeboxtgbot/internal/text"
)

func main() {
	cfg := config.MustConfig()
	logger := l.MustLogger(cfg.IsDebug, l.MustLoggerBot(cfg.LoggerBotToken), cfg.AdminID)

	db := d.MustDB(cfg, logger)

	fsm := fsmManager.New(logger)
	sessionStore := session.NewStore(constants.NonAuthSessionTTL, logger)
	sessionStore.StartCleanupWorker(context.Background(), constants.NonAuthSessionTTL/2)

	userRepo := repo.NewUserRepo(db)
	itemRepo := repo.NewItemRepo(db)
	reminderRepo := repo.NewReminderRepo(db)
	messageLogRepo := repo.NewMessageLogRepo(db)

	userService := user.NewUserService(userRepo, itemRepo, messageLogRepo, sessionStore, logger)
	itemsService := items.NewService(itemRepo, sessionStore, logger)
	reminderScheduler := reminder.NewScheduler()
	reminderService := reminder.NewService(reminderRepo, reminderScheduler, sessionStore, logger)

	replies := text.NewReplies()
	bot := b.MustBot(cfg, fsm, userService, itemsService, reminderService, replies, logger)

	commands.MustInitCommandsHandler(bot)
	keyboard.MustInitKeyboardHandler(bot)
	message.MustInitMessagesHandler(bot)

	llmClient := prompt.MustNewOpenRouterClient(cfg.OpenRouterModelApiKey, logger)
	var groqFallback prompt.LLMGenerator
	if cfg.GroqAPIKey != "" {
		groqClient := prompt.MustNewGroqClient(cfg.GroqAPIKey, logger)
		groqFallback = prompt.MustNewGroqService(groqClient, cfg.GroqModelName, logger)
	}
	llmService := prompt.MustNewLLMService(llmClient, cfg.OpenRouterModelName, groqFallback, logger)
	promptBuilder := prompt.MustNewPromptBuilder(cfg.PromptPath, logger)
	messageGenerator := prompt.MustNewMessageGenerator(promptBuilder, llmService, logger)

	commands.MustInitAdminCommandsHandler(bot, messageGenerator, promptBuilder, groqFallback, cfg.ForcePreviewFallback)

	notifyWorker := notify.NewWorker(userService, itemsService, messageLogRepo, messageGenerator, bot, logger)
	go notifyWorker.Start(context.Background())

	reminderWorker := reminder.NewWorker(reminderService, userService, messageGenerator, bot.Bot, logger)
	go reminderWorker.Start(context.Background())

	logger.Info("Bot successfully started!")
	bot.Start()
}
