package main

import (
	"chronchat/models"
	_"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:3306)/chronchat?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//迁移 schema
	db.AutoMigrate(&models.UserBasic{})
	db.AutoMigrate(&models.Message{})
	db.AutoMigrate(&models.Contact{})
	db.AutoMigrate(&models.GroupBasic{})

	// //create
	// user := &models.UserBasic{}
	// user.Name = "test"
	// result := db.Create(user)
	// if result.Error != nil {
	// 	panic("failed to create user")
	// }

	// //read
	// fmt.Println(db.First(user, 1))
	// db.Model(user).Update("PassWord", "1234")
}
