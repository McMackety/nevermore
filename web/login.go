package web

import (
	"encoding/json"
	"github.com/McMackety/nevermore/database"
	"github.com/McMackety/nevermore/webevents"
	"time"
)

type LoginMessage struct {
	Username string `json:"username"`
	Pin      string `json:"pin"`
}

func login(client *webevents.Client, data json.RawMessage) {
	loginAttempts := client.Get("loginAttempts").(int)
	lastLoginAttemptTime := client.Get("lastLoginAttemptTime").(time.Time)
	if loginAttempts > 3 && time.Since(lastLoginAttemptTime).Seconds() > 9 {
		loginAttempts = 0
		client.Set("loginAttempts", loginAttempts)
	}
	if loginAttempts > 3 {
		client.EmitJSON("error", "You are timed out, please wait for 10 seconds.")
		return
	}
	var message LoginMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		client.EmitJSON("error", "Invalid Message! "+err.Error())
		return
	}
	correct := database.CheckPIN(message.Username, message.Pin)
	if correct {
		user := database.GetByUsername(message.Username)
		client.Set("loggedIn", true)
		client.Set("user", user)
		client.Set("userType", user.UserType)
		client.EmitJSON("loggedIn", true)
	} else {
		client.EmitJSON("loggedIn", false)
		if loginAttempts == 3 {
			client.Set("lastLoginAttemptTime", time.Now())
			client.EmitJSON("error", "You are timed out, please wait for 10 seconds.")
		}
		loginAttempts++
		client.Set("loginAttempts", loginAttempts)
	}
}

func getAllUsers(client *webevents.Client, data json.RawMessage) {
	if client.Get("userType") != database.ADMIN {
		client.EmitJSON("tooLowUserType", true)
		return
	}
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	client.EmitJSON("getAllUsersReply", database.GetAllUsers())
}

type UserByIDMessage struct {
	Id uint `json:"id"`
}

func getUserByID(client *webevents.Client, data json.RawMessage) {
	if client.Get("userType") != database.ADMIN {
		client.EmitJSON("tooLowUserType", true)
		return
	}
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	var message UserByIDMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		client.EmitJSON("error", "Invalid Message! "+err.Error())
		return
	}
	client.EmitJSON("getUserbyIDReply", database.GetByID(message.Id))
}

type UserByUsernameMessage struct {
	Username string `json:"username"`
}

func getUserByUsername(client *webevents.Client, data json.RawMessage) {
	if client.Get("userType") != database.ADMIN {
		client.EmitJSON("tooLowUserType", true)
		return
	}
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	var message UserByUsernameMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		client.EmitJSON("error", "Invalid Message! "+err.Error())
		return
	}
	client.EmitJSON("getUserbyUsernameReply", database.GetByUsername(message.Username))
}

type CreateUserMessage struct {
	Username string            `json:"username"`
	UserType database.UserType `json:"userType"`
	Pin      string            `json:"pin"`
}

func createUser(client *webevents.Client, data json.RawMessage) {
	if client.Get("userType") != database.ADMIN {
		client.EmitJSON("tooLowUserType", true)
		return
	}
	if client.Get("loggedIn") != true {
		client.EmitJSON("notAuthenticated", true)
		return
	}
	var message CreateUserMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		client.EmitJSON("error", "Invalid Message! "+err.Error())
		return
	}
	database.Create(message.Username, message.UserType, message.Pin)
}
