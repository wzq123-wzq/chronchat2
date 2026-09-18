package main

import (
	router "chronchat/router"
	"chronchat/utils"
	"context"
)

func main() {
	ctx := context.Background()
	utils.InitConfig()
	utils.InitMysql()
	utils.InitRedis(ctx)
	r := router.Router()
	r.Run(":8081")
}
