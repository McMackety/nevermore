package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// DefaultConfig is the config loaded at the start of the program
var DefaultConfig Config

// Config is the struct defining the config's JSON structure.
type Config struct {
	WebSocketListenAddress string `json:"websocketListenAddress"`
	Database DatabaseConfig `json:"database"`
}

// DatabaseConfig is the struct defining the database in the
type DatabaseConfig struct {
	Type string `json:"type"`
	Address string `json:"address"`
}


func LoadConfig() {
	// Load the jsonFile from disk
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Panic("Couldn't load config.json, look at https://github.com/McMackety/nevermore/wiki for a guide to properly setup.")
	}
	// Successfully loaded it
	log.Println("Found config.json!")

	defer configFile.Close()

	bytes, _ := ioutil.ReadAll(configFile)

	// Unmarshal the configFile to DefaultConfig, this will be referenced throughout the application
	err = json.Unmarshal(bytes, &DefaultConfig)
	if err != nil {
		log.Panic("Couldn't decipher config.json, check if it is valid JSON.")
	}

	log.Println("Successfully loaded config.json!")
}