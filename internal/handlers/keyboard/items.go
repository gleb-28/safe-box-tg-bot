package keyboard

import (
	"context"
	"errors"
	"fmt"
	"strings"

	b "safeboxtgbot/internal"
	"safeboxtgbot/internal/feat/items"
	fsmManager "safeboxtgbot/internal/fsm"
	"safeboxtgbot/internal/handlers/auth"
	"safeboxtgbot/models"

	"gopkg.in/telebot.v4"
)

var (
	btnAddItem            = telebot.Btn{Unique: "btn_add_item", Text: "➕ Добавить"}
	btnEditItem           = telebot.Btn{Unique: "btn_edit_item", Text: "✏️ Изменить"}
	btnDeleteItem         = telebot.Btn{Unique: "btn_delete_item", Text: "🗑 Удалить"}
	btnCloseItemBox       = telebot.Btn{Unique: "btn_close_item_box", Text: "✖️ Закрыть"}
	btnBackToItemBox      = telebot.Btn{Unique: "btn_back_to_item_box", Text: "⬅️ Назад"}
	btnSelectItemToEdit   = telebot.Btn{Unique: "btn_select_item_to_edit"}
	btnSelectItemToDelete = telebot.Btn{Unique: "btn_select_item_to_delete"}
)

func MustInitItemBoxButtons(bot *b.Bot) {
	bot.Handle(&btnAddItem, createAddItemHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnEditItem, createEditItemHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnDeleteItem, createDeleteItemHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnCloseItemBox, createCloseItemBoxHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnBackToItemBox, createBackToItemBoxHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnSelectItemToEdit, createEditItemSelectHandler(bot), auth.CreateAuthMiddleware(bot))
	bot.Handle(&btnSelectItemToDelete, createDeleteItemSelectHandler(bot), auth.CreateAuthMiddleware(bot))
}

func OpenItemBox(bot *b.Bot, userID int64, sourceMsg *telebot.Message) error {
	clearClosedItemBoxMessage(bot, userID)
	bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
	bot.ItemsService.ClearEditingItemName(userID)
	return renderItemBox(bot, userID, sourceMsg)
}

func createAddItemHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.AwaitingItemAddEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return renderAddItemPrompt(bot, userID, ctx.Message(), "")
	}
}

func CreateValidateAddItemHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		err := bot.ItemsService.CreateItem(userID, ctx.Message().Text)
		bot.MustDelete(ctx.Message())
		if err != nil {
			return handleItemInputError(bot, userID, err, true)
		}

		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return renderItemBox(bot, userID, nil)
	}
}

func createEditItemHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemEditSelectOpenedEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return renderEditItemSelectPrompt(bot, userID, ctx.Message())
	}
}

func CreateValidateEditItemHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		itemName := bot.ItemsService.GetEditingItemName(userID)
		if itemName == "" {
			bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
			return renderItemBox(bot, userID, nil)
		}

		err := bot.ItemsService.UpdateItemName(userID, itemName, ctx.Message().Text)
		bot.MustDelete(ctx.Message())
		if err != nil {
			return handleItemInputError(bot, userID, err, false)
		}

		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return renderItemBox(bot, userID, nil)
	}
}

func createDeleteItemHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemDeleteSelectOpenedEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return renderDeleteSelect(bot, userID, ctx.Message())
	}
}

func createCloseItemBoxHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		bot.ItemsService.ClearEditingItemName(userID)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.InitialEvent)
		bot.ItemsService.SetBotLastMsg(userID, nil)
		clearClosedItemBoxMessage(bot, userID)
		msg := bot.MustSend(userID, bot.Replies.ItemBoxClosed, MainMenuKeyboard())
		saveClosedItemBoxMessage(bot, userID, msg)
		bot.MustDelete(ctx.Message())

		return nil
	}
}

func createBackToItemBoxHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		bot.ItemsService.ClearEditingItemName(userID)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
		return renderItemBox(bot, userID, ctx.Message())
	}
}

func createEditItemSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		itemName, err := parseItemName(ctx)
		if err != nil {
			return renderItemBox(bot, userID, ctx.Message())
		}

		bot.ItemsService.SetEditingItemName(userID, itemName)
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.AwaitingItemEditEvent)
		return renderEditItemPrompt(bot, userID, ctx.Message(), itemName, "")
	}
}

