package models

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/code-sleuth/vending-machine/config"
	"github.com/jinzhu/gorm"

	// Required for pq to work
	_ "github.com/lib/pq"
)

func dbConnect(environment string) *gorm.DB {
	dbConfig := config.GetConfig()

	roleTypeSQLStatement := `
	DO $$ BEGIN
    	CREATE TYPE user_role AS ENUM ('buyer','seller');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`

	// variable to hold name of database
	// for either dev, testing or prod
	databaseName := dbConfig.DB.DBName

	// use test db when in test environment
	if strings.Contains(environment, "test") {
		databaseName = dbConfig.DB.TestDBName
	}

	db, err := gorm.Open(dbConfig.DB.Dialect, fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		dbConfig.DB.Host, dbConfig.DB.Port, dbConfig.DB.Username, databaseName, dbConfig.DB.Password, dbConfig.DB.SSLMode))
	if err != nil {
		log.Fatal(err.Error())
	}

	db.Exec(roleTypeSQLStatement)
	return db
}

// Adapter handles the initial migrations
func Adapter() {
	db := dbConnect(os.Getenv("ENVIRONMENT"))
	db.AutoMigrate(
		User{},
	)
	db.Model(&User{}).AddUniqueIndex("idx_email", "email")
	defer func() {
		if err := db.Close(); err != nil {
			log.Println(err)
		}
	}()
	return
}
