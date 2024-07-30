package server

import (
	"inspection-server/pkg/api"
	"inspection-server/pkg/core"
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

	router.Methods(http.MethodGet).Path("/v1/notify/get/{id}").Handler(api.GetNotify())
	router.Methods(http.MethodGet).Path("/v1/notify/list").Handler(api.ListNotify())
	router.Methods(http.MethodPost).Path("/v1/notify/create").Handler(api.CreateNotify())
	router.Methods(http.MethodPut).Path("/v1/notify/update").Handler(api.UpdateNotify())
	router.Methods(http.MethodDelete).Path("/v1/notify/delete/{id}").Handler(api.DeleteNotify())
	router.Methods(http.MethodPost).Path("/v1/notify/test").Handler(api.TestNotify())

	router.Methods(http.MethodGet).Path("/v1/clusters/list").Handler(api.GetClusters())
	router.Methods(http.MethodGet).Path("/v1/clusters/{id}/resource/list").Handler(api.GetResource())

	router.Methods(http.MethodGet).Path("/v1/agent/list").Handler(api.ListAgent())
	router.Methods(http.MethodDelete).Path("/v1/agent/delete/{id}").Handler(api.DeleteAgent())

	router.Methods(http.MethodGet).Path("/v1/grafana").Handler(core.GetGrafana())

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
