package commands

import (
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/feat/prompt"
	adminMiddleware "safeboxtgbot/internal/middleware/admin"
)

func MustInitAdminCommandsHandler(bot *b.Bot, messageGenerator prompt.MessageGenerator, builder prompt.PromptBuilder, fallback prompt.LLMGenerator, forcePreviewFallback bool) {
	bot.Handle("/preview_llm", createPreviewLLMHandler(bot, messageGenerator, builder, fallback, forcePreviewFallback), adminMiddleware.CreateAdminMiddleware(bot))
}
