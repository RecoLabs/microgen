package template

import (
	"context"

	"github.com/devimteam/microgen/internal"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/strings"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

type endpointsClientTemplate struct {
	info *GenerationInfo
}

func NewEndpointsClientTemplate(info *GenerationInfo) Template {
	return &endpointsClientTemplate{
		info: info,
	}
}

// Renders endpoints file.
//
//		// This file was automatically generated by "microgen" utility.
//		// DO NOT EDIT.
//		package stringsvc
//
//		import (
//			context "context"
//			endpoint "github.com/go-kit/kit/endpoint"
//		)
//
//		type Endpoints struct {
//			CountEndpoint endpoint.Endpoint
//		}
//
//		func (e *Endpoints) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			req := CountRequest{
//				Symbol: symbol,
//				Text:   text,
//			}
//			resp, err := e.CountEndpoint(ctx, &req)
//			if err != nil {
//				return
//			}
//			return resp.(*CountResponse).Count, resp.(*CountResponse).Positions
//		}
//
//		func CountEndpoint(svc StringService) endpoint.Endpoint {
//			return func(ctx context.Context, request interface{}) (interface{}, error) {
//				req := request.(*CountRequest)
//				count, positions := svc.Count(ctx, req.Text, req.Symbol)
//				return &CountResponse{
//					Count:     count,
//					Positions: positions,
//				}, nil
//			}
//		}
//
func (t *endpointsClientTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("transport")
	f.HeaderComment(t.info.FileHeader)
	if internal.Tags(ctx).HasAny(TracingMiddlewareTag) {
		f.Comment("TraceClientEndpoints is used for tracing endpoints on client side.")
		f.Add(t.clientTracingMiddleware()).Line()
	}
	for _, signature := range t.info.Iface.Methods {
		f.Add(t.serviceEndpointMethod(ctx, signature)).Line().Line()
	}
	return f
}

func (endpointsClientTemplate) DefaultPath() string {
	return filenameBuilder(PathTransport, "client")
}

func (t *endpointsClientTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *endpointsClientTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

// Render full endpoints method.
//
//		func (e *Endpoints) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			req := CountRequest{
//				Symbol: symbol,
//				Text:   text,
//			}
//			resp, err := e.CountEndpoint(ctx, &req)
//			if err != nil {
//				return
//			}
//			return resp.(*CountResponse).Count, resp.(*CountResponse).Positions
//		}
//
func (t *endpointsClientTemplate) serviceEndpointMethod(ctx context.Context, signature *types.Function) *Statement {
	normal := normalizeFunction(signature)
	return methodDefinitionFull(ctx, EndpointsSetName, &normal.Function).
		BlockFunc(t.serviceEndpointMethodBody(ctx, signature, &normal.Function))
}

// Render interface method body.
//
//		endpointCountRequest := CountRequest{
//			Symbol: symbol,
//			Text:   text,
//		}
//		endpointCountResponse, err := E.CountEndpoint(ctx, &endpointCountRequest)
//		if err != nil {
//			return
//		}
//		return endpointCountResponse.(*CountResponse).Count, endpointCountResponse.(*CountResponse).Positions, err
//
func (t *endpointsClientTemplate) serviceEndpointMethodBody(ctx context.Context, fn *types.Function, normal *types.Function) func(g *Group) {
	reqName := "request"
	respName := "response"
	return func(g *Group) {
		if !t.info.AllowedMethods[fn.Name] {
			g.Return()
			return
		}
		g.Id(reqName).Op(":=").Id(requestStructName(fn)).Values(dictByNormalVariables(RemoveContextIfFirst(fn.Args), RemoveContextIfFirst(normal.Args)))
		g.Add(endpointResponse(respName, normal)).Id(strings.LastWordFromName(EndpointsSetName)).Dot(endpointsStructFieldName(fn.Name)).Call(Id(firstArgName(normal)), Op("&").Id(reqName))
		g.If(Id(nameOfLastResultError(normal)).Op("!=").Nil().BlockFunc(func(ifg *Group) {
			if internal.Tags(ctx).HasAny(GrpcTag, GrpcClientTag, GrpcServerTag) {
				ifg.Add(checkGRPCError(normal))
			}
			ifg.Return()
		}))
		g.ReturnFunc(func(group *Group) {
			for _, field := range removeErrorIfLast(fn.Results) {
				group.Id(respName).Assert(Op("*").Id(responseStructName(fn))).Op(".").Add(structFieldName(&field))
			}
			group.Id(nameOfLastResultError(normal))
		})
	}
}

func checkGRPCError(fn *types.Function) *Statement {
	s := &Statement{}
	s.If(List(Id("e"), Id("ok")).Op(":=").Qual(PackagePathGoogleGRPCStatus, "FromError").Call(Id(nameOfLastResultError(fn))),
		Id("ok").Op("||").
			Id("e").Dot("Code").Call().Op("==").Qual(PackagePathGoogleGRPCCodes, "Internal").Op("||").
			Id("e").Dot("Code").Call().Op("==").Qual(PackagePathGoogleGRPCCodes, "Unknown"),
	).Block(
		Id(nameOfLastResultError(fn)).Op("=").Qual("errors", "New").Call(Id("e").Dot("Message").Call()),
	)
	return s
}

// Helper func for `serviceEndpointMethodBody`
func endpointResponse(respName string, fn *types.Function) *Statement {
	if len(removeErrorIfLast(fn.Results)) > 0 {
		return List(Id(respName), Id(nameOfLastResultError(fn))).Op(":=")
	}
	return List(Id("_"), Id(nameOfLastResultError(fn))).Op("=")
}

// For custom ctx in service interface (e.g. context or ctxxx).
func firstArgName(signature *types.Function) string {
	return mstrings.ToLowerFirst(signature.Args[0].Name)
}

func (t *endpointsClientTemplate) clientTracingMiddleware() *Statement {
	s := &Statement{}
	s.Func().Id("TraceClientEndpoints").Call(Id("endpoints").Id(EndpointsSetName), Id("tracer").Qual(PackagePathOpenTracingGo, "Tracer")).Id(EndpointsSetName).BlockFunc(func(g *Group) {
		g.Return(Id(EndpointsSetName).Values(DictFunc(func(d Dict) {
			for _, signature := range t.info.Iface.Methods {
				if t.info.AllowedMethods[signature.Name] {
					d[Id(endpointsStructFieldName(signature.Name))] = Qual(PackagePathGoKitTracing, "TraceClient").Call(Id("tracer"), Lit(signature.Name)).Call(Id("endpoints").Dot(endpointsStructFieldName(signature.Name)))
				}
			}
		})))
	})
	return s
}
