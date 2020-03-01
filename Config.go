package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

//Config config for the server
type Config struct {
	Host            string `json:"host"`
	Username        string `json:"username"`
	Pass            string `json:"pass"`
	DatabasePort    int    `json:"dbport"`
	CertFile        string `json:"cert"`
	KeyFile         string `json:"key"`
	ShowTimeInLog   bool   `json:"showLogTime"`
	HTTPPort        int    `json:"port"`
	TLSPort         int    `json:"porttls"`
	DeleteLogsAfter int    `json:"deleteLogsAfter"`
}

func createConfig(configFile string) error {
	log.Error("Couldn't find " + configFile)
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}

	emptyConfig := Config{}
	d, err := json.Marshal(emptyConfig)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, d, "", "\t")

	_, err = f.Write(out.Bytes())
	if err != nil {
		return err
	}

	if f != nil {
		f.Close()
	}

	return nil
}

func readConfig(file string) *Config {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	res := &Config{}
	err = json.Unmarshal(dat, &res)
	if err != nil {
		panic(err)
	}
	return res
}

func checkConfig(configFile string) (exit bool, config *Config) {
	_, err := os.Stat(configFile)
	if err != nil {
		err = createConfig(configFile)
		if err != nil {
			log.Fatalln("Couldn't create config: " + err.Error())
		} else {
			log.Info("Created config sucessfully")
		}
		return true, nil
	}
	config = readConfig(configFile)
	if len(config.Host) == 0 || len(config.Pass) == 0 {
		log.Fatalln("You need to fill the config")
		return true, nil
	}
	return false, config
}
