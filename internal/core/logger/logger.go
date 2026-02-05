package logger

type AppLogger interface {
	Debug(message string)
	Info(message string)
	Error(message string)
}
