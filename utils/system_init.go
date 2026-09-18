package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

func InitConfig() {
	viper.SetConfigName("app")    // 配置文件名称（不含扩展名）
	viper.AddConfigPath("config") // 在当前目录中查找配置文件
	err := viper.ReadInConfig()   // 查找并读取配置文件
	if err != nil {               // 处理读取配置文件时的错误
		panic("Fatal error config file: " + err.Error())
	}
	fmt.Println("config appp init ...")
}

func InitMysql() {
	// 在此处使用 viper.Get("mysql") 获取配置并初始化 MySQL 连接
	//自定义日志模板，打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // 级别
			Colorful:      true,        //彩色
		},
	)
	var err error
	DB, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.database"),
		viper.GetString("mysql.charset"),
		viper.GetBool("mysql.parseTime"),
		viper.GetString("mysql.loc"),
	)), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("Fatal error database connection: " + err.Error())
	}
	fmt.Println("mysql inited ...")
}

func InitRedis(ctx context.Context) {
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdeConn"),
	})

	pong ,err :=Red.Ping(ctx).Result()
	if err != nil{
		fmt.Println("init redis ...",err)
	}else{
		fmt.Println("Redis inited ...",pong)
	}
}

const (
	PublishKey = "websocket"
)


//Publish 发布消息到Redis
func Publish(ctx context.Context,channel string,msg string) error{
	var err error
	fmt.Println("Publish ...",msg)
	err = Red.Publish(ctx,channel,msg).Err()
	if err!=nil{
		fmt.Println(err)
	}
	return err
}

//Subscribe 订阅Redis消息
func Subscribe(ctx context.Context,channel string) (string,error){
	var err error
	fmt.Println("Subs123 ....",ctx)
	sub :=Red.Subscribe(ctx,channel)
	msg,err := sub.ReceiveMessage(ctx)
	if err !=nil{
		fmt.Println("Subscribe error:", err)
		return "",err
	}
	fmt.Println("Subscribe ...",msg.Payload)
	return msg.Payload,err
}