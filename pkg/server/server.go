package server

import (
	"inspection-server/pkg/api"
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

func Start() http.Handler {
	router := mux.NewRouter()
	router.UseEncodedPath()

	debugHandle(router)

	router.Methods(http.MethodGet).Path("/v1/reports/get/{id}").Handler(api.GetReport())
	router.Methods(http.MethodPost).Path("/v1/reports/print").Handler(api.PrintReport())

	router.Methods(http.MethodGet).Path("/v1/plans/get/{id}").Handler(api.GetPlan())
	router.Methods(http.MethodGet).Path("/v1/plans/list").Handler(api.ListPlan())
	router.Methods(http.MethodPost).Path("/v1/plans/create").Handler(api.CreatePlan())
	router.Methods(http.MethodDelete).Path("/v1/plans/delete/{id}").Handler(api.DeletePlan())

	router.Methods(http.MethodGet).Path("/v1/records/get/{id}").Handler(api.GetRecord())
	router.Methods(http.MethodGet).Path("/v1/records/list").Handler(api.ListRecord())
	router.Methods(http.MethodDelete).Path("/v1/records/delete/{id}").Handler(api.DeleteRecord())

	router.Methods(http.MethodGet).Path("/v1/templates/get/{id}").Handler(api.GetTemplate())
	router.Methods(http.MethodGet).Path("/v1/templates/list").Handler(api.ListTemplate())
	router.Methods(http.MethodPost).Path("/v1/templates/create").Handler(api.CreateTemplate())
	router.Methods(http.MethodPut).Path("/v1/templates/update").Handler(api.UpdateTemplate())
	router.Methods(http.MethodDelete).Path("/v1/templates/delete/{id}").Handler(api.DeleteTemplate())

	router.Methods(http.MethodGet).Path("/v1/clusters/list").Handler(api.GetClusters())
	router.Methods(http.MethodGet).Path("/v1/clusters/{id}/resource/list").Handler(api.GetResource())

	return router
}

func debugHandle(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)

	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
}
