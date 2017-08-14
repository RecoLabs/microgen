package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

var (
	defaultProtoTypes = []string{"string", "bool", "byte", "int64", "uint64", "float64", "int32", "uint32", "float32"}
	goToProtoTypesMap = map[string]string{
		"uint": "uint64",
		"int":  "int64",
	}
	defaultGolangTypes = []string{"string", "bool", "int", "uint", "byte", "int64", "uint64", "float64", "int32", "uint32", "float32"}
)

type GRPCConverterTemplate struct {
	PackagePath string
}

func utilPackagePath(path string) string {
	return path + "/util"
}

func (t GRPCClientTemplate) converterPackagePath() string {
	return t.PackagePath + "/transport/converter/protobuf"
}

// Renders converter file.
//
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package transportgrpc
//
//		import (
//			context "context"
//			grpc "github.com/devimteam/go-kit/transportlayer/grpc"
//			svc "github.com/devimteam/microgen/test/svc"
//			util "github.com/devimteam/microgen/test/svc/util"
//			stringsvc "gitlab.devim.team/protobuf/stringsvc"
//		)
//
//		var CountConverter = &grpc.EndpointConverter{
//			func(_ context.Context, request interface{}) (interface{}, error) {
//				req := request.(*svc.CountRequest)
//				return &stringsvc.CountRequest{
//					Symbol: req.Symbol,
//					Text:   req.Text,
//				}, nil
//			},
//			func(_ context.Context, response interface{}) (interface{}, error) {
//				resp := response.(*svc.CountResponse)
//				return &stringsvc.CountResponse{
//					Count:     int64(resp.Count),
//					Positions: util.PositionsToProto(resp.Positions),
//				}, nil
//			},
//			func(_ context.Context, request interface{}) (interface{}, error) {
//				req := request.(*stringsvc.CountRequest)
//				return &svc.CountRequest{
//					Symbol: string(req.Symbol),
//					Text:   string(req.Text),
//				}, nil
//			},
//			func(_ context.Context, response interface{}) (interface{}, error) {
//				resp := response.(*stringsvc.CountResponse)
//				return &svc.CountResponse{
//					Count:     int(resp.Count),
//					Positions: util.ProtoToPositions(resp.Positions),
//				}, nil
//			},
//			stringsvc.CountResponse{},
//		}
//
func (t GRPCConverterTemplate) Render(i *parser.Interface) *File {
	f := NewFile("transportgrpc")

	for _, signature := range i.FuncSignatures {
		f.Var().Id(converterStructName(signature)).Op("=").Op("&").Qual(PackagePathTransportLayerGRPC, "EndpointConverter").
			ValuesFunc(func(g *Group) {
				g.Add(t.encodeRequest(signature, i))
				g.Add(t.encodeResponse(signature, i))
				g.Add(t.decodeRequest(signature, i))
				g.Add(t.decodeResponse(signature, i))
				g.Add(t.replyType(signature, i))
				g.Line()
			})
		f.Line()
	}

	return f
}

// Returns NameToProto.
func nameToProto(name string) string {
	return name + "ToProto"
}

// Returns ProtoToName.
func protoToName(name string) string {
	return "ProtoTo" + name
}

func (GRPCConverterTemplate) Path() string {
	return "./transport/grpc/converter.go"
}

// Renders type conversion (if need) to default protobuf types.
//		req.Symbol
// or
//		int(resp.Count)
// or nothing
func defaultGolangTypeToProto(structName string, field *parser.FuncField) (*Statement, bool) {
	if isDefaultProtoField(field) {
		return Id(structName).Dot(util.ToUpperFirst(field.Name)), false
	} else if field.IsArray || field.IsPointer {
		return Add(), true
	}
	if newType, ok := goToProtoTypesMap[field.Type]; ok {
		newField := &parser.FuncField{
			Type:      newType,
			Name:      field.Name,
			IsArray:   field.IsArray,
			Package:   field.Package,
			IsPointer: field.IsPointer,
		}
		return fieldType(newField).Call(Id(structName).Dot(util.ToUpperFirst(field.Name))), false
	}
	return Add(), true
}

// Renders type conversion to default golang types.
// 		int(resp.Count)
// or nothing
func defaultProtoTypeToGolang(object string, field *parser.FuncField) (*Statement, bool) {
	if isDefaultGolangField(field) {
		return fieldType(field).Call(Id(object).Dot(util.ToUpperFirst(field.Name))), false
	}
	return Add(), true
}

func isDefaultProtoField(field *parser.FuncField) bool {
	if field.Type == "byte" && field.IsArray {
		return true
	} else if field.IsArray || field.IsPointer {
		return false
	}
	return util.IsInStringSlice(field.Type, defaultProtoTypes)
}

func isDefaultGolangField(field *parser.FuncField) bool {
	if field.Type == "byte" && field.IsArray {
		return true
	} else if field.IsArray || field.IsPointer {
		return false
	}
	return util.IsInStringSlice(field.Type, defaultGolangTypes)
}

