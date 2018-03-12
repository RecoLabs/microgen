// This file was automatically generated by "microgen 0.8.0" utility.
// Please, do not edit.
package transporthttp

import (
	generated "github.com/devimteam/microgen/example/generated"
	http2 "github.com/devimteam/microgen/example/generated/transport/converter/http"
	log "github.com/go-kit/kit/log"
	opentracing "github.com/go-kit/kit/tracing/opentracing"
	http "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
	opentracinggo "github.com/opentracing/opentracing-go"
	http1 "net/http"
)

func NewHTTPHandler(endpoints *generated.Endpoints, logger log.Logger, tracer opentracinggo.Tracer, opts ...http.ServerOption) http1.Handler {
	mux := mux.NewRouter()
	mux.Methods("POST").Path("/uppercase").Handler(
		http.NewServer(
			endpoints.UppercaseEndpoint,
			http2.DecodeHTTPUppercaseRequest,
			http2.EncodeHTTPUppercaseResponse,
			append(opts, http.ServerBefore(
				opentracing.HTTPToContext(tracer, "Uppercase", logger)))...))
	mux.Methods("GET").Path("/count/{text}/{symbol}").Handler(
		http.NewServer(
			endpoints.CountEndpoint,
			http2.DecodeHTTPCountRequest,
			http2.EncodeHTTPCountResponse,
			append(opts, http.ServerBefore(
				opentracing.HTTPToContext(tracer, "Count", logger)))...))
	mux.Methods("POST").Path("/test-case").Handler(
		http.NewServer(
			endpoints.TestCaseEndpoint,
			http2.DecodeHTTPTestCaseRequest,
			http2.EncodeHTTPTestCaseResponse,
			append(opts, http.ServerBefore(
				opentracing.HTTPToContext(tracer, "TestCase", logger)))...))
	return mux
}
