package routers

import (
	"net/http"
	"os"

	"coriolis-ovm-exporter/apiserver/auth"
	"coriolis-ovm-exporter/apiserver/controllers"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// NewAPIRouter returns a new gorilla mux router.
func NewAPIRouter(han *controllers.APIController, authMiddleware auth.Middleware) *mux.Router {
	router := mux.NewRouter()
	log := gorillaHandlers.CombinedLoggingHandler

	apiSubRouter := router.PathPrefix("/api/v1").Subrouter()

	// Login
	authRouter := apiSubRouter.PathPrefix("/auth").Subrouter()
	authRouter.Handle("/{login:login\\/?}", log(os.Stdout, http.HandlerFunc(han.LoginHandler))).Methods("POST")

	// Private API endpoints
	apiRouter := apiSubRouter.PathPrefix("").Subrouter()
	apiRouter.Use(authMiddleware.Middleware)

	// list VMs
	apiRouter.Handle("/vms", log(os.Stdout, http.HandlerFunc(han.ListVMsHandler))).Methods("GET")
	apiRouter.Handle("/vms/", log(os.Stdout, http.HandlerFunc(han.ListVMsHandler))).Methods("GET")
	// get VM
	apiRouter.Handle("/vms/{vmID}", log(os.Stdout, http.HandlerFunc(han.GetVMHandler))).Methods("GET")
	apiRouter.Handle("/vms/{vmID}/", log(os.Stdout, http.HandlerFunc(han.GetVMHandler))).Methods("GET")
	// list VM snapshots
	apiRouter.Handle("/vms/{vmID}/snapshots", log(os.Stdout, http.HandlerFunc(han.ListSnapshotsHandler))).Methods("GET")
	apiRouter.Handle("/vms/{vmID}/snapshots/", log(os.Stdout, http.HandlerFunc(han.ListSnapshotsHandler))).Methods("GET")
	// delete all VM snapshots
	apiRouter.Handle("/vms/{vmID}/snapshots", log(os.Stdout, http.HandlerFunc(han.PurgeSnapshotsHandler))).Methods("DELETE")
	apiRouter.Handle("/vms/{vmID}/snapshots/", log(os.Stdout, http.HandlerFunc(han.PurgeSnapshotsHandler))).Methods("DELETE")
	// create VM snapshot
	apiRouter.Handle("/vms/{vmID}/snapshots", log(os.Stdout, http.HandlerFunc(han.CreateSnapshotHandler))).Methods("POST")
	apiRouter.Handle("/vms/{vmID}/snapshots/", log(os.Stdout, http.HandlerFunc(han.CreateSnapshotHandler))).Methods("POST")
	// get VM snapshot
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}", log(os.Stdout, http.HandlerFunc(han.GetSnapshotHandler))).Methods("GET")
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}/", log(os.Stdout, http.HandlerFunc(han.GetSnapshotHandler))).Methods("GET")
	// delete VM snapshot
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}", log(os.Stdout, http.HandlerFunc(han.DeleteSnapshotHandler))).Methods("DELETE")
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}/", log(os.Stdout, http.HandlerFunc(han.DeleteSnapshotHandler))).Methods("DELETE")
	// Read snapshotted disk
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}/disks/{diskID}", log(os.Stdout, http.HandlerFunc(han.ConsumeSnapshotHandler))).Methods("GET")
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}/disks/{diskID}/", log(os.Stdout, http.HandlerFunc(han.ConsumeSnapshotHandler))).Methods("GET")

	return router
}
