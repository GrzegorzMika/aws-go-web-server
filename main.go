package main

import (
	"aws-web-server/controllers"
	"aws-web-server/models"
	"aws-web-server/webserver"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"os"
)

var db *sql.DB
var rdb *redis.Client
var tpl *template.Template

const logFile = "./logs/webserver.log"
const webPort = ":80"

func check(err error) {
	if err != nil {
		log.Error(err)
	}
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {
	var err error

	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "@timestamp",
			log.FieldKeyMsg:  "message",
		},
	})
	log.SetLevel(log.TraceLevel)

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	db, err = webserver.Connect()
	check(err)
	defer db.Close()

	rdb, err = webserver.ConnectRedis()
	check(err)
	defer rdb.Close()

	s3bucket := models.NewS3Bucket("eu-north-1")
	check(err)

	err = webserver.CreateTableUsers(db)
	check(err)
	err = webserver.CreateTableTask(db)
	check(err)

	appController := controllers.NewAppController(db, rdb, tpl, s3bucket)

	http.HandleFunc("/", appController.ShowList)
	http.HandleFunc("/add", appController.AddTask)
	http.HandleFunc("/delete", appController.DeleteTask)
	http.HandleFunc("/success", appController.SuccessPage)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/login", appController.LoginUser)
	http.HandleFunc("/logout", appController.LogoutUser)
	http.HandleFunc("/instance", webserver.Instance)
	http.HandleFunc("/ping", webserver.Ping)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(webPort, nil))
}
