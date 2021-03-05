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
	"coriolis-ovm-exporter/util"
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

	logWriter, err := util.GetLoggingWriter(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logWriter)

	controller, err := controllers.NewAPIController(cfg)
	if err != nil {
		log.Fatalf("failed to create controller: %q", err)
	}

	jwt, err := auth.NewJWTMiddleware(&cfg.JWTAuth)
	if err != nil {
		log.Fatalf("failed to get authentication middleware: %q", err)
	}
	router := routers.NewAPIRouter(controller, jwt, logWriter)

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
