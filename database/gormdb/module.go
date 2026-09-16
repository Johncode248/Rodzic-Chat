package gormdb

import (
	"fmt"
	"time"

	"github.com/Johncode248/Rodzic-Chat/variables"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GormDB *gorm.DB

func init() {

	db, err := gorm.Open(postgres.Open(variables.DatabaseURL), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	} else {
		fmt.Println("Connected with DB")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to set setting for database: %v", err))
	}

	// Maksymalna liczba otwartych połączeń do bazy (zapobiega "too many clients" w Postgresie)
	sqlDB.SetMaxOpenConns(21)

	// Maksymalna liczba bezczynnych połączeń w puli
	sqlDB.SetMaxIdleConns(21)

	sqlDB.SetConnMaxLifetime(time.Hour)

	sqlDB.SetConnMaxIdleTime(2 * time.Hour) // Warto dodać: czyści stare, bezczynne połączenia

	GormDB = db
}
