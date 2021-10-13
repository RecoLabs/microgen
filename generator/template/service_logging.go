package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/recolabs/microgen/generator/strings"
	"github.com/recolabs/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

const (
	_next_                   = "next"
	serviceLoggingStructName = "loggingMiddleware"

	logIgnoreTag = "logs-ignore"
	lenTag       = "logs-len"
)

var ServiceLoggingMiddlewareName = mstrings.ToUpperFirst(serviceLoggingStructName)

type loggingTemplate struct {
	info         *GenerationInfo
	ignoreParams map[string][]string
	lenParams    map[string][]string
}

func NewLoggingTemplate(info *GenerationInfo) Template {
	return &loggingTemplate{
		info: info,
	}
}

// Render all logging.go file.
//
//		// This file was automatically generated by "microgen" utility.
//		// DO NOT EDIT.
//		package middleware
//
//		import (
//			context "context"
//			svc "github.com/recolabs/microgen/examples/svc"
//			log "github.com/go-kit/kit/log"
//			time "time"
//		)
//
//		func ServiceLogging(logger log.Logger) Middleware {
//			return func(next svc.StringService) svc.StringService {
//				return &serviceLogging{
//					logger: logger,
//					next:   next,
//				}
//			}
//		}
//
//		type serviceLogging struct {
//			logger log.Logger
//			next   svc.StringService
//		}
//
//		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text,
// 					"symbol", symbol,
//					"count", count,
// 					"positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func (t *loggingTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("service")
	f.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	f.HeaderComment(t.info.FileHeader)

	f.Comment(ServiceLoggingMiddlewareName + " writes params, results and working time of method call to provided logger after its execution.").
		Line().Func().Id(ServiceLoggingMiddlewareName).Params(Id(_logger_).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newLoggingBody(t.info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceLoggingStructName).Struct(
		Id(_logger_).Qual(PackagePathGoKitLog, "Logger"),
		Id(_next_).Qual(t.info.SourcePackageImport, t.info.Iface.Name),
	)

	// Render functions
	for _, signature := range t.info.Iface.Methods {
		f.Line()
		f.Add(t.loggingFunc(ctx, signature)).Line()
	}
	if len(t.info.Iface.Methods) > 0 {
		f.Type().Op("(")
	}
	for _, signature := range t.info.Iface.Methods {
		if params := RemoveContextIfFirst(signature.Args); t.calcParamAmount(signature.Name, params) > 0 {
			f.Add(t.loggingEntity(ctx, "log"+requestStructName(signature), signature, params))
		}
		if params := removeErrorIfLast(signature.Results); t.calcParamAmount(signature.Name, params) > 0 {
			f.Add(t.loggingEntity(ctx, "log"+responseStructName(signature), signature, params))
		}
	}
	if len(t.info.Iface.Methods) > 0 {
		f.Op(")")
	}

	return f
}

func (loggingTemplate) DefaultPath() string {
	return filenameBuilder(PathService, "logging")
}

func (t *loggingTemplate) Prepare(ctx context.Context) error {
	t.ignoreParams = make(map[string][]string)
	t.lenParams = make(map[string][]string)
	for _, fn := range t.info.Iface.Methods {
		t.ignoreParams[fn.Name] = mstrings.FetchTags(fn.Docs, TagMark+logIgnoreTag)
		t.lenParams[fn.Name] = mstrings.FetchTags(fn.Docs, TagMark+lenTag)
	}
	return nil
}

func (t *loggingTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

// Render body for new logging middleware.
//
//		return func(next svc.StringService) svc.StringService {
//			return &serviceLogging{
//				logger: logger,
//				next:   next,
//			}
//		}
//
func (t *loggingTemplate) newLoggingBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(_next_).Qual(t.info.SourcePackageImport, i.Name),
	).Params(
		Qual(t.info.SourcePackageImport, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceLoggingStructName).Values(
			Dict{
				Id(_logger_): Id(_logger_),
				Id(_next_):   Id(_next_),
			},
		))
	}))
}

func (t *loggingTemplate) loggingEntity(ctx context.Context, name string, fn *types.Function, params []types.Variable) Code {
	if len(params) == 0 {
		return Empty()
	}
	if !t.info.AllowedMethods[fn.Name] {
		return Empty()
	}
	return Id(name).StructFunc(func(g *Group) {
		ignore := t.ignoreParams[fn.Name]
		lenParams := t.lenParams[fn.Name]
		for _, field := range params {
			if !mstrings.IsInStringSlice(field.Name, ignore) {
				g.Id(mstrings.ToUpperFirst(field.Name)).Add(fieldType(ctx, field.Type, false))
			}
			if mstrings.IsInStringSlice(field.Name, lenParams) {
				g.Id("Len" + mstrings.ToUpperFirst(field.Name)).Int().Tag(map[string]string{"json": "len(" + mstrings.ToUpperFirst(field.Name) + ")"})
			}
		}
	})
}