func createDeleteItemSelectHandler(bot *b.Bot) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		userID := ctx.Chat().ID
		respondSilently(ctx)
		itemName, err := parseItemName(ctx)
		if err != nil {
			return renderItemBox(bot, userID, ctx.Message())
		}

		if err := bot.ItemsService.DeleteItem(userID, itemName); err != nil {
			if errors.Is(err, items.ErrItemNotFound) {
				bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
				return renderItemBox(bot, userID, ctx.Message())
			}
			return upsertBotLastMessage(bot, userID, ctx.Message(), bot.Replies.Error, itemBoxMarkup())
		}

		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
		return renderItemBox(bot, userID, ctx.Message())
	}
}

func renderItemBox(bot *b.Bot, userID int64, sourceMsg *telebot.Message) error {
	itemList, err := bot.ItemsService.GetItemList(userID)
	if err != nil {
		return upsertBotLastMessage(bot, userID, sourceMsg, bot.Replies.Error, itemBoxMarkup())
	}

	var text string
	if len(itemList) == 0 {
		text = bot.Replies.ItemsMenuEmpty
	} else {
		var builder strings.Builder
		builder.WriteString(bot.Replies.ItemsMenuHeader)
		for _, item := range itemList {
			builder.WriteString(bot.Replies.ItemsMenuItemPrefix)
			builder.WriteString(item.Name)
			builder.WriteString("\n")
		}
		builder.WriteString(bot.Replies.ItemsMenuFooter)
		text = builder.String()
	}

	return upsertBotLastMessage(bot, userID, sourceMsg, text, itemBoxMarkup())
}

func renderAddItemPrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, note string) error {
	text := bot.Replies.AddNewItem
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertBotLastMessage(bot, userID, sourceMsg, text, backToItemBoxMarkup())
}

func renderEditItemSelectPrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message) error {
	itemList, err := bot.ItemsService.GetItemList(userID)
	if err != nil {
		return upsertBotLastMessage(bot, userID, sourceMsg, bot.Replies.Error, itemBoxMarkup())
	}

	text := bot.Replies.WhatDoWeEdit
	if len(itemList) == 0 {
		text = bot.Replies.ListIsEmpty
	}
	return upsertBotLastMessage(bot, userID, sourceMsg, text, selectItemMarkup(itemList, btnSelectItemToEdit, func(item models.Item) string {
		return item.Name
	}))
}

func renderEditItemPrompt(bot *b.Bot, userID int64, sourceMsg *telebot.Message, itemName string, note string) error {
	name := ""
	if entries, err := bot.ItemsService.GetItemList(userID); err == nil {
		for _, item := range entries {
			if item.Name == itemName {
				name = item.Name
				break
			}
		}
	}
	text := bot.Replies.WriteNewItemName
	if name != "" {
		text = fmt.Sprintf(bot.Replies.NewNameForValue, name)
	}
	if note != "" {
		text = note + "\n\n" + text
	}
	return upsertBotLastMessage(bot, userID, sourceMsg, text, backToItemBoxMarkup())
}

func renderDeleteSelect(bot *b.Bot, userID int64, sourceMsg *telebot.Message) error {
	itemList, err := bot.ItemsService.GetItemList(userID)
	if err != nil {
		return upsertBotLastMessage(bot, userID, sourceMsg, bot.Replies.Error, itemBoxMarkup())
	}

	text := bot.Replies.WhatDoWeDelete
	if len(itemList) == 0 {
		text = bot.Replies.ListIsEmpty
	}
	return upsertBotLastMessage(bot, userID, sourceMsg, text, selectItemMarkup(itemList, btnSelectItemToDelete, func(item models.Item) string {
		return item.Name
	}))
}

func itemBoxMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnAddItem, btnEditItem, btnDeleteItem),
		markup.Row(btnCloseItemBox),
	)
	return markup
}

func backToItemBoxMarkup() *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	markup.Inline(markup.Row(btnBackToItemBox))
	return markup
}

