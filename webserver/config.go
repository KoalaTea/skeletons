package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type FileConfig struct {
	OAuth struct {
		ClientID  string `json:"client_id"`
		SecretKey string `json:"secret_key"`
	} `json:"oauth"`
	PProfEnabled bool `json:"enable_pprof,omitempty"`
}

type Config struct {
	ClientID     string
	SecretKey    string
	PProfEnabled bool
}

func getConfig(fileName string) *Config {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))
	}
	defer f.Close()

	jsonBytes, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))
	}

	cfg := &FileConfig{}
	err = json.Unmarshal(jsonBytes, cfg)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))

	}

	f, err = os.Open(cfg.OAuth.ClientID)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))

	}
	defer f.Close()
	clientIDBytes, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))

	}
	clientID := string(clientIDBytes)

	f, err = os.Open(cfg.OAuth.SecretKey)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))

	}
	defer f.Close()
	secretKeyBytes, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("%v", fmt.Errorf("%w", err))

	}
	secretKey := string(secretKeyBytes)

	return &Config{ClientID: clientID, SecretKey: secretKey, PProfEnabled: cfg.PProfEnabled}
}
