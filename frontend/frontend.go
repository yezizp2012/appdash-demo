package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/yezizp2012/appdash-demo/utils"
)

var (
	api           = flag.String("api", ":8699", "frontend service address")
	collectorAddr = flag.String("collector", "", "appdash collector server address")
	backendAddr   = flag.String("backend", "", "backend server address")
	debug         = flag.Bool("debug", false, "whether to log debug messages")
	tracer        *utils.TraceService
)

func main() {
	flag.Parse()
	log.Info("frontend service start")
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
	router.HandleFunc("/", Home)
	router.HandleFunc("/endpoint", Endpoint)

	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(*api)
}

func Home(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan(r)
	defer span.Finish()

	httpClient := http.DefaultClient
	req, err := http.NewRequest("GET", *backendAddr+"/api/v1", nil)
	if err != nil {
		log.Errorf("new request fail: %v", err)
		fmt.Fprintf(w, `<p>call /api/v1 fail: %v</p>`, err)
		return
	}
	tracer.WrapperRequest(span, req)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("call /api/v1 fail:", err)
		fmt.Fprintf(w, `<p>call /api/v1 fail: %v</p>`, err)
	} else {
		defer resp.Body.Close()
		raw, _ := ioutil.ReadAll(resp.Body)
		log.Infof("call /api/v1 succ: reply=%s", string(raw))
		fmt.Fprintf(w, `<p>call /api/v1 succ: %s</p>`, string(raw))
	}
}

func Endpoint(w http.ResponseWriter, r *http.Request) {
	span := tracer.StartSpan(r)
	defer span.Finish()

	httpClient := http.DefaultClient
	req, err := http.NewRequest("GET", *backendAddr+"/api/v2", nil)
	if err != nil {
		log.Errorf("new request fail: %v", err)
		fmt.Fprintf(w, `<p>call /api/v2 fail: %v</p>`, err)
		return
	}
	tracer.WrapperRequest(span, req)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("call /api/v2 fail:", err)
		fmt.Fprintf(w, `<p>call /api/v2 fail: %v</p>`, err)
	} else {
		defer resp.Body.Close()
		raw, _ := ioutil.ReadAll(resp.Body)
		log.Infof("call /api/v2 succ: reply=%s", string(raw))
		fmt.Fprintf(w, `<p>call /api/v2 succ: %s</p>`, string(raw))
	}
}
