package common

import "net/http"

func HandleError(rw http.ResponseWriter, code int, err error) {
	rw.WriteHeader(code)
	rw.Write([]byte(err.Error()))
}
