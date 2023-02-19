package main

import (
	"aws-web-server/controllers"
	"aws-web-server/models"
	"aws-web-server/webserver"
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
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
		log.Fatalf("Error opening log file: %v", err)
	}
	defer webserver.DeferredError(f)
	log.SetOutput(f)

	ctx := context.Background()

	db, err = webserver.Connect()
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to connect to PostgresSQl database"))
	}
	defer webserver.DeferredError(db)

	rdb, err = webserver.ConnectRedis()
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to connect to Redis database"))
	}
	defer webserver.DeferredError(rdb)

	s3bucket := models.NewS3Bucket("eu-north-1")
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to connect to S3 bucket"))
	}

	err = webserver.CreateTableUsers(db)
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to create users table"))
	}
	err = webserver.CreateTableTask(db)
	if err != nil {
		log.Error(errors.Wrap(err, "Failed to create task table"))
	}

	appController := controllers.NewAppController(ctx, db, rdb, tpl, s3bucket)

	http.HandleFunc("/", appController.ShowList)
	http.HandleFunc("/add", appController.AddTask)
	http.HandleFunc("/delete", appController.DeleteTask)
	http.HandleFunc("/success", appController.SuccessPage)
	http.HandleFunc("/login", appController.LoginUser)
	http.HandleFunc("/logout", appController.LogoutUser)
	http.HandleFunc("/instance", webserver.Instance)
	http.HandleFunc("/ping", webserver.Ping)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(webPort, nil))
}
