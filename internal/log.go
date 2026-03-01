package sisr_remote_helper

import (
	"log/slog"
	"os"
)

func PrepareLogger() (*os.File, error) {
	logFile, err := os.OpenFile("sisr-remote-helper.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	// logFile := os.Stdout

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	return logFile, nil
}
