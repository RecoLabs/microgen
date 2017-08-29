package template

import (
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

type GRPCClientTemplate struct {
	PackagePath string
}

func (t GRPCClientTemplate) grpcConverterPackagePath() string {
	return t.PackagePath + "/transport/converter/protobuf"
}

// Render whole grpc client file.
//
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package transportgrpc
//
//		import (
//			transportlayer "github.com/devimteam/go-kit/transportlayer"
//			grpc1 "github.com/devimteam/go-kit/transportlayer/grpc"
//			svc "github.com/devimteam/microgen/test/svc"
//			protobuf "github.com/devimteam/microgen/test/svc/transport/converter/protobuf"
//			grpc "google.golang.org/grpc"
//		)
//
//		func NewClient(conn *grpc.ClientConn) svc.StringService {
//			endpoints := []transportlayer.Endpoint{
//				transportlayer.NewEndpoint(
//					"Count",
//					nil,
//					transportlayer.WithConverter(protobuf.CountConverter),
// 				),
// 			}
//			return svc.NewClient(
//				grpc1.NewClient(
//					"devim.string.protobuf.StringService",
//					conn,
//					endpoints,
// 				),
// 			)
//		}
//
func (t GRPCClientTemplate) Render(i *parser.Interface) *File {
	f := NewFile("transportgrpc")

	f.Func().Id("NewClient").
		Call(Id("conn").
			Op("*").Qual(PackagePathGoogleGRPC, "ClientConn")).Qual(t.PackagePath, i.Name).
		BlockFunc(func(g *Group) {
			g.Id("endpoints").Op(":=").Index().Qual(PackagePathTransportLayer, "Endpoint").ValuesFunc(func(group *Group) {
				for _, signature := range i.FuncSignatures {
					group.Line().Qual(PackagePathTransportLayer, "NewEndpoint").Call(
						Line().Lit(signature.Name),
						Line().Nil(),
						Line().Qual(PackagePathTransportLayer, "WithConverter").Call(Qual(t.grpcConverterPackagePath(), converterStructName(signature))),
						Line(),
					)
				}
				group.Line()
			})
			g.Return().Qual(t.PackagePath, "NewClient").Call(
				Line().Qual(PackagePathTransportLayerGRPC, "NewClient").Call(
					Line().Lit("devim."+strings.ToLower(strings.TrimSuffix(i.Name, "Service"))+".protobuf."+i.Name),
					Line().Id("conn"),
					Line().Id("endpoints"),
					Line(),
				),
				Line(),
			)
		})
	return f
}

func (GRPCClientTemplate) Path() string {
	return "./transport/grpc/client.go"
}
