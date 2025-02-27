package template

import (
	"context"
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/recolabs/microgen/generator/strings"
	"github.com/recolabs/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

const (
	commonHTTPResponseEncoderName = "CommonHTTPResponseEncoder"
	commonHTTPRequestEncoderName  = "CommonHTTPRequestEncoder"
)

type httpConverterTemplate struct {
	info                         *GenerationInfo
	encodersRequest              []*types.Function
	decodersRequest              []*types.Function
	encodersResponse             []*types.Function
	decodersResponse             []*types.Function
	state                        WriteStrategyState
	isCommonEncoderRequestExist  bool
	isCommonEncoderResponseExist bool
}

func NewHttpConverterTemplate(info *GenerationInfo) Template {
	return &httpConverterTemplate{
		info: info,
	}
}

func (t *httpConverterTemplate) DefaultPath() string {
	return filenameBuilder(PathTransport, "http", "converters")
}

func (t *httpConverterTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	if err := statFile(t.info.OutputFilePath, t.DefaultPath()); err != nil {
		t.state = FileStrat
		return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
	}
	file, err := parsePackage(filepath.Join(t.info.OutputFilePath, t.DefaultPath()))
	if err != nil {
		return nil, err
	}

	removeAlreadyExistingFunctions(file.Functions, &t.encodersRequest, encodeRequestName)
	removeAlreadyExistingFunctions(file.Functions, &t.decodersRequest, decodeRequestName)
	removeAlreadyExistingFunctions(file.Functions, &t.encodersResponse, encodeResponseName)
	removeAlreadyExistingFunctions(file.Functions, &t.decodersResponse, decodeResponseName)

	for i := range file.Functions {
		if file.Functions[i].Name == commonHTTPResponseEncoderName {
			t.isCommonEncoderResponseExist = true
			continue
		}
		if file.Functions[i].Name == commonHTTPRequestEncoderName {
			t.isCommonEncoderRequestExist = true
			continue
		}
		if t.isCommonEncoderRequestExist && t.isCommonEncoderResponseExist {
			break
		}
	}

	t.state = AppendStrat
	return write_strategy.NewAppendToFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

func (t *httpConverterTemplate) Prepare(ctx context.Context) error {
	for _, fn := range t.info.Iface.Methods {
		if !t.info.AllowedMethods[fn.Name] {
			continue
		}
		t.decodersRequest = append(t.decodersRequest, fn)
		t.encodersRequest = append(t.encodersRequest, fn)
		t.decodersResponse = append(t.decodersResponse, fn)
		t.encodersResponse = append(t.encodersResponse, fn)
	}
	return nil
}

// Render http converters: for exchanges and common.
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not change functions names!
//		package httpconv
//
//		import (
//			bytes "bytes"
//			context "context"
//			json "encoding/json"
//			svc "github.com/recolabs/microgen/examples/svc"
//			ioutil "io/ioutil"
//			http "net/http"
//		)
//
//		func DefaultRequestEncoder(_ context.Context, r *http.Request, request interface{}) error {
//			var buf bytes.Buffer
//			if err := json.NewEncoder(&buf).Encode(request); err != nil {
//				return err
//			}
//			r.Body = iomstrings.NopCloser(&buf)
//			return nil
//		}
//
//		func DefaultResponseEncoder(_ context.Context, w http.ResponseWriter, response interface{}) error {
//			w.Header().Set("Content-Type", "application/json; charset=utf-8")
//			return json.NewEncoder(w).Encode(response)
//		}
//
//		func DecodeHTTPCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
//			var req svc.CountRequest
//			err := json.NewDecoder(r.Body).Decode(&req)
//			return req, err
//		}
//
//		func DecodeHTTPCountResponse(_ context.Context, r *http.Response) (interface{}, error) {
//			var resp svc.CountResponse
//			err := json.NewDecoder(r.Body).Decode(&resp)
//			return resp, err
//		}
//
//		func EncodeHTTPCountRequest(ctx context.Context, r *http.Request, request interface{}) error {
//			return DefaultRequestEncoder(ctx, r, request)
//		}
//
//		func EncodeHTTPCountResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
//			return DefaultResponseEncoder(ctx, w, response)
//		}
//
func (t *httpConverterTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := &Statement{}

	if !t.isCommonEncoderRequestExist {
		f.Line().Add(commonHTTPRequestEncoder()).Line()
	}
	if !t.isCommonEncoderResponseExist {
		f.Line().Add(commonHTTPResponseEncoder()).Line()
	}

	for _, fn := range t.decodersRequest {
		f.Line().Add(t.decodeHTTPRequest(fn)).Line()
	}
	for _, fn := range t.decodersResponse {
		f.Line().Add(t.decodeHTTPResponse(fn)).Line()
	}
	for _, fn := range t.encodersRequest {
		f.Line().Add(t.encodeHTTPRequest(fn)).Line()
	}
	for _, fn := range t.encodersResponse {
		f.Line().Add(encodeHTTPResponse(fn)).Line()
	}

	if t.state == AppendStrat {
		return f
	}

	file := NewFile("transporthttp")
	file.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	file.HeaderComment(t.info.FileHeader)
	file.PackageComment(`Please, do not change functions names!`)
	file.Add(f)

	return file
}

// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L201
func commonHTTPRequestEncoder() *Statement {
	return Func().Id(commonHTTPRequestEncoderName).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("r").Op("*").Qual(PackagePathHttp, "Request"),
			Id("request").Interface(),
		).Params(
		Error(),
	).BlockFunc(func(g *Group) {
		g.Var().Id("buf").Qual(PackagePathBytes, "Buffer")
		g.If(
			Err().Op(":=").Qual(PackagePathJson, "NewEncoder").Call(Op("&").Id("buf")).Dot("Encode").Call(Id("request")),
			Err().Op("!=").Nil(),
		).Block(
			Return(Err()),
		)
		g.Id("r").Dot("Body").Op("=").Qual(PackagePathIOUtil, "NopCloser").Call(Op("&").Id("buf"))
		g.Return(Nil())
	})
}

// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L212
func commonHTTPResponseEncoder() *Statement {
	return Func().Id(commonHTTPResponseEncoderName).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("w").Qual(PackagePathHttp, "ResponseWriter"),
			Id("response").Interface(),
		).Params(
		Error(),
	).BlockFunc(func(g *Group) {
		g.Id("w").Dot("Header").Call().Dot("Set").Call(Lit("Content-Type"), Lit("application/json; charset=utf-8"))
		g.Return(
			Qual(PackagePathJson, "NewEncoder").Call(Id("w")).Dot("Encode").Call(Id("response")),
		)
	})
}

