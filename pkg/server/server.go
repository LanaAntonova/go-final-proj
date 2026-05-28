package server

import (
	"net/http"
	"os"
	"strconv"

	"github.com/LanaAntonova/go-final-proj/pkg/api"
)

var Port = 7540

func Start() error {
	port := getPort()
	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		webDir = "../web" // для тестов
	}

	api.Init()

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func getPort() int {
	if envPort := os.Getenv("TODO_PORT"); len(envPort) > 0 {
		if p, err := strconv.Atoi(envPort); err == nil {
			return p
		}
	}
	return Port
}
