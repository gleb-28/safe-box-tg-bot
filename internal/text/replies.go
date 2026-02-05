package text

type Replies struct {
	Start             string
	EnterKey          string
	EnterKeySuccess   string
	KeyAlreadyEntered string
	EnterKeyWrong     string
	Error             string
}

func NewReplies() *Replies {
	return &Replies{
		"Привет! я бот безопасной шкатулки. Я умею напоминать тебе войти в безопасный режим на основе твоих вещей из шкатулки! Добавляй сюда свои любимые вещи, а я буду периодически напоминать тебе о них!",
		"Введи секретный ключ:",
		"Супер! Теперь можешь пользоваться!",
		"Ты уже активировал секретный ключ",
		"Неверный ключ",
		"Произошла ошибка :(",
	}
}
