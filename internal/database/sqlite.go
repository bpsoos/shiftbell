package database

import "net/url"

func SQLiteDSN(databaseFilepath string) string {
	return sqliteURL("file", databaseFilepath)
}

func SQLiteMigrationURL(databaseFilepath string) string {
	return sqliteURL("sqlite", databaseFilepath)
}

func sqliteURL(scheme string, databaseFilepath string) string {
	dsn := url.URL{Scheme: scheme, Path: databaseFilepath}
	query := dsn.Query()
	query.Set("_pragma", "foreign_keys(1)")
	dsn.RawQuery = query.Encode()

	return dsn.String()
}
