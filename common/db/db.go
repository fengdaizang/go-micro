package db

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"tanghu.com/go-micro/common/db/mysql"
	"tanghu.com/go-micro/common/db/pg"
)

var (
	gdb      *gorm.DB
	initOnce sync.Once
)

// GetDB gets the gorm.DB instance which is safe for concurrent use by multiple goroutines.
func GetDB() *gorm.DB {
	initOnce.Do(func() {
		var err error
		gdb, err = newDB()
		if err != nil {
			panic(err)
		}
	})
	if gdb == nil {
		panic("gorm.DB is nil")
	}

	return gdb
}

func newDB() (db *gorm.DB, err error) {
	defaultDB := viper.GetString("database.default")
	if defaultDB == "" {
		defaultDB = "mysql"
	}

	connectionKey := fmt.Sprintf("database.%s.connection", defaultDB)
	connection := viper.GetString(connectionKey)
	if connection == "" {
		return nil, fmt.Errorf("Invalid database connection addresses ")
	}

	switch defaultDB {
	case "postgres":
		db, err = pg.NewPostgresDB(connection)
	case "mysql":
		db, err = mysql.NewMySqlDB(connection)
	default:
		return nil, fmt.Errorf("not support database type %s ", defaultDB)
	}

	return db, err
}
