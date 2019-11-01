package service

import (
	"net/http"

	"strconv"

	"log"

	"time"
	"context"
	"github.com/pangolin-lab/atom/linuxAP/service/router"
	"github.com/pangolin-lab/atom/linuxAP/config"
)

var (
	webserver *http.Server
)

func StartWebDaemon() {
	mux := http.NewServeMux()

	mux.Handle("/ajax/", &router.AjaxRouter{})

	addr := ":" + strconv.Itoa(config.GetAPConfigInst().HttpServerPort)

	log.Println("Web Server Start at", addr)

	webserver = &http.Server{Addr: addr, Handler: mux}


	log.Fatal(webserver.ListenAndServe())
}


func StopWebDaemon() {


	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	webserver.Shutdown(ctx)
}
