package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(debug bool) {
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
	}

	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	Log = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}

func Debug() *zerolog.Event {
	return Log.Debug()
}

func Info() *zerolog.Event {
	return Log.Info()
}

func Warn() *zerolog.Event {
	return Log.Warn()
}

func Error() *zerolog.Event {
	return Log.Error()
}

func Fatal() *zerolog.Event {
	return Log.Fatal()
}
