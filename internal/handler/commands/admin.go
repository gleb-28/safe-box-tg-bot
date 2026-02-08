package commands

import (
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/feat/prompt"
	adminMiddleware "safeboxtgbot/internal/middleware/admin"
)

func MustInitAdminCommandsHandler(bot *b.Bot, messageGenerator prompt.MessageGenerator) {
	bot.Handle("/preview_llm", createPreviewLLMHandler(bot, messageGenerator), adminMiddleware.CreateAdminMiddleware(bot))
}
