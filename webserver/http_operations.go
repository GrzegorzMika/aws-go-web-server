package webserver

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func HandleError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
	}
}

//goland:noinspection GoUnusedParameter
func Ping(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pong")
	return
}

//goland:noinspection GoUnusedParameter
func Instance(w http.ResponseWriter, req *http.Request) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		log.Warning(err)
		return
	}

	bs := make([]byte, resp.ContentLength)
	resp.Body.Read(bs)
	resp.Body.Close()
	io.WriteString(w, string(bs))
}