// Renders function for encoding request, golang type converts to proto type.
//
//		func(_ context.Context, request interface{}) (interface{}, error) {
//			req := request.(*svc.CountRequest)
//			return &stringsvc.CountRequest{
//				Symbol: req.Symbol,
//				Text:   req.Text,
//			}, nil
//		}
//
func (t GRPCConverterTemplate) encodeRequest(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	methodParams := removeContextIfFirst(signature.Params)
	return Line().Func().Call(Op("_").Qual(PackagePathContext, "Context"), Id("request").Interface()).Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			if len(methodParams) > 0 {
				group.Id("req").Op(":=").Id("request").Assert(Op("*").Qual(t.PackagePath, requestStructName(signature)))
			}
			group.Return().List(Op("&").Qual(protobufPath(i), requestStructName(signature)).Values(DictFunc(func(dict Dict) {
				for _, field := range methodParams {
					code, isCustom := defaultGolangTypeToProto("req", field)
					if isCustom {
						if field.Type == "error" {
							code = Qual(utilPackagePath(t.PackagePath), "ErrorToString").Call(Id("req").Dot(util.ToUpperFirst(field.Name)))
						} else {
							code = Qual(utilPackagePath(t.PackagePath), nameToProto(util.ToUpperFirst(field.Name))).
								Call(Id("req").
									Dot(util.ToUpperFirst(field.Name)))
						}
					}
					dict[structFieldName(field)] = Line().Add(code)
				}
			})), Nil())
		},
	)
}

// Renders function for encoding response, golang type converts to proto type.
//
//		func(_ context.Context, response interface{}) (interface{}, error) {
//			resp := response.(*svc.CountResponse)
//			return &stringsvc.CountResponse{
//				Count:     int64(resp.Count),
//				Positions: []int64(resp.Positions),
//			}, nil
//		}
//
func (t GRPCConverterTemplate) encodeResponse(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	methodResults := removeContextIfFirst(signature.Results)
	return Line().Func().Call(Op("_").Qual(PackagePathContext, "Context"), Id("response").Interface()).Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			if len(methodResults) > 0 {
				group.Id("resp").Op(":=").Id("response").Assert(Op("*").Qual(t.PackagePath, responseStructName(signature)))
			}
			group.Return().List(Op("&").Qual(protobufPath(i), responseStructName(signature)).Values(DictFunc(func(dict Dict) {
				for _, field := range methodResults {
					code, isCustom := defaultGolangTypeToProto("resp", field)
					if isCustom {
						if field.Type == "error" {
							code = Qual(utilPackagePath(t.PackagePath), "ErrorToString").Call(Id("resp").Dot(util.ToUpperFirst(field.Name)))
						} else {
							code = Qual(utilPackagePath(t.PackagePath), nameToProto(util.ToUpperFirst(field.Name))).
								Call(Id("resp").
									Dot(util.ToUpperFirst(field.Name)))
						}
					}
					dict[structFieldName(field)] = Line().Add(code)
				}
			})), Nil())
		},
	)
}

// Renders function for decoding request, proto type converts to golang type.
//
//		func(_ context.Context, request interface{}) (interface{}, error) {
//			req := request.(*stringsvc.CountRequest)
//			return &svc.CountRequest{
//				Symbol: string(req.Symbol),
//				Text:   string(req.Text),
//			}, nil
//		},
//
func (t GRPCConverterTemplate) decodeRequest(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	methodParams := removeContextIfFirst(signature.Params)
	return Line().Func().Call(Op("_").Qual(PackagePathContext, "Context"), Id("request").Interface()).Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			if len(methodParams) > 0 {
				group.Id("req").Op(":=").Id("request").Assert(Op("*").Qual(protobufPath(i), requestStructName(signature)))
			}
			group.Return().List(Op("&").Qual(t.PackagePath, requestStructName(signature)).Values(DictFunc(func(dict Dict) {
				for _, field := range methodParams {
					code, isCustom := defaultProtoTypeToGolang("req", field)
					if isCustom {
						if field.Type == "error" {
							code = Qual(utilPackagePath(t.PackagePath), "StringToError").Call(Id("req").Dot(util.ToUpperFirst(field.Name)))
						} else {
							code = Qual(utilPackagePath(t.PackagePath), protoToName(util.ToUpperFirst(field.Name))).
								Call(Id("req").
									Dot(util.ToUpperFirst(field.Name)))
						}
					}
					dict[structFieldName(field)] = Line().Add(code)
				}
			})), Nil())
		},
	)
}

// Renders function for decoding response, proto type converts to golang type.
//
//		func(_ context.Context, response interface{}) (interface{}, error) {
//			resp := response.(*stringsvc.CountResponse)
//			return &svc.CountResponse{
//				Count:     int(resp.Count),
//				Positions: []int(resp.Positions),
//			}, nil
//		}
//
func (t GRPCConverterTemplate) decodeResponse(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	methodResults := removeContextIfFirst(signature.Results)
	return Line().Func().Call(Op("_").Qual(PackagePathContext, "Context"), Id("response").Interface()).Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			if len(methodResults) > 0 {
				group.Id("resp").Op(":=").Id("response").Assert(Op("*").Qual(protobufPath(i), responseStructName(signature)))
			}
			group.Return().List(Op("&").Qual(t.PackagePath, responseStructName(signature)).Values(DictFunc(func(dict Dict) {
				for _, field := range methodResults {
					code, isCustom := defaultProtoTypeToGolang("resp", field)
					if isCustom {
						if field.Type == "error" {
							code = Qual(utilPackagePath(t.PackagePath), "StringToError").Call(Id("resp").Dot(util.ToUpperFirst(field.Name)))
						} else {
							code = Qual(utilPackagePath(t.PackagePath), protoToName(util.ToUpperFirst(field.Name))).
								Call(Id("resp").
									Dot(util.ToUpperFirst(field.Name)))
						}
					}
					dict[structFieldName(field)] = Line().Add(code)
				}
			})), Nil())
		},
	)
}

// Renders reply type argument
// 		stringsvc.CountResponse{},
func (t GRPCConverterTemplate) replyType(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	return Line().Qual(protobufPath(i), responseStructName(signature)).Values()
}
