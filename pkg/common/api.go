package common

import (
	"log"
	"net/http"
)

func HandleError(rw http.ResponseWriter, code int, err error) {
	// Logging the error with status code
	log.Printf("HTTP %d - %s", code, err.Error())

	// Setting the response header with the error code
	rw.WriteHeader(code)

	// Writing the error message to the response body
	_, writeErr := rw.Write([]byte(err.Error()))
	if writeErr != nil {
		// Logging in case of failure to write the error message to response
		log.Printf("Failed to write error message to response: %s", writeErr.Error())
	}
}