func selectItemMarkup(items []models.Item, actionBtn telebot.Btn, data func(models.Item) string) *telebot.ReplyMarkup {
	markup := &telebot.ReplyMarkup{}
	if len(items) > 0 {
		rows := make([]telebot.Row, 0, len(items)+1)
		for _, item := range items {
			btn := markup.Data(item.Name, actionBtn.Unique, data(item))
			rows = append(rows, markup.Row(btn))
		}
		rows = append(rows, markup.Row(btnBackToItemBox))
		markup.Inline(rows...)
		return markup
	}

	markup.Inline(markup.Row(btnBackToItemBox))
	return markup
}

func upsertBotLastMessage(bot *b.Bot, userID int64, sourceMsg *telebot.Message, text string, markup *telebot.ReplyMarkup) error {
	msg := sourceMsg
	if msg == nil {
		msg = bot.ItemsService.GetBotLastMsg(userID)
	}

	if msg != nil {
		edited, err := bot.Edit(msg, text, markup)
		if err == nil {
			if edited != nil {
				bot.ItemsService.SetBotLastMsg(userID, edited)
			} else {
				bot.ItemsService.SetBotLastMsg(userID, msg)
			}
			return nil
		}
	}

	sent := bot.MustSend(userID, text, markup)
	bot.ItemsService.SetBotLastMsg(userID, sent)
	return nil
}

func handleItemInputError(bot *b.Bot, userID int64, err error, isAdd bool) error {
	switch {
	case errors.Is(err, items.ErrItemLimitReached):
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return upsertBotLastMessage(bot, userID, nil, bot.Replies.ItemsLimitReached, itemBoxMarkup())
	case errors.Is(err, items.ErrItemDuplicate):
		return renderInputPrompt(bot, userID, isAdd, bot.Replies.ItemDuplicate)
	case errors.Is(err, items.ErrItemNameEmpty):
		return renderInputPrompt(bot, userID, isAdd, bot.Replies.ItemNameEmpty)
	case errors.Is(err, items.ErrItemNameTooLong):
		return renderInputPrompt(bot, userID, isAdd, bot.Replies.ItemNameTooLong)
	case errors.Is(err, items.ErrItemNotFound):
		bot.Fsm.UserEvent(context.Background(), userID, fsmManager.ItemsMenuOpenedEvent)
		bot.ItemsService.ClearEditingItemName(userID)
		return renderItemBox(bot, userID, nil)
	default:
		bot.ItemsService.ClearEditingItemName(userID)
		return upsertBotLastMessage(bot, userID, nil, bot.Replies.Error, itemBoxMarkup())
	}
}

func renderInputPrompt(bot *b.Bot, userID int64, isAdd bool, note string) error {
	if isAdd {
		return renderAddItemPrompt(bot, userID, nil, note)
	}

	itemID := bot.ItemsService.GetEditingItemName(userID)
	return renderEditItemPrompt(bot, userID, nil, itemID, note)
}

func parseItemName(ctx telebot.Context) (string, error) {
	raw := ctx.Data()
	if raw == "" && ctx.Callback() != nil {
		raw = ctx.Callback().Data
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty item name")
	}
	return raw, nil
}

func respondSilently(ctx telebot.Context) {
	_ = ctx.Respond()
}

func clearClosedItemBoxMessage(bot *b.Bot, userID int64) {
	userDTO := bot.UserService.GetUser(userID)
	if userDTO == nil || userDTO.ItemBoxClosedMsgID == 0 {
		return
	}
	bot.MustDelete(&telebot.Message{
		ID:   userDTO.ItemBoxClosedMsgID,
		Chat: &telebot.Chat{ID: userID},
	})
	if err := bot.UserService.UpdateItemBoxClosedMsgID(userID, 0); err != nil {
		bot.Logger.Error(fmt.Sprintf("Error clearing closed item box message for userID=%d: %v", userID, err))
	}
}

func saveClosedItemBoxMessage(bot *b.Bot, userID int64, msg *telebot.Message) {
	if msg == nil {
		return
	}
	if err := bot.UserService.UpdateItemBoxClosedMsgID(userID, msg.ID); err != nil {
		bot.Logger.Error(fmt.Sprintf("Error saving closed item box message for userID=%d: %v", userID, err))
	}
}
