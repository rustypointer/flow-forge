package logger

import (
	"io"
	"log/slog"
	"os"
)

var Log *slog.Logger

func init() {
	Log = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func Init() {
	Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

