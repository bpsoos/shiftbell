package appcfg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bpsoos/shiftbell/internal/logging"
)

type AppCfg struct {
	DatabaseFilepath string
	LogHandler       logging.Handler
	AppTimezone      *time.Location
}

func Load() (*AppCfg, error) {
	dbFilepath, err := loadDatabaseFilepath()
	if err != nil {
		return nil, err
	}
	logHandler, err := loadLogHandler()
	if err != nil {
		return nil, err
	}
	appTimezone, err := loadAppTimezone()
	if err != nil {
		return nil, err
	}

	return &AppCfg{
		DatabaseFilepath: dbFilepath,
		LogHandler:       logHandler,
		AppTimezone:      appTimezone,
	}, nil
}

func loadDatabaseFilepath() (string, error) {
	databaseFilepath := os.Getenv("DATABASE_FILEPATH")
	if databaseFilepath == "" {
		return "", errors.New("DATABASE_FILEPATH is required")
	}
	if err := validateDatabaseFilepath(databaseFilepath); err != nil {
		return "", err
	}

	return databaseFilepath, nil
}

func loadLogHandler() (logging.Handler, error) {
	logFormat := os.Getenv("LOG_FORMAT")
	switch logFormat {
	case "", "json":
		return logging.HandlerJSON, nil
	case "text":
		return logging.HandlerConsole, nil
	default:
		return logging.HandlerJSON, fmt.Errorf(
			"invalid LOG_FORMAT %q: expected text or json",
			logFormat,
		)
	}
}

func loadAppTimezone() (*time.Location, error) {
	appTimezone := os.Getenv("APP_TIMEZONE")
	if appTimezone == "" {
		return time.UTC, nil
	}

	location, err := time.LoadLocation(appTimezone)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_TIMEZONE %q: %w", appTimezone, err)
	}

	return location, nil
}

func validateDatabaseFilepath(databaseFilepath string) error {
	parentDirectory := filepath.Dir(databaseFilepath)
	parentInfo, err := os.Stat(parentDirectory)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_FILEPATH parent directory: %w", err)
	}
	if !parentInfo.IsDir() {
		return errors.New("invalid DATABASE_FILEPATH parent directory is not a directory")
	}

	info, err := os.Stat(databaseFilepath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("invalid DATABASE_FILEPATH: %w", err)
	}
	if info.IsDir() {
		return errors.New("invalid DATABASE_FILEPATH points to a directory")
	}

	return nil
}
