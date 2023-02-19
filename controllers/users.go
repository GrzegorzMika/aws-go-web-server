package controllers

import (
	"aws-web-server/models"
	"aws-web-server/webserver"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UserController struct {
	rdbmsSession *sql.DB
	redisSession *redis.Client
	appContext   context.Context
}

func NewUserController(ctx context.Context, rdbmsSession *sql.DB, redisSession *redis.Client) *UserController {
	return &UserController{
		rdbmsSession: rdbmsSession,
		redisSession: redisSession,
		appContext:   ctx,
	}
}

func (uc *UserController) LogoutUser(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie(models.SessionCookieName)
	if err != nil {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}
	err = webserver.DelRedis(uc.redisSession, c.Value)
	if err != nil {
		webserver.HandleError(err, w)
	}
	c.MaxAge = -1
	http.SetCookie(w, c)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func (uc *UserController) IsAuthenticated(req *http.Request) bool {
	c, err := req.Cookie(models.SessionCookieName)
	if err != nil {
		return false
	}

	v, err := webserver.GetRedis(uc.redisSession, c.Value)
	if v == "" {
		return false
	}
	if err != nil {
		if err, ok := err.(*webserver.RedisError); ok {
			log.Error("Error happened in redis: ", err.Error())
		}
		return false
	}
	return true
}

func (uc *UserController) LoginUser(w http.ResponseWriter, req *http.Request) {
	user, err := models.GetUser(uc.appContext, uc.rdbmsSession, req.FormValue("username"))
	if err != nil {
		http.Error(w, "Username and/or password do not match", http.StatusForbidden)
		return
	}
	// does the entered password match the stored password?
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.FormValue("password")))
	if err != nil {
		http.Error(w, "Username and/or password do not match", http.StatusForbidden)
		return
	}
	// create session
	sID := uuid.NewV4()
	c := &http.Cookie{
		Name:   models.SessionCookieName,
		Value:  sID.String(),
		MaxAge: models.SessionTimeout,
	}
	http.SetCookie(w, c)
	err = webserver.SetRedis(uc.redisSession, c.Value, user.UserName, models.SessionTimeout)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
	return
}

func (uc *UserController) refreshUserSession(w http.ResponseWriter, req *http.Request) error {
	c, err := req.Cookie(models.SessionCookieName)
	if err != nil {
		return errors.Wrap(err, "Failed to get session cookie")
	}
	userName, err := webserver.GetRedis(uc.redisSession, c.Value)
	if err == nil {
		err = webserver.SetRedis(uc.redisSession, c.Value, userName, models.SessionTimeout)
		if err != nil {
			if err, ok := err.(*webserver.RedisError); ok {
				return errors.Wrap(err.Err, err.Msg)
			}
			return errors.Wrap(err, "Failed to set redis session info")
		}
	} else {
		if err, ok := err.(*webserver.RedisError); ok {
			return errors.Wrap(err.Err, err.Msg)
		}
		return errors.Wrap(err, "Failed to connect to redis")
	}
	// refresh session
	c.MaxAge = models.SessionTimeout
	http.SetCookie(w, c)
	return nil
}

func (uc *UserController) RefreshUserSession(w http.ResponseWriter, req *http.Request) {
	err := uc.refreshUserSession(w, req)
	if err != nil {
		webserver.HandleError(err, w)
	}
}
