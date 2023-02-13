package main

import (
	"aws-web-server/models"
	"aws-web-server/webserver"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var db *sql.DB
var rdb *redis.Client
var tpl *template.Template

const logFile = "./webserver.log"
const webPort = ":80"
const sessionTimeout = 600

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {
	var err error

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	db, err = webserver.Connect()
	check(err)
	defer db.Close()

	rdb = webserver.ConnectRedis()
	defer rdb.Close()

	err = webserver.CreateTableUsers(db)
	check(err)
	err = webserver.CreateTableTask(db)
	check(err)

	http.HandleFunc("/", index)
	http.HandleFunc("/add", addTask)
	http.HandleFunc("/delete", deleteTask)
	http.HandleFunc("/dog", dog)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/login", loginUser)
	http.HandleFunc("/logout", logoutUser)
	http.HandleFunc("/instance", webserver.Instance)
	http.HandleFunc("/ping", webserver.Ping)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(webPort, nil))
}

func IsAuthenticated(req *http.Request) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}

	v, err := webserver.GetRedis(rdb, c.Value)
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

func dog(w http.ResponseWriter, req *http.Request) {
	if !IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}
	err := webserver.RefreshCookie(w, req, rdb, sessionTimeout)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

	files, err := os.ReadDir("./assets/")
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
	img := files[rand.Intn(len(files))].Name()
	log.Println(img)
	//img = fmt.Sprintf(`<img src="/assets/%s">`, img)
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	err = tpl.ExecuteTemplate(w, "dog.gohtml", img)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	if !IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	err := webserver.RefreshCookie(w, req, rdb, sessionTimeout)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

	var Tasks []models.Task

	rows, err := db.Query("SELECT task_name, due_date FROM task ORDER BY due_date DESC;")
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
	for rows.Next() {
		var task models.Task
		err = rows.Scan(&task.TaskName, &task.DueDate)
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
		Tasks = append(Tasks, task)
	}

	if len(Tasks) == 0 {
		http.Redirect(w, req, "/dog", http.StatusSeeOther)
		return
	}

	err = tpl.ExecuteTemplate(w, "index.gohtml", Tasks)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

}

func addTask(w http.ResponseWriter, req *http.Request) {
	if !IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	var err error
	var id int

	err = webserver.RefreshCookie(w, req, rdb, sessionTimeout)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

	if req.Method == http.MethodPost {
		var task models.Task
		err = req.ParseForm()
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
		task.TaskName = req.PostForm["TaskName"][0]
		task.DueDate = req.PostForm["DueDate"][0]
		_, err = time.Parse("2006-01-02", task.DueDate)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Invalid value for due date provided: %s. Expected is date of form 2023-10-01", task.DueDate)
			return
		}
		err, id = webserver.InsertTask(db, &task)
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
	}
	err = tpl.ExecuteTemplate(w, "add.gohtml", id)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
}

func deleteTask(w http.ResponseWriter, req *http.Request) {
	if !IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	var err error
	var id int

	err = webserver.RefreshCookie(w, req, rdb, sessionTimeout)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

	if req.Method == http.MethodPost {
		err = req.ParseForm()
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
		taskName := req.PostForm["deleteTaskName"][0]
		err, id = webserver.DeleteTask(db, taskName)
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
	}

	err = tpl.ExecuteTemplate(w, "delete.gohtml", id)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
}

func loginUser(w http.ResponseWriter, req *http.Request) {
	if IsAuthenticated(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		var user models.AppUser
		un := req.FormValue("username")
		p := req.FormValue("password")
		user, err := webserver.GetUser(db, un)
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// does the entered password match the stored password?
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(p))
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// create session
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name:   "session",
			Value:  sID.String(),
			MaxAge: sessionTimeout,
		}
		http.SetCookie(w, c)
		err = webserver.SetRedis(rdb, c.Value, user.UserName, sessionTimeout)
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	err := tpl.ExecuteTemplate(w, "login.gohtml", nil)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
}

func logoutUser(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie("session")
	if err != nil {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}
	err = webserver.DelRedis(rdb, c.Value)
	if err != nil {
		webserver.HandleError(err, w)
	}
	c.MaxAge = -1
	http.SetCookie(w, c)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}
