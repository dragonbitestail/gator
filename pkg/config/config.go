package gatorconfig

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const confFile = ".gatorconfig.json"

type Config struct {
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	var config = Config{}

	fullPath, err := getFullConfigPath()
	if err != nil {
		return nil, err
	}
	
	oFile, err := os.Open(fullPath)
	if err != nil {
	    return &config, err
	}
	defer oFile.Close()

	jDec := json.NewDecoder(oFile)
	if err := jDec.Decode(&config); err != nil {
		log.Fatal(err)
	}

	return &config, nil
}

func (c *Config) SetUser(user string) error {

	fullPath, err := getFullConfigPath()
	if err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var buffer bytes.Buffer

	c.CurrentUserName = user

	json.NewEncoder(&buffer).Encode(c)

	f.Write(buffer.Bytes())

	return nil
}


func getFullConfigPath() (string, error) {
	cFile := confFile
	if homeDir, err := os.UserHomeDir(); err != nil {
		return cFile, err
	} else {
		cFile = filepath.Join(homeDir, cFile)
	}

	return cFile, nil
}
