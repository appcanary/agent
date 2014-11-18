package data

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
)

var lg = logging.MustGetLogger("app-canary")

type Gemfile struct {
	Id        int64
	AppName   string
	Path      string
	Gems      []Gem
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type Gem struct {
	Id        int64
	GemfileId int64
	Name      string
	Version   string
	Source    string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func Initialize() gorm.DB {
	db, err := gorm.Open("sqlite3", "canary.db")
	if err != nil {
		lg.Fatal(err)
	}

	db.AutoMigrate(&Gem{})
	db.AutoMigrate(&Gemfile{})
	return db
}
