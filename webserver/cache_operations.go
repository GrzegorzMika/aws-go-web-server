package webserver

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

const (
	RedisHost     = "gowebservercache-003.g2o7mj.0001.eun1.cache.amazonaws.com"
	RedisPort     = "6379"
	RedisPassword = ""
	RedisDB       = 0
)

var ctx = context.Background()

type RedisError struct {
	Err error
	Msg string
}

func (e *RedisError) Error() string {
	return e.Err.Error()
}

func ConnectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     RedisHost + ":" + RedisPort,
		Password: RedisPassword,
		DB:       RedisDB,
	})
	log.Println("Connected to Redis server")
	return rdb
}

func SetRedis(rdb *redis.Client, key string, value string, timeout int) error {
	expire := time.Second * time.Duration(timeout)
	err := rdb.SetEx(ctx, key, value, expire).Err()
	if err == nil {
		return nil
	}
	return &RedisError{
		Err: err,
		Msg: fmt.Sprintf("Failed to set %s to %s within %s seconds", key, value, expire),
	}
}

func GetRedis(rdb *redis.Client, key string) (string, error) {
	v, err := rdb.Get(ctx, key).Result()
	if err == nil {
		return v, nil
	}
	if err == redis.Nil {
		return "", nil
	}
	return v, &RedisError{
		Err: err,
		Msg: fmt.Sprintf("Failed to get %s", key),
	}
}

func DelRedis(rdb *redis.Client, key string) error {
	err := rdb.Del(ctx, key).Err()
	if err == nil {
		return nil
	}
	return &RedisError{
		Err: err,
		Msg: fmt.Sprintf("Failed to delete %s", key),
	}
}
