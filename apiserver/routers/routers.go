// Coriolis OVM exporter
// Copyright (C) 2021 Cloudbase Solutions SRL
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}/disks/{diskID}", log(os.Stdout, http.HandlerFunc(han.ConsumeSnapshotHandler))).Methods("GET", "HEAD")
	apiRouter.Handle("/vms/{vmID}/snapshots/{snapshotID}/disks/{diskID}/", log(os.Stdout, http.HandlerFunc(han.ConsumeSnapshotHandler))).Methods("GET", "HEAD")

	// Not found handler
	apiRouter.PathPrefix("/").Handler(log(os.Stdout, http.HandlerFunc(han.NotFoundHandler)))

	return router
}
