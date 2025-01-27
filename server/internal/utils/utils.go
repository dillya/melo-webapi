package utils

import (
	"database/sql"
	"net"
)

func Uint64FromHwAddress(address string) uint64 {
	hw_addr, err := net.ParseMAC(address)
	if err != nil {
		return 0
	}
	return uint64(hw_addr[0])<<40 | uint64(hw_addr[1])<<32 | uint64(hw_addr[2])<<24 |
		uint64(hw_addr[3])<<16 | uint64(hw_addr[4])<<8 | uint64(hw_addr[5])
}

func Uint64ToHwAddress(address uint64) string {
	hw_addr := net.HardwareAddr{
		byte(address >> 40),
		byte(address >> 32),
		byte(address >> 24),
		byte(address >> 16),
		byte(address >> 8),
		byte(address),
	}
	return hw_addr.String()
}

func InitializeVersionTable(db *sql.DB) error {
	version := `CREATE TABLE IF NOT EXISTS version (
  name VARCHAR(32) NOT NULL,
  version SMALLINT(5) UNSIGNED NOT NULL,
  PRIMARY KEY (name)
);`
	_, err := db.Exec(version)
	return err
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
