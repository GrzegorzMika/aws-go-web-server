package main

import (
	"aws-web-server/controllers"
	"aws-web-server/models"
	"aws-web-server/webserver"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
)

var db *sql.DB
var rdb *redis.Client
var tpl *template.Template

const logFile = "./webserver.log"
const webPort = ":80"

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

	appController := controllers.NewAppController(db, rdb, tpl)

	http.HandleFunc("/", index)
	http.HandleFunc("/add", appController.AddTask)
	http.HandleFunc("/delete", appController.DeleteTask)
	http.HandleFunc("/dog", dog)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/login", appController.LoginUser)
	http.HandleFunc("/logout", appController.LogoutUser)
	http.HandleFunc("/instance", webserver.Instance)
	http.HandleFunc("/ping", webserver.Ping)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(webPort, nil))
}

func dog(w http.ResponseWriter, req *http.Request) {
	if !IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}
	err := webserver.RefreshCookie(w, req, rdb, models.SessionTimeout)
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

	err := webserver.RefreshCookie(w, req, rdb, models.SessionTimeout)
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
