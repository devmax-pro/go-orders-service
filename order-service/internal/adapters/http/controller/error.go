package controller

import (
	logs "github.com/devmax-pro/order-service/internal/adapters/logger"
	"net/http"
)

func InternalError(message string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, message, w, r, "Internal server error", http.StatusInternalServerError)
}

func Unauthorised(message string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, message, w, r, "Unauthorised", http.StatusUnauthorized)
}

func BadRequest(message string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, message, w, r, "Bad request", http.StatusBadRequest)
}

func NotFound(message string, err error, w http.ResponseWriter, r *http.Request) {
	httpRespondWithError(err, message, w, r, "Entity not found", http.StatusNotFound)
}

func httpRespondWithError(err error, logMsg string, w http.ResponseWriter, r *http.Request, respMsg string, status int) {
	logs.Error(logMsg, err)
	resp := Response{Message: respMsg, httpStatus: status}

	if err := JsonRender(w, resp); err != nil {
		panic(err)
	}
}
