package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"chronchat/utils"
)

// UserBasic 用户表
// 嵌入 gorm.Model 自带字段: ID(主键), CreatedAt(创建时间), UpdatedAt(更新时间), DeletedAt(软删除时间)
type UserBasic struct {
	gorm.Model
	Name          string     // 用户名,登录用
	Password      string     // 密码,已加盐加密存储
	Avatar        string     // 头像URL地址
	Phone         string     // 手机号,正则校验 1开头11位
	Email         string     // 邮箱地址,格式校验
	ClientIp      string     // 最近一次登录的客户端IP
	ClientPort    string     // 最近一次登录的客户端端口
	Salt          string     // 密码加密盐值,注册时生成
	Identity      string     // 登录唯一标识,即会话token,登录成功时生成
	LoginTime     *time.Time // 最近一次登录时间
	HeartbeatTime *time.Time // 最近一次心跳时间,用于判断在线状态
	LoginOutTime  *time.Time // 最近一次主动退出登录时间
	IsLogout      bool       // 是否已被注销/冻结,true表示账户不可用
	DeviceInfo    string     // 最近一次登录的设备信息(浏览器/移动端标识)
}

func FindUserByNameAndPwd(name, password string) (user UserBasic, err error) {
	user = UserBasic{}
	err = utils.DB.Where("name = ? AND password = ?", name, password).First(&user).Error

	//token加密
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.Md5Encode(str)
	utils.DB.Model(&user).Where("id = ?", user.ID).Update("identity", temp)
	user.Identity = temp
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

// UpdateUserLoginTime 更新用户登录时间
func UpdateUserLoginTime(userId uint) {
	now := time.Now()
	utils.DB.Model(&UserBasic{}).Where("id = ?", userId).Update("login_time", &now)
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

// UpdateUser 更新用户信息
func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(UserBasic{
		Name:     user.Name,
		Password: user.Password,
		Avatar:   user.Avatar,
		Phone:    user.Phone,
		Email:    user.Email,
	})
}
