package errors

import (
	"log"
	"net/http"

	"github.com/kacperborowieckb/gen-sql/utils/json"
)

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	json.WriteJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	json.WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	json.WriteJSONError(w, http.StatusNotFound, "not found")
}
