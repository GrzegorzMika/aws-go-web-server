package webserver

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

var ctx = context.Background()

type RedisError struct {
	Err error
	Msg string
}

func (e *RedisError) Error() string {
	return e.Err.Error()
}

func ConnectRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	log.Println("INFO: Connected to Redis server")
	return rdb, nil
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
