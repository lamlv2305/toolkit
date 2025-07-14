package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/pkgerrors"
)

func Setup(logLevel zerolog.Level) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	zerolog.SetGlobalLevel(logLevel)
}

func New(options ...func(*LogOption)) zerolog.Logger {
	opt := LogOption{}

	for idx := range options {
		options[idx](&opt)
	}

	wr := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages", missed)
	})

	logger := zerolog.New(wr).With().Caller().Timestamp().Logger()

	if opt.Pretty {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC1123})
	}

	return logger
}
