package notify

import (
	"context"
	"errors"
	"fmt"
	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/core/constants"
	"safeboxtgbot/internal/core/logger"
	"safeboxtgbot/internal/feat/items"
	"safeboxtgbot/internal/feat/prompt"
	"safeboxtgbot/internal/feat/user"
	"safeboxtgbot/internal/helpers"
	"safeboxtgbot/models"
	"safeboxtgbot/pkg/utils"
	"strings"
	"time"
)

type LLMPreviewService struct {
	userService      *user.Service
	itemsService     *items.Service
	messageGenerator prompt.MessageGenerator
	bot              *b.Bot
	logger           logger.AppLogger
}

func NewLLMPreviewService(
	userService *user.Service,
	itemsService *items.Service,
	messageGenerator prompt.MessageGenerator,
	bot *b.Bot,
	logger logger.AppLogger,
) *LLMPreviewService {
	return &LLMPreviewService{
		userService:      userService,
		itemsService:     itemsService,
		messageGenerator: messageGenerator,
		bot:              bot,
		logger:           logger,
	}
}

func (s *LLMPreviewService) SendPreviews(ctx context.Context, userID int64) error {
	if s == nil {
		return errors.New("llm preview service is nil")
	}
	if s.userService == nil {
		return errors.New("user service is nil")
	}
	if s.itemsService == nil {
		return errors.New("items service is nil")
	}
	if s.messageGenerator == nil {
		return errors.New("message generator is nil")
	}
	if s.bot == nil {
		return errors.New("bot is nil")
	}
	if userID == 0 {
		return errors.New("user id is zero")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	userDTO := s.userService.GetUser(userID)
	if userDTO == nil || userDTO.TelegramID == 0 {
		return fmt.Errorf("user not found: %d", userID)
	}

	itemList, err := s.itemsService.GetItemList(userDTO.TelegramID)
	if err != nil {
		return fmt.Errorf("get itemList: %w", err)
	}
	if len(itemList) == 0 {
		return errors.New("user has no itemList")
	}

	loc := previewUserLocation(s.logger, *userDTO)
	localNow := time.Now().In(loc)
	timeOfDayValue := helpers.TimeOfDay(localNow)
	style := helpers.ModeToStyle(userDTO.Mode)

	for _, item := range itemList {
		input := prompt.LLMInput{
			CurrentEntity: item.Name,
			TimeOfDay:     timeOfDayValue,
			StyleMode:     style,
			RandomSeed:    utils.RandomIntRange(1, 1_000_000),
		}

		genCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		text, genErr := s.messageGenerator.Generate(genCtx, input)
		cancel()

		text = strings.TrimSpace(text)
		if genErr != nil || text == "" {
			if s.logger != nil {
				s.logger.Error(fmt.Sprintf("LLM preview generation failed for userID=%d item=%q: %v", userDTO.TelegramID, item.Name, genErr))
			}
			text = helpers.FallbackText(item.Name, constants.FallbackEmojis)
		}

		payload := fmt.Sprintf("%s: %s", item.Name, text)
		s.bot.MustSend(userDTO.TelegramID, payload)
	}

	return nil
}

func previewUserLocation(logger logger.AppLogger, user models.User) *time.Location {
	if user.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("Invalid timezone %q for userID=%d: %v", user.Timezone, user.TelegramID, err))
		}
		return time.UTC
	}
	return loc
}
