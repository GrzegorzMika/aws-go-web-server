package webserver

import (
	"fmt"
	"github.com/pkg/errors"
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
	_, err := fmt.Fprintf(w, "Pong")
	if err != nil {
		HandleError(errors.Wrap(err, "Failed to response with status OK"), w)
		return
	}
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
	_, err = resp.Body.Read(bs)
	if err != nil {
		log.WithField("error", err).Warning("Failed to get instance id")
		return
	}
	_ = resp.Body.Close()
	_, err = io.WriteString(w, string(bs))
	if err != nil {
		HandleError(errors.Wrap(err, "Failed to write the instance ID to output"), w)
		return
	}
}
