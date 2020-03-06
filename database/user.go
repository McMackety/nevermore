package database

import "github.com/jinzhu/gorm"

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

func GetByID(id uint) User {
	var userFromDatabase User
	Database.Where("id = ?", id).Select("id, createdAt, updatedAt, username, userType").First(&userFromDatabase)
	return userFromDatabase
}

func GetByUsername(username string) User {
	var userFromDatabase User
	Database.Where(&User{Username: username}).Select("id, createdAt, updatedAt, username, userType").First(&userFromDatabase)
	return userFromDatabase
}

func CheckPIN(username string, pin string) bool {
	var userFromDatabase User
	Database.Where(&User{Username: username}).First(&userFromDatabase)
	if userFromDatabase.Pin == pin {
		return true
	}
	return false
}

func Create(username string, userType UserType, pin string) {
	Database.Create(&User{Username: username, UserType: userType, Pin: pin})
}

func (user *User) Update(updates map[string]interface{}) {
	Database.Model(user).Updates(updates)
}
