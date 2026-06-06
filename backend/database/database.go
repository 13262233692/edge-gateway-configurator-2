package database

import (
	"database/sql"
	"edge-gateway-configurator/models"

	_ "modernc.org/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() error {
	sqlDB, err := sql.Open("sqlite", "gateway_config.db")
	if err != nil {
		return err
	}

	DB, err = gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		return err
	}

	if err := models.Migrate(DB); err != nil {
		return err
	}

	if err := models.SeedData(DB); err != nil {
		return err
	}

	return nil
}