//		func DecodeHTTPCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
//			var req svc.CountRequest
//			err := json.NewDecoder(r.Body).Decode(&req)
//			return req, err
//		}
func (t *httpConverterTemplate) decodeHTTPRequest(fn *types.Function) *Statement {
	return Func().Id(decodeRequestName(fn)).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("r").Op("*").Qual(PackagePathHttp, "Request"),
		).Params(
		Interface(),
		Error(),
	).BlockFunc(func(g *Group) {
		arguments := RemoveContextIfFirst(fn.Args)
		if FetchHttpMethodTag(fn.Docs) == "GET" {
			if len(arguments) > 0 {
				g.Var().Call(Id("_param").String())
				g.Var().Id("ok").Bool()
				g.Id("_vars").Op(":=").Qual(PackagePathGorillaMux, "Vars").Call(Id("r"))
				for _, arg := range arguments {
					g.List(Id("_param"), Id("ok")).Op("=").Id("_vars").Index(Lit(arg.Name)).
						Line().If(Op("!").Id("ok")).Block(
						Return(Nil(), Qual(PackagePathErrors, "New").Call(Lit("param "+arg.Name+" not found"))),
					)
					g.Add(stringToTypeConverter(&arg))
				}
			}
			g.Return(Op("&").Qual(t.info.OutputPackageImport+"/transport", requestStructName(fn)).Values(DictFunc(func(d Dict) {
				for _, arg := range arguments {
					typename := types.TypeName(arg.Type)
					if typename == nil {
						panic("need to check and update validation rules: (1)")
					}
					d[structFieldName(&arg)] = Line().Id(*typename).Call(Id(arg.Name))
				}
			})), Nil())
		} else {
			g.Var().Id("req").Qual(t.info.OutputPackageImport+"/transport", requestStructName(fn))
			g.Err().Op(":=").Qual(PackagePathJson, "NewDecoder").Call(Id("r").Dot("Body")).Dot("Decode").Call(Op("&").Id("req"))
			g.Return(Op("&").Id("req"), Err())
		}
	})
}

func stringToTypeConverter(arg *types.Variable) *Statement {
	typename := types.TypeName(arg.Type)
	if typename == nil {
		panic("need to check and update validation rules (2)")
	}
	switch *typename {
	case "string":
		return Id(arg.Name).Op(":=").Id("_param")
	case "int", "int64":
		return List(Id(arg.Name), Err()).Op(":=").Qual(PackagePathStrconv, "ParseInt").Call(Id("_param"), Lit(10), Lit(64)).
			Line().If(Err().Op("!=").Nil()).Block(
			Return(Nil(), Err()),
		)
	case "int32":
		return List(Id(arg.Name), Err()).Op(":=").Qual(PackagePathStrconv, "ParseInt").Call(Id("_param"), Lit(10), Lit(32)).
			Line().If(Err().Op("!=").Nil()).Block(
			Return(Nil(), Err()),
		)
	case "uint", "uint64":
		return List(Id(arg.Name), Err()).Op(":=").Qual(PackagePathStrconv, "ParseUint").Call(Id("_param"), Lit(10), Lit(64)).
			Line().If(Err().Op("!=").Nil()).Block(
			Return(Nil(), Err()),
		)
	case "uint32":
		return List(Id(arg.Name), Err()).Op(":=").Qual(PackagePathStrconv, "ParseUint").Call(Id("_param"), Lit(10), Lit(32)).
			Line().If(Err().Op("!=").Nil()).Block(
			Return(Nil(), Err()),
		)
	}
	return Line().Lit(arg.Name)
}

