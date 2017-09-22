package main

import (
	"errors"
	"flag"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/yezizp2012/appdash-demo/utils"
)

var (
	api           = flag.String("api", ":8200", "frontend service address")
	collectorAddr = flag.String("collector", "", "appdash collector server address")
	debug         = flag.Bool("debug", false, "whether to log debug messages")
	tracer        *utils.TraceService
)

func main() {
	flag.Parse()
	log.Info("backend service start")
	if *collectorAddr == "" {
		log.Infof("appdash collector should not be empty")
		panic(errors.New("empty appdash collector"))
	}

	localAddr, err := utils.GetLocalAddress()
	if err != nil {
		panic(err)
	}

	tracer = utils.NewTraceService(*collectorAddr, localAddr+*api, *debug)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1", API)
	router.HandleFunc("/api/v2", APIv2)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(*api)
}

func API(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan(r)
	defer span.Finish()
	log.Infof("got span info: %v", span)

	time.Sleep(time.Millisecond * 100)
	utils.WriteJSON(w, http.StatusOK, utils.M{"span": span.Context(), "sleep": time.Millisecond * 100})
}

func APIv2(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan(r)
	defer span.Finish()
	log.Infof("got span info: %v", span)

	time.Sleep(time.Millisecond * 200)
	utils.WriteJSON(w, http.StatusOK, utils.M{"span": span.Context(), "sleep": time.Millisecond * 200})
}
