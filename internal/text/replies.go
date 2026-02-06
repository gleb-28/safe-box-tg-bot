package text

type Replies struct {
	Start             string
	EnterKey          string
	EnterKeySuccess   string
	KeyAlreadyEntered string
	EnterKeyWrong     string
	Error             string

	Done string

	AddNewItem          string
	WriteNewItemName    string
	NewNameForValue     string
	WhatDoWeEdit        string
	WhatDoWeDelete      string
	ListIsEmpty         string
	ItemsMenuEmpty      string
	ItemsMenuHeader     string
	ItemsMenuFooter     string
	ItemsMenuItemPrefix string
	ItemsLimitReached   string
	ItemDuplicate       string
	ItemNameEmpty       string
	ItemNameTooLong     string
	ItemsErrEmptyID     string
	ItemsErrInvalidID   string
	ItemsErrEmptyName   string
}

func NewReplies() *Replies {
	return &Replies{
		Start:             "Привет! Я бот «Безопасная шкатулка» ✨\nДобавляй сюда свои любимые вещи, а я буду иногда напоминать тебе про них, чтобы помогать возвращаться в безопасный режим 🙂",
		EnterKey:          "Введи секретный ключ:",
		EnterKeySuccess:   "Супер! Теперь можешь пользоваться!",
		KeyAlreadyEntered: "Ты уже активировал секретный ключ",
		EnterKeyWrong:     "Неверный ключ",
		Error:             "Произошла ошибка :(",

		Done: "Готово 😎",

		AddNewItem:          "✍️ Напиши новую вещь 👇",
		WriteNewItemName:    "✏️ Напиши новое имя 👇",
		NewNameForValue:     "✏️ Новое имя для \"%s\" 👇",
		WhatDoWeEdit:        "Что изменить?",
		WhatDoWeDelete:      "Что удалить?",
		ListIsEmpty:         "Список пуст",
		ItemsMenuEmpty:      "📦 Твои вещи:\n\n(пока пусто)\n\nЧто делаем?",
		ItemsMenuHeader:     "📦 Твои вещи:\n\n",
		ItemsMenuFooter:     "\nЧто делаем?",
		ItemsMenuItemPrefix: "• ",
		ItemsLimitReached:   "Достигнут лимит вещей. Удали что-то и попробуй снова",
		ItemDuplicate:       "Такая вещь уже есть. Напиши другую",
		ItemNameEmpty:       "Пустое название. Напиши ещё раз",
		ItemNameTooLong:     "Слишком длинно. Сократи название",
	}
}