//		func DecodeHTTPCountResponse(_ context.Context, r *http.Response) (interface{}, error) {
//			var resp svc.CountResponse
//			err := json.NewDecoder(r.Body).Decode(&resp)
//			return resp, err
//		}
func (t *httpConverterTemplate) decodeHTTPResponse(fn *types.Function) *Statement {
	return Func().Id(decodeResponseName(fn)).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("r").Op("*").Qual(PackagePathHttp, "Response"),
		).Params(
		Interface(),
		Error(),
	).
		BlockFunc(func(g *Group) {
			g.Var().Id("resp").Qual(t.info.OutputPackageImport+"/transport", responseStructName(fn))
			g.Err().Op(":=").Qual(PackagePathJson, "NewDecoder").Call(Id("r").Dot("Body")).Dot("Decode").Call(Op("&").Id("resp"))
			g.Return(Op("&").Id("resp"), Err())
		})
}

// Render response encoder.
//		func EncodeHTTPCountResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
//			return DefaultResponseEncoder(ctx, w, response)
//		}
//
func encodeHTTPResponse(fn *types.Function) *Statement {
	return Func().Id(encodeResponseName(fn)).Params(
		Id("ctx").Qual(PackagePathContext, "Context"),
		Id("w").Qual(PackagePathHttp, "ResponseWriter"),
		Id("response").Interface(),
	).Params(
		Error(),
	).Block(
		Return().Id(commonHTTPResponseEncoderName).Call(Id("ctx"), Id("w"), Id("response")),
	)
}

// Render request encoder.
//		func EncodeHTTPCountRequest(ctx context.Context, r *http.Request, request interface{}) error {
//			return DefaultRequestEncoder(ctx, r, request)
//		}
//
func (t *httpConverterTemplate) encodeHTTPRequest(fn *types.Function) *Statement {
	return Func().Id(encodeRequestName(fn)).Params(
		Id("ctx").Qual(PackagePathContext, "Context"),
		Id("r").Op("*").Qual(PackagePathHttp, "Request"),
		Id("request").Interface(),
	).Params(
		Error(),
	).Block(
		Add(t.encodeHTTPRequestBody(fn)),
	)
}

func (t *httpConverterTemplate) encodeHTTPRequestBody(fn *types.Function) *Statement {
	s := &Statement{}
	pathVars := Lit(mstrings.ToURLSnakeCase(fn.Name))
	if FetchHttpMethodTag(fn.Docs) == "GET" {
		s.Id("req").Op(":=").Id("request").Assert(Op("*").Qual(t.info.OutputPackageImport+"/transport", requestStructName(fn))).Line()
		pathVars.Add(t.pathConverters(fn))
	}
	s.Id("r").Dot("URL").Dot("Path").Op("=").
		Qual(PackagePathPath, "Join").Call(Id("r").Dot("URL").Dot("Path"), pathVars)
	if FetchHttpMethodTag(fn.Docs) == "GET" {
		s.Line().Return(Nil())
	} else {
		s.Line().Return(Id(commonHTTPRequestEncoderName).Call(Id("ctx"), Id("r"), Id("request")))
	}
	return s
}

func (t *httpConverterTemplate) pathConverters(fn *types.Function) *Statement {
	converters := &Statement{}
	for _, arg := range RemoveContextIfFirst(fn.Args) {
		typename := types.TypeName(arg.Type)
		if typename == nil {
			panic("need to check and update validation rules (3)")
		}
		converters.Op(",").Add(typeToStringConverters(&arg))
	}
	return converters.Op(",").Line()
}

func typeToStringConverters(arg *types.Variable) *Statement {
	typename := types.TypeName(arg.Type)
	if typename == nil {
		panic("need to check and update validation rules")
	}
	switch *typename {
	case "string":
		return Line().Id("req").Op(".").Add(structFieldName(arg))
	case "int", "int32", "int64":
		return Line().Qual(PackagePathStrconv, "FormatInt").Call(Int64().Call(Id("req").Op(".").Add(structFieldName(arg))), Lit(10))
	case "uint", "uint32", "uint64":
		return Line().Qual(PackagePathStrconv, "FormatUint").Call(Int64().Call(Id("req").Op(".").Add(structFieldName(arg))), Lit(10))
	}
	return Line().Lit(arg.Name)
}
