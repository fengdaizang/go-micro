package pg

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // inject pg driver to go sql
)

func NewPostgresDB(source string) (*gorm.DB, error) {
	gdb, err := gorm.Open("postgres", source)
	if err != nil {
		return nil, err
	}

	gdb.DB().SetMaxIdleConns(3)
	return gdb, nil
}
