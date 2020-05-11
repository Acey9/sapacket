package outputer

type Logger struct{}

func (l *Logger) Output(msg string) {
	logging(msg)
}
