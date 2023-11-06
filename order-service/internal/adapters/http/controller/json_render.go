package controller

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message    string `json:"message"`
	httpStatus int
}

func JsonRender(w http.ResponseWriter, resp Response) error {
	w.WriteHeader(resp.httpStatus)
	w.Header().Set("Content-Type", "application/json")

	out, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}
