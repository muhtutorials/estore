package database

import (
	"estore/src/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	dsn := "igor:secret@tcp(db)/estore?charset=utf8&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Couldn't connect to database")
	}
}

func AutoMigrate() {
	DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Link{},
		&models.Order{},
		&models.OrderItem{},
	)
}
