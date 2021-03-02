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

	return router
}
