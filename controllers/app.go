package controllers

import (
	"aws-web-server/models"
	"aws-web-server/webserver"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type AppController struct {
	UserController
	TaskController
	AssetController
	tpl *template.Template
}

func NewAppController(rdbmsSession *sql.DB, redisSession *redis.Client, appTemplates *template.Template, s3bucket *models.S3Bucket) *AppController {
	return &AppController{
		UserController:  *NewUserController(rdbmsSession, redisSession),
		TaskController:  *NewTaskController(rdbmsSession),
		AssetController: *NewAssetController(s3bucket),
		tpl:             appTemplates,
	}
}

func (ac *AppController) ExecuteTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := ac.tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
}

func (ac *AppController) LoginUser(w http.ResponseWriter, req *http.Request) {
	if ac.IsAuthenticated(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		ac.UserController.LoginUser(w, req)
	}

	ac.ExecuteTemplate(w, "login.gohtml", nil)
}

func (ac *AppController) LogoutUser(w http.ResponseWriter, req *http.Request) {
	ac.UserController.LogoutUser(w, req)
}

func (ac *AppController) DeleteTask(w http.ResponseWriter, req *http.Request) {
	if !ac.IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}
	var id int
	var err error

	if req.Method == http.MethodPost {
		err = req.ParseForm()
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
		taskName := req.PostForm["deleteTaskName"][0]
		err, id = models.DeleteTask(ac.TaskController.rdbmsSession, taskName)
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
	}

	ac.RefreshUserSession(w, req)
	ac.ExecuteTemplate(w, "delete.gohtml", id)
}

func (ac *AppController) AddTask(w http.ResponseWriter, req *http.Request) {
	if !ac.IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	var err error
	var id int

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
			err = errors.Wrap(err, fmt.Sprintf("Invalid value for due date provided: %s. "+
				"Expected is date of form 2023-10-01", task.DueDate))
			webserver.HandleError(err, w)
			return
		}
		err, id = models.InsertTask(ac.TaskController.rdbmsSession, &task)
		if err != nil {
			webserver.HandleError(err, w)
			return
		}
	}

	ac.RefreshUserSession(w, req)
	ac.ExecuteTemplate(w, "add.gohtml", id)
}

func (ac *AppController) ShowList(w http.ResponseWriter, req *http.Request) {
	if !ac.IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	var Tasks []models.Task
	Tasks, err := models.GetAllTasks(ac.TaskController.rdbmsSession)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

	if len(Tasks) == 0 {
		http.Redirect(w, req, "/success", http.StatusSeeOther)
		return
	}

	ac.RefreshUserSession(w, req)
	ac.ExecuteTemplate(w, "index.gohtml", Tasks)
}

func (ac *AppController) SuccessPage(w http.ResponseWriter, req *http.Request) {
	if !ac.IsAuthenticated(req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	files, err := ac.AssetController.sb.ListS3Content("go-web-server-assets")
	if err != nil {
		webserver.HandleError(err, w)
		return
	}

	imgName := files[rand.Intn(len(files))]
	log.Printf("imgName: %s", imgName)
	img, err := ac.AssetController.sb.GetURL("go-web-server-assets", imgName)
	if err != nil {
		webserver.HandleError(err, w)
		return
	}
	log.Printf("img: %s", img)
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	ac.RefreshUserSession(w, req)
	ac.ExecuteTemplate(w, "success.gohtml", img)
}
