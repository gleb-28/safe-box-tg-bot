package fsmManager

import (
	"context"
	"fmt"
	"safeboxtgbot/internal/core/logger"
	"sync"

	f "github.com/looplab/fsm"
)

const (
	StateInitial                = "initial"
	StateAwaitingKey            = "awaiting_key"
	StateItemsMenuOpened        = "items_menu_opened"
	StateAwaitingItemAdd        = "awaiting_items_add"
	StateItemEditSelectOpened   = "item_edit_select_opened"
	StateAwaitingItemEdit       = "awaiting_item_edit"
	StateItemDeleteSelectOpened = "item_delete_select_opened"
)

const (
	InitialEvent                = "initial__event"
	AwaitingKeyEvent            = "awaiting_key__event"
	ItemsMenuOpenedEvent        = "items_menu_opened__event"
	AwaitingItemAddEvent        = "awaiting_item_add__event"
	ItemEditSelectOpenedEvent   = "item_edit_select_opened__event"
	AwaitingItemEditEvent       = "awaiting_item_edit__event"
	ItemDeleteSelectOpenedEvent = "item_delete_select_opened__event"
)

var events = []f.EventDesc{
	{
		Name: InitialEvent,
		Src: []string{
			StateInitial,
			StateAwaitingKey,
			StateItemsMenuOpened,
			StateAwaitingItemAdd,
			StateItemEditSelectOpened,
			StateAwaitingItemEdit,
			StateItemDeleteSelectOpened,
		},
		Dst: StateInitial,
	},
	{Name: AwaitingKeyEvent, Src: []string{StateInitial}, Dst: StateAwaitingKey},
	{
		Name: ItemsMenuOpenedEvent,
		Src: []string{
			StateInitial,
			StateItemsMenuOpened,
			StateAwaitingItemAdd,
			StateItemEditSelectOpened,
			StateAwaitingItemEdit,
			StateItemDeleteSelectOpened,
		},
		Dst: StateItemsMenuOpened,
	},
	{Name: AwaitingItemAddEvent, Src: []string{StateItemsMenuOpened}, Dst: StateAwaitingItemAdd},
	{Name: ItemEditSelectOpenedEvent, Src: []string{StateItemsMenuOpened}, Dst: StateItemEditSelectOpened},
	{Name: AwaitingItemEditEvent, Src: []string{StateItemEditSelectOpened}, Dst: StateAwaitingItemEdit},
	{Name: ItemDeleteSelectOpenedEvent, Src: []string{StateItemsMenuOpened}, Dst: StateItemDeleteSelectOpened},
}

type FSMState struct {
	users  map[int64]*f.FSM
	mu     *sync.Mutex
	logger logger.AppLogger
}

func New(logger logger.AppLogger) *FSMState {
	return &FSMState{
		users:  make(map[int64]*f.FSM),
		mu:     &sync.Mutex{},
		logger: logger,
	}
}

func (fsm *FSMState) GetFSMForUser(userID int64) *f.FSM {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	if userFsm, exists := fsm.users[userID]; exists {
		return userFsm
	}

	userFsm := f.NewFSM(
		StateInitial,
		events,
		f.Callbacks{
			"before_event": func(_ context.Context, e *f.Event) { fsm.beforeEvent(userID, e) },
			"after_event":  func(_ context.Context, e *f.Event) { fsm.afterEvent(userID, e) },
		},
	)
	fsm.users[userID] = userFsm
	return userFsm
}

func (fsm *FSMState) beforeEvent(userID int64, e *f.Event) {
	fsm.logger.Debug(fmt.Sprintf("User %d - Before event '%s': State '%s' -> '%s'\n", userID, e.Event, e.Src, e.Dst))
}

func (fsm *FSMState) afterEvent(userID int64, e *f.Event) {
	fsm.logger.Debug(fmt.Sprintf("User %d - After event '%s': New state '%s'\n", userID, e.Event, e.Dst))
}

func (fsm *FSMState) UserEvent(ctx context.Context, chatId int64, event string, args ...interface{}) {
	userFSM := fsm.GetFSMForUser(chatId)
	err := userFSM.Event(ctx, event, args...)
	if err != nil {
		fsm.logger.Error(err.Error())
	}
}
