package controllers

import (
	"aws-web-server/models"
	"aws-web-server/webserver"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

type UserController struct {
	rdbmsSession *sql.DB
	redisSession *redis.Client
}

func NewUserController(rdbmsSession *sql.DB, redisSession *redis.Client) *UserController {
	return &UserController{
		rdbmsSession: rdbmsSession,
		redisSession: redisSession,
	}
}

func (uc *UserController) GetUser(userName string) (*models.AppUser, error) {
	rows, err := uc.rdbmsSession.Query("SELECT user_name, password FROM users WHERE user_name = $1 LIMIT 1;", userName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return &models.AppUser{}, nil
		}
		log.Println("Some error occurred while querying user:", err)
		return &models.AppUser{}, err
	}
	var user models.AppUser
	for rows.Next() {
		err = rows.Scan(&user.UserName, &user.Password)
		if err != nil {
			log.Println("Some error occurred while scanning user:", err)
			return &models.AppUser{}, err
		}
	}
	return &user, nil
}

func (uc *UserController) InsertUser(user *models.AppUser) (error, int) {
	sqlStatement := `
	INSERT INTO users (user_name, password)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	err := uc.rdbmsSession.QueryRow(sqlStatement, user.UserName, string(password)).Scan(&id)
	if err != nil {
		return err, 0
	}
	log.Println("New record ID is:", id)
	return nil, id
}

func (uc *UserController) DeleteUser(userName string) (error, int) {
	sqlStatement := `
    DELETE FROM users
    WHERE user_name = $1
    RETURNING id`
	id := 0
	err := uc.rdbmsSession.QueryRow(sqlStatement, userName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.Println("Deleted record ID is:", id)
	return nil, id
}

func (uc *UserController) LogoutUser(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie("session")
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
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}

	v, err := webserver.GetRedis(uc.redisSession, c.Value)
	if v == "" {
		return false
	}
	if err != nil {
		if err, ok := err.(*webserver.RedisError); ok {
			log.Println("Error happened in redis: ", err.Error())
		}
		return false
	}
	return true
}

func (uc *UserController) LoginUser(w http.ResponseWriter, req *http.Request) {
	user, err := uc.GetUser(req.FormValue("username"))
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
		Name:   "session",
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

func (uc *UserController) RefreshUserSession(w http.ResponseWriter, req *http.Request) error {
	c, err := req.Cookie("session")
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