// Render logging middleware for interface method.
//
//		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text, "symbol", symbol,
//					"count", count, "positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func (t *loggingTemplate) loggingFunc(ctx context.Context, signature *types.Function) *Statement {
	normal := normalizeFunction(signature)
	return methodDefinition(ctx, serviceLoggingStructName, &normal.Function).
		BlockFunc(t.loggingFuncBody(signature))
}

// Render logging function body with request/response and time tracking.
//
//		defer func(begin time.Time) {
//			s.logger.Log(
//				"method", "Count",
//				"text", text, "symbol", symbol,
//				"count", count, "positions", positions,
//				"took", time.Since(begin))
//		}(time.Now())
//		return s.next.Count(ctx, text, symbol)
//
func (t *loggingTemplate) loggingFuncBody(signature *types.Function) func(g *Group) {
	normal := normalizeFunction(signature)
	return func(g *Group) {
		if !t.info.AllowedMethods[signature.Name] {
			s := &Statement{}
			if len(normal.Results) > 0 {
				s.Return()
			}
			s.Id(rec(serviceLoggingStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(normal.Args))
			g.Add(s)
			return
		}
		g.Defer().Func().Params(Id("begin").Qual(PackagePathTime, "Time")).Block(
			Id(rec(serviceLoggingStructName)).Dot(_logger_).Dot("Log").CallFunc(func(g *Group) {
				g.Line().Lit("method")
				g.Lit(signature.Name)
				g.Line().Lit("message")
				g.Lit(signature.Name + " called")

				if t.calcParamAmount(signature.Name, RemoveContextIfFirst(signature.Args)) > 0 {
					g.Line().List(Lit("request"), t.logRequest(normal))
				}
				if t.calcParamAmount(signature.Name, removeErrorIfLast(signature.Results)) > 0 {
					g.Line().List(Lit("response"), t.logResponse(normal))
				}
				if !mstrings.IsInStringSlice(nameOfLastResultError(signature), t.ignoreParams[signature.Name]) {
					g.Line().List(Lit(nameOfLastResultError(signature)), Id(nameOfLastResultError(&normal.Function)))
				}

				g.Line().Lit("took")
				g.Qual(PackagePathTime, "Since").Call(Id("begin"))
			}),
		).Call(Qual(PackagePathTime, "Now").Call())
		g.Return().Id(rec(serviceLoggingStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(normal.Args))
	}
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		"err", err,
// 		"result", result,
//		"count", count,
//
func (t *loggingTemplate) paramsNameAndValue(fields []types.Variable, functionName string) *Statement {
	return ListFunc(func(g *Group) {
		ignore := t.ignoreParams[functionName]
		lenParams := t.lenParams[functionName]
		for _, field := range fields {
			if !mstrings.IsInStringSlice(field.Name, ignore) {
				g.Line().List(Lit(field.Name), Id(field.Name))
			}
			if mstrings.IsInStringSlice(field.Name, lenParams) {
				g.Line().List(Lit("len("+field.Name+")"), Len(Id(field.Name)))
			}
		}
	})
}

func (t *loggingTemplate) fillMap(fn *types.Function, params, normal []types.Variable) *Statement {
	return Values(DictFunc(func(d Dict) {
		ignore := t.ignoreParams[fn.Name]
		lenParams := t.lenParams[fn.Name]
		for i, field := range params {
			if !mstrings.IsInStringSlice(field.Name, ignore) {
				d[Id(mstrings.ToUpperFirst(field.Name))] = Id(normal[i].Name)
			}
			if mstrings.IsInStringSlice(field.Name, lenParams) {
				d[Id("Len"+mstrings.ToUpperFirst(field.Name))] = Len(Id(normal[i].Name))
			}
		}
	}))
}

func (t *loggingTemplate) logRequest(fn *normalizedFunction) *Statement {
	paramAmount := t.calcParamAmount(fn.parent.Name, RemoveContextIfFirst(fn.parent.Args))
	if paramAmount <= 0 {
		return Lit("")
	}
	return Id("log" + requestStructName(fn.parent)).Add(t.fillMap(fn.parent, RemoveContextIfFirst(fn.parent.Args), RemoveContextIfFirst(fn.Args)))
}

func (t *loggingTemplate) logResponse(fn *normalizedFunction) *Statement {
	paramAmount := t.calcParamAmount(fn.parent.Name, removeErrorIfLast(fn.parent.Results))
	if paramAmount <= 0 {
		return Lit("")
	}
	return Id("log" + responseStructName(fn.parent)).Add(t.fillMap(fn.parent, removeErrorIfLast(fn.parent.Results), RemoveContextIfFirst(fn.Results)))
}

func (t *loggingTemplate) calcParamAmount(name string, params []types.Variable) int {
	ignore := t.ignoreParams[name]
	lenParams := t.lenParams[name]
	paramAmount := len(params)
	for _, field := range params {
		if mstrings.IsInStringSlice(field.Name, ignore) {
			paramAmount -= 1
		}
		if mstrings.IsInStringSlice(field.Name, lenParams) {
			paramAmount += 1
		}
	}
	return paramAmount
}
