package mysql

import (
	_ "github.com/go-sql-driver/mysql" // inject mysql driver to go sql
	"github.com/jinzhu/gorm"
)

func NewMySqlDB(source string) (*gorm.DB, error) {
	gdb, err := gorm.Open("mysql", source)
	if err != nil {
		return nil, err
	}

	gdb.DB().SetMaxIdleConns(3)

	return gdb, nil
}
