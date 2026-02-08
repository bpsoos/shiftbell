package main

import (
	"os"

	"github.com/bpsoos/shiftbell/internal/migrations"
)

func main() {
	migrations.NewMigrator().Migrate(os.Getenv("DATABASE_URL"))
}
