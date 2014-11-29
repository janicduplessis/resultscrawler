package logger

// The Logger interface
type Logger interface {
	Log(message string)
	Logf(message string, args ...interface{})
	Error(err error)
	Fatal(err error)
}
