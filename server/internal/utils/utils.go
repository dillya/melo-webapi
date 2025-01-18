package utils

import (
	"strconv"
	"strings"
	"database/sql"
)

func Uint64FromHwAddress(address string) uint64 {
	value := uint64(0)
	parts := strings.Split(address, ":")
	for i, part := range parts {
		v, err := strconv.ParseUint(part, 16, 8)
		if err == nil {
			value |= v << (8 * (5 - i))
		}
	}
	return value
}

func Uint64ToHwAddress(address uint64) string {
	value := ""
	for i := 5; i >= 0; i-- {
		v := (address >> (8 * i)) & 0xff
		value += strconv.FormatUint(v, 16) + ":"
	}
	return value[:len(value)-1]
}

func InitializeVersionTable(db *sql.DB) error {
	version := `CREATE TABLE IF NOT EXISTS version (
  name VARCHAR(32) NOT NULL,
  version SMALLINT(5) UNSIGNED NOT NULL,
  PRIMARY KEY (name)
);`
	_, err := db.Exec(version)
	return err;
}

func GetTableVersion(db *sql.DB, name string) uint {
	row := db.QueryRow("SELECT version FROM version WHERE name = ?", name)
	if row == nil {
		return 0
	}

	var version uint
	if err := row.Scan(&version); err != nil {
		return 0
	}

	return version
}

func UpdateTableVersion(db *sql.DB, name string, version uint) bool {
	_, err := db.Exec("INSERT INTO version (name, version) VALUES(?, ?) ON DUPLICATE KEY UPDATE name=?, version=?", name, version, name, version)
	return err == nil
}

