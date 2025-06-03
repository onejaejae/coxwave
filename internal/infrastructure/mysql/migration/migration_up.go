package main

import (
	"log"

	"github.com/coxwave/coupon-system/internal/infrastructure/mysql"
	"github.com/coxwave/coupon-system/internal/infrastructure/mysql/model"
)

func main() {
	db, err := mysql.NewMySQL()
	if err != nil {
		panic(err)
	}

	err = db.Migrator().DropTable(&model.Campaign{})
	err = db.Migrator().DropTable(&model.Coupon{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&model.Campaign{})
	err = db.AutoMigrate(&model.Coupon{})
	if err != nil {
		panic(err)
	}

	log.Println("Migration up successfully")
}
