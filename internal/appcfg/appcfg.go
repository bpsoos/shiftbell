package appcfg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type AppCfg struct {
	DatabaseFilepath string
}

func Load(ctx context.Context) (*AppCfg, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	databaseFilepath := os.Getenv("DATABASE_FILEPATH")
	if databaseFilepath == "" {
		return nil, fmt.Errorf("DATABASE_FILEPATH is required")
	}
	if err := validateDatabaseFilepath(databaseFilepath); err != nil {
		return nil, err
	}

	return &AppCfg{DatabaseFilepath: databaseFilepath}, nil
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
