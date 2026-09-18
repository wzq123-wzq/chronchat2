package models

import (
	"chronchat/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	Name          string // 用户名
	Password      string // 密码
	Phone         string `valid:"matches(^1[3-9]\\d{9}$)"`
	Email         string `valid:"email"`
	ClientIp      string // 客户端IP
	ClinetPort    string // 客户端端口
	Salt          string // 加密盐
	Identity      string // 唯一标识
	LoginTime     *time.Time // 登录时间
	HeartbeatTime *time.Time // 心跳时间
	LoginOutTime  *time.Time `gorm:"column:login_out_time" json:"login_out_time"`
	IsLogout      bool // 是否注销
	DeviceInfo    string // 设备信息
}

func FindUserByNameAndPwd(name, password string) (user UserBasic, err error) {
	user = UserBasic{}
	err = utils.DB.Where("name = ? AND password = ?", name, password).First(&user).Error

	//token加密
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.Md5Encode(str)
	utils.DB.Model(&user).Where("id = ?",user.ID).Update("identity", temp)
	return user, err
}

func FindUserById(id uint) (user UserBasic, err error) {
	user = UserBasic{}
	err = utils.DB.Where("id = ?", id).First(&user).Error
	return user, err
}

func FindUserByName(name string) (user UserBasic, err error) {
	user = UserBasic{}
	err = utils.DB.Where("name = ?", name).First(&user).Error
	return user, err
}

func FindUserByPhone(phone string) (user UserBasic, err error) {
	user = UserBasic{}
	err = utils.DB.Where("phone = ?", phone).First(&user).Error
	return user, err
}

func FindUserByEmail(email string) (user UserBasic, err error) {
	user = UserBasic{}
	err = utils.DB.Where("email = ?", email).First(&user).Error
	return user, err
}

func (UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList() ([]*UserBasic, error) {
	users := make([]*UserBasic, 10)
	err := utils.DB.Find(&users).Error
	for _, user := range users {
		fmt.Println(user)
	}
	return users, err
}

func CreateUser(user UserBasic) *gorm.DB {
	return utils.DB.Create(&user)
}

func DeleteUser(user UserBasic) *gorm.DB {
	return utils.DB.Delete(&user)
}

func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name, Password: user.Password, Phone: user.Phone, Email: user.Email})
}
