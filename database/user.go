package database

import (
	"errors"
	"github.com/jinzhu/gorm"
)

type UserType int

const (
	ADMIN UserType = iota
	FTA
	HEADREFEREE
	REFEREE
	SCORER
)

type User struct {
	gorm.Model
	Username string
	UserType UserType
	Pin      string
}

func GetAllUsers() []User {
	var users []User
	Database.Select("id, createdAt, updatedAt, username, userType").Find(&users)
	return users
}

func GetUserByID(id uint) (user User, err error) {
	var userFromDatabase User
	if err := Database.Where("id = ?", id).Select("id, createdAt, updatedAt, username, userType").First(&userFromDatabase).Error; err != nil {
		return userFromDatabase, nil
	}
	return userFromDatabase, errors.New("couldn't find user")
}

func GetUserByUsername(username string) (user User, err error) {
	var userFromDatabase User
	if err := Database.First(&userFromDatabase, "username = ?", username).Select("id, createdAt, updatedAt, username, userType").Error; err != nil {
		return userFromDatabase, nil
	}
	return userFromDatabase, errors.New("couldn't find user")
}

func CheckUserPIN(username string, pin string) bool {
	var userFromDatabase User
	if err := Database.Where("username = ?", username).Select("id, createdAt, updatedAt, username, userType").First(&userFromDatabase).Error; err != nil {
		return false
	}
	if userFromDatabase.Pin == pin {
		return true
	}
	return false
}

func CreateUser(username string, userType UserType, pin string) {
	Database.Create(&User{Username: username, UserType: userType, Pin: pin})
}

func (user *User) Update(updates map[string]interface{}) {
	Database.Model(user).Updates(updates)
}
