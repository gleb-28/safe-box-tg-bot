package text

type Replies struct {
	Start             string
	EnterKey          string
	EnterKeySuccess   string
	KeyAlreadyEntered string
	EnterKeyWrong     string
	Error             string

	ItemBoxClosed string

	ChangeModePrompt           string
	ChangeModeUpdated          string
	ChangeIntervalPrompt       string
	ChangeIntervalUpdated      string
	ToggleNotificationsPrompt  string
	ToggleNotificationsUpdated string
	ChangeDayStartPrompt       string
	ChangeDayEndPrompt         string
	ChangeDayUpdated           string

	AddNewItem          string
	WriteNewItemName    string
	NewNameForValue     string
	WhatDoWeEdit        string
	WhatDoWeDelete      string
	ListIsEmpty         string
	ItemsMenuEmpty      string
	ItemsMenuHeader     string
	ItemsMenuStatus     string
	ItemsMenuFooter     string
	ItemsMenuItemPrefix string
	ItemsLimitReached   string
	ItemDuplicate       string
	ItemNameEmpty       string
	ItemNameTooLong     string
	ItemsErrEmptyID     string
	ItemsErrInvalidID   string
	ItemsErrEmptyName   string

	OpenReminderBox         string
	ReminderBoxClosed       string
	RemindersMenuEmpty      string
	RemindersMenuHeader     string
	RemindersMenuFooter     string
	RemindersMenuItemRow    string
	RemindersLimitReached   string
	ReminderNamePrompt      string
	ReminderNameEmpty       string
	ReminderNameTooLong     string
	ReminderDuplicate       string
	ReminderSelectTypeFirst string
	ReminderIntervalPrompt  string
	ReminderIntervalInvalid string
	ReminderWeekdayPrompt   string
	ReminderWeekdayInvalid  string
	ReminderMonthDayPrompt  string
	ReminderTimePrompt      string
	ReminderTimeFormatError string
	ReminderScheduleInvalid string
	ReminderSchedulePrompt  string
	ReminderOnceDatePrompt  string
	ReminderOnceDateInvalid string
	ReminderOnceDatePast    string
	ReminderOnceTimePast    string
	ReminderMonthDayInvalid string
	ReminderHumanOnce       string
	ReminderHumanInterval   string
	ReminderHumanDaily      string
	ReminderHumanWeekly     string
	ReminderHumanMonthly    string
	ReminderHumanFallback   string
}

func NewReplies() *Replies {
	return &Replies{
		Start:             "Привет! Я бот «Безопасная шкатулка» ✨\nДобавляй сюда свои любимые вещи, а я буду иногда напоминать тебе про них, чтобы помогать возвращаться в безопасный режим 🙂",
		EnterKey:          "Введи секретный ключ:",
		EnterKeySuccess:   "Супер! Теперь можешь пользоваться!",
		KeyAlreadyEntered: "Ты уже активировал секретный ключ",
		EnterKeyWrong:     "Неверный ключ",
		Error:             "Произошла ошибка :(",

		ItemBoxClosed: "Шкатулка закрыта 🔒",

		ChangeModePrompt:           "Выбери режим (сейчас: \"%s\")",
		ChangeModeUpdated:          "Режим переключён на \"%s\" ✅",
		ChangeIntervalPrompt:       "Выбери частоту напоминаний (сейчас: \"%s\")",
		ChangeIntervalUpdated:      "Частота переключена на \"%s\" (%s) ✅",
		ToggleNotificationsPrompt:  "Уведомления сейчас %s. Переключить?",
		ToggleNotificationsUpdated: "Уведомления %s ✅",
		ChangeDayStartPrompt:       "🕒 Когда можно писать?\nТекущий интервал: %s–%s\n\nВыбери начало дня:",
		ChangeDayEndPrompt:         "Выбери конец дня (начало: %s):",
		ChangeDayUpdated:           "Готово ✨\nЯ буду писать с %s до %s",

		AddNewItem:          "✍️ Напиши новую вещь 👇",
		WriteNewItemName:    "✏️ Напиши новое имя 👇",
		NewNameForValue:     "✏️ Новое имя для \"%s\" 👇",
		WhatDoWeEdit:        "Что изменить?",
		WhatDoWeDelete:      "Что удалить?",
		ListIsEmpty:         "Список пуст",
		ItemsMenuEmpty:      "%s\n📦 Твои вещи\n\n(пока пусто)\n\nЧто делаем?",
		ItemsMenuHeader:     "%s\n📦 Твои вещи:\n\n",
		ItemsMenuStatus:     "Режим: <b>%s</b> • Вещей: <b>%d</b> • Окно: <b>%s–%s</b>\n",
		ItemsMenuFooter:     "\nЧто делаем?",
		ItemsMenuItemPrefix: "• ",
		ItemsLimitReached:   "Достигнут лимит вещей. Удали что-то и попробуй снова",
		ItemDuplicate:       "Такая вещь уже есть. Напиши другую",
		ItemNameEmpty:       "Пустое название. Напиши ещё раз",
		ItemNameTooLong:     "Слишком длинно. Сократи название",

		OpenReminderBox:         "Открыть напоминания 🔔",
		ReminderBoxClosed:       "Напоминания закрыты 🔒",
		RemindersMenuEmpty:      "%s\n🔔 Твои напоминания\n\n(пока пусто)\n\nЧто делаем?",
		RemindersMenuHeader:     "%s\n🔔 Твои напоминания:\n\n",
		RemindersMenuFooter:     "\nЧто делаем?",
		RemindersMenuItemRow:    "• %s — %s\n",
		RemindersLimitReached:   "Слишком много напоминаний. Удали что-то и попробуй снова",
		ReminderNamePrompt:      "✍️ Назови напоминание",
		ReminderNameEmpty:       "Пустое название. Напиши ещё раз",
		ReminderNameTooLong:     "Слишком длинно. Сократи название",
		ReminderDuplicate:       "Такое напоминание уже есть. Введи другое название",
		ReminderSelectTypeFirst: "Сначала выбери тип напоминания",
		ReminderIntervalPrompt:  "⏱ Напиши интервал в минутах",
		ReminderIntervalInvalid: "Напиши положительное число минут",
		ReminderWeekdayPrompt:   "📅 Выбери день недели",
		ReminderWeekdayInvalid:  "Выбери день недели",
		ReminderMonthDayPrompt:  "📅 Напиши число месяца (1–31)\nЕсли нужен последний день месяца — введи 31",
		ReminderMonthDayInvalid: "День должен быть 1–31",
		ReminderTimePrompt:      "⌚️ Напиши время в формате HH:MM",
		ReminderTimeFormatError: "Формат HH:MM",
		ReminderScheduleInvalid: "Неверное расписание. Попробуй снова",
		ReminderSchedulePrompt:  "Как часто напоминать?",
		ReminderOnceDatePrompt:  "📅 Напиши дату в формате ДД.ММ",
		ReminderOnceDateInvalid: "Дата не подходит. Формат ДД.ММ",
		ReminderOnceDatePast:    "Дата уже в прошлом. Введи будущую",
		ReminderOnceTimePast:    "Время уже прошло для выбранной даты",
		ReminderHumanOnce:       "разово %s",
		ReminderHumanInterval:   "каждые %d мин",
		ReminderHumanDaily:      "ежедневно в %s",
		ReminderHumanWeekly:     "по %s в %s",
		ReminderHumanMonthly:    "каждый %d день в %s",
		ReminderHumanFallback:   "по расписанию",
	}
}
