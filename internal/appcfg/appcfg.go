package appcfg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bpsoos/shiftbell/internal/logging"
)

type AppCfg struct {
	DatabaseFilepath string
	LogHandler       logging.Handler
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

	return &AppCfg{
		DatabaseFilepath: dbFilepath,
		LogHandler:       logHandler,
	}, nil
}

func loadDatabaseFilepath() (string, error) {
	databaseFilepath := os.Getenv("DATABASE_FILEPATH")
	if databaseFilepath == "" {
		return "", fmt.Errorf("DATABASE_FILEPATH is required")
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
		return logging.HandlerJSON, fmt.Errorf("invalid LOG_FORMAT %q: expected text or json", logFormat)
	}
}

func validateDatabaseFilepath(databaseFilepath string) error {
	parentDirectory := filepath.Dir(databaseFilepath)
	parentInfo, err := os.Stat(parentDirectory)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_FILEPATH parent directory: %w", err)
	}
	if !parentInfo.IsDir() {
		return fmt.Errorf("invalid DATABASE_FILEPATH parent directory is not a directory")
	}

	info, err := os.Stat(databaseFilepath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("invalid DATABASE_FILEPATH: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid DATABASE_FILEPATH points to a directory")
	}

	return nil
}
