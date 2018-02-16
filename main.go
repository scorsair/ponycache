package main

import (
	// "fmt"
	"os"
)

func main() {
	var config Config
	fillConfig(&config)
	parseArgs(&config)

	// Initializing storagetr
	storage := initAwsConfig(&config)
	redis := initRedis(&config)
	defer redis.Close()

	// Check redis connection
	if redis.Ping().Err() != nil {
		os.Stderr.WriteString("Connection to redis db has failed!")
		os.Exit(1)
	}

	// var err error
	if !config.purpose {
		for _, v := range config.entries {
			if config.action == 0 {
				hgetEnv(redis, v)
			} else if config.action == 1 {
				hsetEnv(redis, v)
			} else {
				os.Exit(1)
			}
		}
	} else {
		for _, v := range config.entries {
			if config.action == 0 {

				if v.key == "" {
					pack := fillKeysFromBucket(v, &config, storage, redis)
					for _, v := range pack {
						download(v, &config, storage, redis)
					}
				} else {
					download(v, &config, storage, redis)
				}
			} else if config.action == 1 {
				upload(v, &config, storage, redis)
			} else {
				os.Exit(1)
			}
		}
	}
	os.Exit(0)
}
