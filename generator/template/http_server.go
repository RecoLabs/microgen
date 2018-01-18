package template

import (
	"path/filepath"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
)

const (
	httpMethodTag = "http-method"
)

type httpServerTemplate struct {
	Info *GenerationInfo
}

func NewHttpServerTemplate(info *GenerationInfo) Template {
	return &httpServerTemplate{
		Info: info.Copy(),
	}
}

func (t *httpServerTemplate) DefaultPath() string {
	return "./transport/http/server.go"
}

func (t *httpServerTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	if err := util.StatFile(t.Info.AbsOutPath, t.DefaultPath()); !t.Info.Force && err == nil {
		return nil, nil
	}
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *httpServerTemplate) Prepare() error {
	tags := util.FetchTags(t.Info.Iface.Docs, TagMark+ForceTag)
	if util.IsInStringSlice("http", tags) || util.IsInStringSlice("http-server", tags) {
		t.Info.Force = true
	}
	return nil
}

// Render http server constructor.
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package transporthttp
//
//		import (
//			svc "github.com/devimteam/microgen/example/svc"
//			http2 "github.com/devimteam/microgen/example/svc/transport/converter/http"
//			http "github.com/go-kit/kit/transport/http"
//			http1 "net/http"
//		)
//
//		func NewHTTPHandler(endpoints *svc.Endpoints, opts ...http.ServerOption) http1.Handler {
//			handler := http1.NewServeMux()
//			handler.Handle("/test_case", http.NewServer(
//				endpoints.TestCaseEndpoint,
//				http2.DecodeHTTPTestCaseRequest,
//				http2.EncodeHTTPTestCaseResponse,
//				opts...))
//			handler.Handle("/empty_req", http.NewServer(
//				endpoints.EmptyReqEndpoint,
//				http2.DecodeHTTPEmptyReqRequest,
//				http2.EncodeHTTPEmptyReqResponse,
//				opts...))
//			handler.Handle("/empty_resp", http.NewServer(
//				endpoints.EmptyRespEndpoint,
//				http2.DecodeHTTPEmptyRespRequest,
//				http2.EncodeHTTPEmptyRespResponse,
//				opts...))
//			return handler
//		}
//

/*

func MakeHTTPHandler(s EchoService, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	r.Methods("POST").Path("/echo").Handler(httptransport.NewServer(
		makeEchoEndpoint(s),
		decodeEchoRequest,
		encodeResponse,
		httptransport.ServerErrorLogger(logger),
	))
	return r
}
*/

func (t *httpServerTemplate) Render() write_strategy.Renderer {
	f := NewFile("transporthttp")
	f.PackageComment(FileHeader)
	f.PackageComment(`Please, do not edit.`)

	f.Func().Id("NewHTTPHandler").Params(
		Id("endpoints").Op("*").Qual(t.Info.ServiceImportPath, "Endpoints"),
		Id("opts").Op("...").Qual(PackagePathGoKitTransportHTTP, "ServerOption"),
	).Params(
		Qual(PackagePathHttp, "Handler"),
	).BlockFunc(func(g *Group) {
		g.Id("mux").Op(":=").Qual(PackageGorillaMux, "NewRouter").Call()
		for _, fn := range t.Info.Iface.Methods {
			tags := util.FetchTags(fn.Docs, TagMark+httpMethodTag)
			tag := ""
			if len(tags) == 1 {
				tag = strings.ToUpper(tags[0])
			} else {
				tag = "GET"
			}
			g.Id("mux").Dot("Methods").Call(Lit(tag)).Dot("Path").
				Call(Lit("/" + util.ToURLSnakeCase(fn.Name))).Dot("Handler").Call(
				Qual(PackagePathGoKitTransportHTTP, "NewServer").Call(
					Line().Id("endpoints").Dot(endpointStructName(fn.Name)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), httpDecodeRequestName(fn)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), httpEncodeResponseName(fn)),
					Line().Id("opts").Op("...")),
			)
		}
		g.Return(Id("mux"))
	})

	return f
}

func pathToHttpConverter(servicePath string) string {
	return filepath.Join(servicePath, "transport/converter/http")
}
