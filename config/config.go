package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Google struct {
		ClientID     string `yaml:"ClientID"`
		ClientSecret string `yaml:"ClientSecret"`
		Issuer       string `yaml:"ISSUER"`
		Redirect_url string `yaml:"REDIRECT_URL"`
	} `yaml:"google"`

	Database struct {
		HOST         string `yaml:"HOST"`
		PORT         string `yaml:"PORT"`
		DATABASENAME string `yaml:"DATABASENAME"`
		USER_NAME    string `yaml:"USER_NAME"`
		PASSWORD     string `yaml:"PASSWORD"`
	} `yaml:"database"`
}

var AppConfig Config

func LoadConfig() {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(file, &AppConfig)
	if err != nil {
		log.Fatal("Error parsing config file: %v", err)
	}
	// fmt.Print(file)
}

// type Google_request struct {
// 	ClientID      string
// 	ClientSecrete string
// 	Issuer        string
// 	Redirect_url  string
// }

// func ConfigGoogle() *Google_request {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	client_id := os.Getenv("ClientID")
// 	client_secrete := os.Getenv("ClientSecrete")
// 	issuer := os.Getenv("ISSUER")
// 	redirect_url := os.Getenv("REDIRECT_URL")

// 	return &Google_request{
// 		ClientID:      client_id,
// 		ClientSecrete: client_secrete,
// 		Issuer:        issuer,
// 		Redirect_url:  redirect_url,
// 	}
// }
