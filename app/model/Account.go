package model

import (
	"github.com/open_tool/app/common"
	"gorm.io/gorm"
)

type Account struct {
	Id           int64  `json:"id"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (c *Account) TableName() string {
	return "account"
}

func (c *Account) GetDB() *gorm.DB {
	return common.GetDB()
}

// Insert 插入
func (c *Account) Insert() error {
	db := c.GetDB()

	// 判断是否存在
	err := db.Table(c.TableName()).Where("email = ?", c.Email).First(c).Error
	if err == nil {
		// 已存在 就更新
		return c.update()
	}
	err = db.Table(c.TableName()).Create(c).Error
	if err != nil {
		return err
	}
	return nil
}

// Update 更新
func (c *Account) update() error {
	db := c.GetDB()

	err := db.Table(c.TableName()).Where("email = ?", c.Email).Updates(c).Error
	if err != nil {
		return err
	}
	return nil
}

func (c *Account) GetAccountByEmail(email string) error {
	db := c.GetDB()

	err := db.Table(c.TableName()).Where("email = ?", email).First(c).Error
	if err != nil {
		return err
	}
	return nil
}
