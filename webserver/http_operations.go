package webserver

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
)

func HandleError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalln(err)
	}
}

func Ping(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pong")
	return
}

func Instance(w http.ResponseWriter, req *http.Request) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		fmt.Println(err)
		return
	}

	bs := make([]byte, resp.ContentLength)
	resp.Body.Read(bs)
	resp.Body.Close()
	io.WriteString(w, string(bs))
}

func RefreshCookie(w http.ResponseWriter, req *http.Request, cacheDb *redis.Client, expiration int) error {
	c, err := req.Cookie("session")
	if err != nil {
		return errors.Wrap(err, "failed to get session cookie")
	}
	userName, err := GetRedis(cacheDb, c.Value)
	if err == nil {
		err = SetRedis(cacheDb, c.Value, userName, expiration)
		if err != nil {
			if err, ok := err.(*RedisError); ok {
				return errors.Wrap(err.Err, err.Msg)
			}
			return errors.Wrap(err, "failed to set redis session info")
		}
	} else {
		if err, ok := err.(*RedisError); ok {
			return errors.Wrap(err.Err, err.Msg)
		}
		return errors.Wrap(err, "failed to connect to redis")
	}
	// refresh session
	c.MaxAge = expiration
	http.SetCookie(w, c)
	return nil
}
