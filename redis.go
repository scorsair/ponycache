package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"path"
	"time"
)

func initRedis(c *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Redisdb.Addr,
		Password: c.Redisdb.Passwd,
		DB:       c.Redisdb.Db,
	})
}

func formatStdout(k, v string) {
	if v != "" {
		str := fmt.Sprintf("%s=%s\n", k, v)
		os.Stdout.WriteString(str)
	}
}

func hsetEnv(r *redis.Client, e Entry) error {
	var err error

	if e.value != "" {
		_, err = r.HSet(e.id, e.key, e.value).Result()
	} else {
		resp := fmt.Sprintf("WARNING: Value is empty for id: %s, key: %s!\n", e.id, e.key)
		os.Stderr.WriteString(resp)
		return err
	}

	if e.ttl > 0 {
		expire := time.Duration(e.ttl) * time.Minute
		r.Expire(e.id, expire)
	}
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%s: %s", "ERROR: Something went wrong!\n", err))
		os.Exit(1)
	}
	return err
}

func hgetEnv(r *redis.Client, e Entry) error {
	if e.key != "" {
		result, err := r.HGet(e.id, e.key).Result()
		formatStdout(e.key, result)
		return err
	} else {
		result, err := r.HGetAll(e.id).Result()
		for i, v := range result {
			formatStdout(i, v)
		}
		return err
	}
}

// Set filename to the redis for S3
func setEnv(r *redis.Client, e Entry) error {
	basename := path.Base(e.value)
	return r.Set(e.sha1, basename, 0).Err()
}

// Return filename from sha1
func getEnv(r *redis.Client, s string) (string, error) {
	return r.Get(s).Result()
}
