package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"coriolis-ovm-exporter/apiserver/auth"
	"coriolis-ovm-exporter/apiserver/controllers"
	"coriolis-ovm-exporter/apiserver/routers"
	"coriolis-ovm-exporter/config"
)

var (
	conf = flag.String("config", config.DefaultConfigFile, "exporter config file")
)

func main() {
	flag.Parse()
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)

	cfg, err := config.ParseConfig(*conf)
	if err != nil {
		log.Fatalf("failed to parse config %s: %q", *conf, err)
	}

	controller, err := controllers.NewAPIController(cfg)
	if err != nil {
		log.Fatalf("failed to create controller: %q", err)
	}

	jwt, err := auth.NewJWTMiddleware(&cfg.JWTAuth)
	if err != nil {
		log.Fatalf("failed to get authentication middleware: %q", err)
	}
	router := routers.NewAPIRouter(controller, jwt)

	tlsCfg, err := cfg.APIServer.TLSConfig.TLSConfig()
	if err != nil {
		log.Fatalf("failed to get TLS config: %q", err)
	}

	srv := &http.Server{
		Addr:      cfg.APIServer.BindAddress(),
		TLSConfig: tlsCfg,
		// Pass our instance of gorilla/mux in.
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServeTLS(
			cfg.APIServer.TLSConfig.Cert,
			cfg.APIServer.TLSConfig.Key); err != nil {

			log.Fatal(err)
		}
	}()

	<-stop
}
