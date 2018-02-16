package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

type Entry struct {
	id    string
	key   string
	value string
	ttl   int
	sha1  string
}

type redisdb struct {
	Addr   string
	Passwd string
	Db     int
}

type storage struct {
	Access_key string
	Secret_key string
	Token      string
	Bucket     string
	Endpoint   string
	Region     string
}

type basic struct {
	Ttl int
}

type Config struct {
	purpose bool    // false: variable, true: file
	action  int     // 0: get, 1: set, 2: rm/unset
	Basic   basic   `toml:"default"` // the default value is 60 minutes
	Redisdb redisdb `toml:"redis"`
	Storage storage
	entries []Entry
}

// The return value shows a status of the file
func isFileExist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	// Be sure that path is not a directory
	if !stat.IsDir() {
		return true
	} else {
		return false
	}
}

// Parsing of toml config file
func fillConfig(c *Config) *Config {
	// Default paths for config
	var config_paths = [2]string{"ponycache.toml", "/etc/ponycache.toml"}

	// Conf selection
	var toml_conf string
	for _, v := range config_paths {
		if isFileExist(v) {
			toml_conf = v
			break
		} else {
			toml_conf = "ponycache.toml"
		}
		// fmt.Println(toml_conf)
	}

	if _, err := toml.DecodeFile(toml_conf, c); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return c
}
