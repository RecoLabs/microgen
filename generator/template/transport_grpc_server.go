package template

import (
	"context"
	"fmt"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/recolabs/microgen/generator/strings"
	"github.com/recolabs/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

type gRPCServerTemplate struct {
	info *GenerationInfo
}

func NewGRPCServerTemplate(info *GenerationInfo) Template {
	return &gRPCServerTemplate{
		info: info,
	}
}

func serverStructName(iface *types.Interface) string {
	return iface.Name + "Server"
}

func privateServerStructName(iface *types.Interface) string {
	return mstrings.ToLower(iface.Name) + "Server"
}

// Render whole grpc server file.
//
//		// This file was automatically generated by "microgen" utility.
//		// DO NOT EDIT.
//		package transportgrpc
//
//		import (
//			svc "github.com/recolabs/microgen/examples/svc"
//			protobuf "github.com/recolabs/microgen/examples/svc/transport/converter/protobuf"
//			grpc "github.com/go-kit/kit/transport/grpc"
//			stringsvc "gitlab.devim.team/protobuf/stringsvc"
//			context "golang.org/x/net/context"
//		)
//
//		type stringServiceServer struct {
//			count grpc.Handler
//		}
//
//		func NewGRPCServer(endpoints *svc.Endpoints, opts ...grpc.ServerOption) stringsvc.StringServiceServer {
//			return &stringServiceServer{count: grpc.NewServer(
//				endpoints.CountEndpoint,
//				protobuf.DecodeCountRequest,
//				protobuf.EncodeCountResponse,
//				opts...,
//			)}
//		}
//
//		func (s *stringServiceServer) Count(ctx context.Context, req *stringsvc.CountRequest) (*stringsvc.CountResponse, error) {
//			_, resp, err := s.count.ServeGRPC(ctx, req)
//			if err != nil {
//				return nil, err
//			}
//			return resp.(*stringsvc.CountResponse), nil
//		}
//
func (t *gRPCServerTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("transportgrpc")
	f.ImportAlias(t.info.ProtobufPackageImport, "pb")
	f.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	f.HeaderComment(t.info.FileHeader)
	f.PackageComment(`DO NOT EDIT.`)

	f.Type().Id(privateServerStructName(t.info.Iface)).StructFunc(func(g *Group) {
		unimplementedServerEmbedString := fmt.Sprintf("pb.Unimplemented%s", serverStructName(t.info.Iface))
		g.Id(unimplementedServerEmbedString)
		for _, method := range t.info.Iface.Methods {
			if t.info.OneToManyStreamMethods[method.Name] {
				g.Id(mstrings.ToLowerFirst(method.Name)).Qual(t.info.OutputPackageImport+"/transport", OneToManyStreamEndpoint)
				continue
			}
			if t.info.ManyToManyStreamMethods[method.Name] {
				g.Id(mstrings.ToLowerFirst(method.Name)).Qual(t.info.OutputPackageImport+"/transport", ManyToManyStreamEndpoint)
				continue
			}
			if t.info.ManyToOneStreamMethods[method.Name] {
				g.Id(mstrings.ToLowerFirst(method.Name)).Qual(t.info.OutputPackageImport+"/transport", ManyToOneStreamEndpoint)
				continue
			}
			if !t.info.AllowedMethods[method.Name] {
				continue
			}
			g.Id(mstrings.ToLowerFirst(method.Name)).Qual(PackagePathGoKitTransportGRPC, "Handler")
		}
	}).Line()

	f.Func().Id("NewGRPCServer").
		ParamsFunc(func(p *Group) {
			p.Id("endpoints").Op("*").Qual(t.info.OutputPackageImport+"/transport", EndpointsSetName)
			if Tags(ctx).Has(TracingMiddlewareTag) {
				p.Id("logger").Qual(PackagePathGoKitLog, "Logger")
			}
			if Tags(ctx).Has(TracingMiddlewareTag) {
				p.Id("tracer").Qual(PackagePathOpenTracingGo, "Tracer")
			}
			p.Id("opts").Op("...").Qual(PackagePathGoKitTransportGRPC, "ServerOption")
		}).Params(
		Qual(t.info.ProtobufPackageImport, serverStructName(t.info.Iface)),
	).
		Block(
			Return().Op("&").Id(privateServerStructName(t.info.Iface)).Values(DictFunc(func(g Dict) {
				for _, m := range t.info.Iface.Methods {
					if t.info.OneToManyStreamMethods[m.Name] {
						g[(&Statement{}).Id(mstrings.ToLowerFirst(m.Name))] = Line().Id("newOneToManyStreamServer").
							Call(
								Line().Id("endpoints").Dot(endpointsStructFieldName(m.Name)),
							)
						continue
					}
					if t.info.ManyToManyStreamMethods[m.Name] {
						g[(&Statement{}).Id(mstrings.ToLowerFirst(m.Name))] = Line().Id("newManyToManyStreamServer").
							Call(
								Line().Id("endpoints").Dot(endpointsStructFieldName(m.Name)),
							)
						continue
					}
					if t.info.ManyToOneStreamMethods[m.Name] {
						g[(&Statement{}).Id(mstrings.ToLowerFirst(m.Name))] = Line().Id("newManyToOneStreamServer").
							Call(
								Line().Id("endpoints").Dot(endpointsStructFieldName(m.Name)),
							)
						continue
					}
					if !t.info.AllowedMethods[m.Name] {
						continue
					}
					g[(&Statement{}).Id(mstrings.ToLowerFirst(m.Name))] = Qual(PackagePathGoKitTransportGRPC, "NewServer").
						Call(
							Line().Id("endpoints").Dot(endpointsStructFieldName(m.Name)),
							Line().Id(decodeRequestName(m)),
							Line().Id(encodeResponseName(m)),
							Line().Add(t.serverOpts(ctx, m)).Op("...").Line(),
						)
				}
			}),
			),
		)
	f.Line()

	f.Func().
		Id("newOneToManyStreamServer").
		Params(
			Id("endpoint").Qual(t.info.OutputPackageImport+"/transport", OneToManyStreamEndpoint),
		).
		Params(
			Qual(t.info.OutputPackageImport+"/transport", OneToManyStreamEndpoint),
		).
		Block(
			Return().Id("endpoint"),
		)

	f.Line()

	f.Func().
		Id("newManyToOneStreamServer").
		Params(
			Id("endpoint").Qual(t.info.OutputPackageImport+"/transport", ManyToOneStreamEndpoint),
		).
		Params(
			Qual(t.info.OutputPackageImport+"/transport", ManyToOneStreamEndpoint),
		).
		Block(
			Return().Id("endpoint"),
		)

	f.Line()

	f.Func().
		Id("newManyToManyStreamServer").
		Params(
			Id("endpoint").Qual(t.info.OutputPackageImport+"/transport", ManyToManyStreamEndpoint),
		).
		Params(
			Qual(t.info.OutputPackageImport+"/transport", ManyToManyStreamEndpoint),
		).
		Block(
			Return().Id("endpoint"),
		)
	f.Line()

	for _, signature := range t.info.Iface.Methods {
		if t.info.OneToManyStreamMethods[signature.Name] {
			f.Add(t.grpcOneToManyStreamServerFunc(signature, t.info)).Line()
			continue
		}
		if t.info.ManyToManyStreamMethods[signature.Name] {
			f.Add(t.grpcManyToManyStreamServerFunc(signature, t.info)).Line()
			continue
		}
		if t.info.ManyToOneStreamMethods[signature.Name] {
			f.Add(t.grpcManyToOneStreamServerFunc(signature, t.info)).Line()
			continue
		}
		if !t.info.AllowedMethods[signature.Name] {
			continue
		}
		f.Add(t.grpcServerFunc(signature, t.info.Iface)).Line()
	}

	return f
}

func (t *gRPCServerTemplate) grpcOneToManyStreamServerFunc(signature *types.Function, info *GenerationInfo) *Statement {
	return Func().
		Params(Id(rec(privateServerStructName(info.Iface))).Op("*").Id(privateServerStructName(info.Iface))).
		Id(signature.Name).
		Call(
			Id("req").Add(t.grpcServerReqStruct(signature)),
			Id("stream").Qual(info.ProtobufPackageImport, streamStructName(info.Iface.Name, signature))).
		Params(Error()).
		BlockFunc(t.grpcOneToManyStreamServerFuncBody(signature, info.Iface))
}
func (t *gRPCServerTemplate) grpcOneToManyStreamServerFuncBody(signature *types.Function, i *types.Interface) func(g *Group) {
	return func(g *Group) {
		g.List(Id("decoded_req"), Err()).
			Op(":=").
			Id(decodeRequestName(signature)).Call(
			Id("context.Background()"),
			Id("req"))

		g.If(Err().Op("!=").Nil()).Block(
			Return().List(Err()),
		)

		g.Return().Id(rec(privateServerStructName(i))).Dot(mstrings.ToLowerFirst(signature.Name)).Call(
			Id("decoded_req"),
			Id("stream"),
		)
	}
}
func (t *gRPCServerTemplate) grpcManyToManyStreamServerFunc(signature *types.Function, info *GenerationInfo) *Statement {
	return t.grpcManyToOneStreamServerFunc(signature, info)
}
func (t *gRPCServerTemplate) grpcManyToOneStreamServerFunc(signature *types.Function, info *GenerationInfo) *Statement {

	return Func().
		Params(Id(rec(privateServerStructName(info.Iface))).Op("*").Id(privateServerStructName(info.Iface))).
		Id(signature.Name).
		Call(Id("stream").Qual(info.ProtobufPackageImport, streamStructName(info.Iface.Name, signature))).
		Params(Error()).
		BlockFunc(t.grpcManyToOneStreamServerFuncBody(signature, info.Iface))
}
func (t *gRPCServerTemplate) grpcManyToOneStreamServerFuncBody(signature *types.Function, i *types.Interface) func(g *Group) {
	return func(g *Group) {
		g.Return().Id(rec(privateServerStructName(i))).Dot(mstrings.ToLowerFirst(signature.Name)).Call(Id("stream"))
	}
}

func (gRPCServerTemplate) DefaultPath() string {
	return filenameBuilder(PathTransport, "grpc", "server")
}

func (t *gRPCServerTemplate) Prepare(ctx context.Context) error {
	if t.info.ProtobufPackageImport == "" {
		return ErrProtobufEmpty
	}
	return nil
}

func (t *gRPCServerTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

// Render service interface method for grpc server.
//
//		func (s *stringServiceServer) Count(ctx context.Context, req *stringsvc.CountRequest) (*stringsvc.CountResponse, error) {
//			_, resp, err := s.count.ServeGRPC(ctx, req)
//			if err != nil {
//				return nil, err
//			}
//			return resp.(*stringsvc.CountResponse), nil
//		}
//
func (t *gRPCServerTemplate) grpcServerFunc(signature *types.Function, i *types.Interface) *Statement {
	return Func().
		Params(Id(rec(privateServerStructName(i))).Op("*").Id(privateServerStructName(i))).
		Id(signature.Name).
		Call(Id("ctx").Qual(PackagePathNetContext, "Context"), Id("req").Add(t.grpcServerReqStruct(signature))).
		Params(t.grpcServerRespStruct(signature), Error()).
		BlockFunc(t.grpcServerFuncBody(signature, i))
}

// Special case for empty request
// Render
//		*empty.Empty
// or
//		*stringsvc.CountRequest
func (t *gRPCServerTemplate) grpcServerReqStruct(fn *types.Function) *Statement {
	args := RemoveContextIfFirst(fn.Args)
	if len(args) == 0 {
		return Op("*").Qual(PackagePathEmptyProtobuf, "Empty")
	}
	if len(args) == 1 {
		sp := specialTypeConverter(args[0].Type)
		if sp != nil {
			return sp
		}
	}
	return Op("*").Qual(t.info.ProtobufPackageImport, requestStructName(fn))
}

// Special case for empty response
// Render
//		*empty.Empty
// or
//		*stringsvc.CountResponse
func (t *gRPCServerTemplate) grpcServerRespStruct(fn *types.Function) *Statement {
	results := removeErrorIfLast(fn.Results)
	if len(results) == 0 {
		return Op("*").Qual(PackagePathEmptyProtobuf, "Empty")
	}
	if len(results) == 1 {
		sp := specialTypeConverter(results[0].Type)
		if sp != nil {
			return sp
		}
	}
	return Op("*").Qual(t.info.ProtobufPackageImport, responseStructName(fn))
}

// Render service method body for grpc server.
//
//		_, resp, err := s.count.ServeGRPC(ctx, req)
//		if err != nil {
//			return nil, err
//		}
//		return resp.(*stringsvc.CountResponse), nil
//
func (t *gRPCServerTemplate) grpcServerFuncBody(signature *types.Function, i *types.Interface) func(g *Group) {
	return func(g *Group) {
		g.List(Id("_"), Id("resp"), Err()).
			Op(":=").
			Id(rec(privateServerStructName(i))).Dot(mstrings.ToLowerFirst(signature.Name)).Dot("ServeGRPC").Call(Id("ctx"), Id("req"))

		g.If(Err().Op("!=").Nil()).Block(
			Return().List(Nil(), Err()),
		)

		g.Return().List(Id("resp").Assert(t.grpcServerRespStruct(signature)), Nil())
	}
}

func (t *gRPCServerTemplate) serverOpts(ctx context.Context, fn *types.Function) *Statement {
	s := &Statement{}
	if Tags(ctx).Has(TracingMiddlewareTag) {
		s.Op("append(")
		defer s.Op(")")
	}
	s.Id("opts")
	if Tags(ctx).Has(TracingMiddlewareTag) {
		s.Op(",").Qual(PackagePathGoKitTransportGRPC, "ServerBefore").Call(
			Line().Qual(PackagePathGoKitTracing, "GRPCToContext").Call(Id("tracer"), Lit(fn.Name), Id("logger")),
		)
	}
	return s
}
