package main

import (
	"flag"
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/Sirupsen/logrus"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

var (
	api           = flag.String("api", ":7701", "app dash collector start address")
	ui            = flag.String("ui", ":7700", "app dash web ui start address")
	debug         = flag.Bool("debug", false, "whether to log debug messages")
	trace         = flag.Bool("trace", true, "whether to log all data that is received")
	defaultServer *Server
)

type Server struct {
	CollectorAddr string `long:"collector" description:"collector listen address" default:":7701"`
	HTTPAddr      string `long:"http" description:"HTTP listen address" default:":7700"`

	StoreFile       string        `short:"f" long:"store-file" description:"persisted store file" default:"/tmp/appdash.gob"`
	PersistInterval time.Duration `short:"p" long:"persist-interval" description:"interval between persisting store to file" default:"2s"`

	Debug bool `short:"d" long:"debug" description:"debug log"`
	Trace bool `long:"trace" description:"trace log"`

	DeleteAfter time.Duration `long:"delete-after" description:"delete traces after a certain age (0 to disable)" default:"30m"`

	BasicAuth string `long:"basic-auth" description:"if set to 'user:passwd', require HTTP Basic Auth for web app"`
}

func main() {
	flag.Parse()
	log.Info("starting appdash server!")

	var (
		memStore = appdash.NewMemoryStore()
		Store    = appdash.Store(memStore)
		Queryer  = memStore
	)

	defaultServer = &Server{
		CollectorAddr: *api,
		HTTPAddr:      *ui,
		Debug:         *debug,
		Trace:         *trace,
	}

	url, err := url.Parse("http://localhost" + defaultServer.HTTPAddr)
	if err != nil {
		panic(err)
	}
	app, err := traceapp.New(nil, url)
	if err != nil {
		log.Errorf("new trace app fail: %v", err)
		panic(err)
	}
	app.Store = Store
	app.Queryer = Queryer

	handle := app
	l, err := net.Listen("tcp", defaultServer.CollectorAddr)
	if err != nil {
		log.Errorf("start collector fail:%v", err)
		panic(err)
	}
	log.Infof("appdash collector listening on %s", defaultServer.CollectorAddr)

	cs := appdash.NewServer(l, appdash.NewLocalCollector(Store))
	cs.Debug = defaultServer.Debug
	cs.Trace = defaultServer.Trace
	go cs.Start()

	log.Infof("appdash Web UI start on %s", defaultServer.HTTPAddr)
	log.Fatal(http.ListenAndServe(defaultServer.HTTPAddr, handle))
}
