package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

//Config config for the server
type Config struct {
	Host          string `json:"host"`
	Username      string `json:"username"`
	Pass          string `json:"pass"`
	DatabasePort  int    `json:"dbport"`
	CertFile      string `json:"cert"`
	KeyFile       string `json:"key"`
	IPdataAPIKey  string `json:"ipdataAPIkey"`
	ShowTimeInLog bool   `json:"showLogTime"`
	HTTPPort      int    `json:"port"`
	TLSPort       int    `json:"porttls"`
}

func createConfig(configFile string) error {
	LogError("Couldn't find " + configFile)
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
