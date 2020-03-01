package main

import (
	"github.com/JojiiOfficial/configService"
	"os"

	log "github.com/sirupsen/logrus"
)

const defaultConfigFile = "./data/config.yaml"

//Config config for the server
type Config struct {
	Database          dbConfig
	WebserverConfig   webserverConfig
	DeleteLogInterval int `json:"deleteLogsAfter"`
}

type dbConfig struct {
	Host         string `required:"true"`
	Username     string `required:"true"`
	Pass         string `required:"true"`
	DatabasePort int    `required:"true"`
}

type webserverConfig struct {
	CertFile string `json:"cert"`
	KeyFile  string `json:"key"`
	HTTPPort int    `json:"port"`
	TLSPort  int    `json:"porttls"`
}

//InitConfig init the config file
func InitConfig(configFile string) (*Config, bool) {
	if len(configFile) == 0 {
		configFile = defaultConfigFile
	}

	var config Config
	var createMode bool
	_, err := os.Stat(configFile)
	if err != nil {
		createMode = true

		config = Config{
			Database: dbConfig{
				Host:         "localhost",
				DatabasePort: 3306,
				Username:     "gologger",
				Pass:         "gologger",
			},
			WebserverConfig: webserverConfig{
				HTTPPort: 80,
			},
		}
	}

	isDefault, err := configService.SetupConfig(&config, configFile, configService.NoChange)
	if err != nil {
		log.Fatalln(err.Error())
		return nil, true
	}
	if isDefault {
		log.Println("New config created.")
		if createMode {
			log.Println("Exiting")
			return nil, true
		}
	}

	if err = configService.Load(&config, configFile); err != nil {
		log.Fatalln(err.Error())
		return nil, true
	}

	return &config, false
}
