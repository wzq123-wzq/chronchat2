package main

import (
	"context"
	"fmt"

	"chronchat/models"
	router "chronchat/router"
	"chronchat/utils"
)

func main() {
	ctx := context.Background()
	utils.InitConfig()
	utils.InitMysql()
	utils.InitRedis(ctx)
	//自动建表/补列 幂等
	if err := utils.DB.AutoMigrate(
		&models.UserBasic{},
		&models.Message{},
		&models.Contact{},
		&models.GroupBasic{},
	); err != nil {
		panic("Fatal error database migrate: " + err.Error())
	}
	fmt.Println("mysql migrate inited ...")
	r := router.Router()
	r.Run(":8081")
}
