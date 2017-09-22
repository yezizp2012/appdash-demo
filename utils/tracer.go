package utils

import (
	"net/http"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/appdash"
	appdashtracer "sourcegraph.com/sourcegraph/appdash/opentracing"
)

const headerTagPrefix = "Request.Header."

type TraceService struct {
	collector *appdash.RemoteCollector
	tracer    opentracing.Tracer
	prefix    string
}

func addHeaderTags(span opentracing.Span, h http.Header) {
	for k, v := range h {
		span.SetTag(headerTagPrefix+k, strings.Join(v, ", "))
	}
}

func NewTraceService(addr, prefix string, debug bool) *TraceService {
	collector := appdash.NewRemoteCollector(addr)
	collector.Debug = debug

	tracer := appdashtracer.NewTracer(collector)
	opentracing.InitGlobalTracer(tracer)

	return &TraceService{
		collector: collector,
		tracer:    tracer,
		prefix:    prefix,
	}
}

func (ts *TraceService) StartSpan(r *http.Request) (span opentracing.Span) {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		span = opentracing.StartSpan(ts.prefix + r.URL.Path)
	} else {
		span = opentracing.StartSpan(ts.prefix+r.URL.Path, opentracing.ChildOf(spanCtx))
	}

	// OpenTracing allows for arbritary tags to be added to a Span.
	span.SetTag("Request.Host", r.Host)
	span.SetTag("Request.Address", r.RemoteAddr)
	addHeaderTags(span, r.Header)

	span.SetBaggageItem("User", os.Getenv("USER"))
	return span
}

func (ts *TraceService) WrapperRequest(span opentracing.Span, req *http.Request) {
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
}
