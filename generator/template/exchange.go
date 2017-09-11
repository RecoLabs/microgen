package template

import (
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
)

type ExchangeTemplate struct {
	Info *GenerationInfo
}

func requestStructName(signature *types.Function) string {
	return signature.Name + "Request"
}

func responseStructName(signature *types.Function) string {
	return signature.Name + "Response"
}

// Renders exchanges file.
//
//  package visitsvc
//
//  import (
//  	"gitlab.devim.team/microservices/visitsvc/entity"
//  )
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
//  type CreateVisitResponse struct {
//  	Res *entity.Visit `json:"res"`
//  	Err error         `json:"err"`
//  }
//
func (t *ExchangeTemplate) Render(i *GenerationInfo) *Statement {
	f := Statement{}

	for _, signature := range i.Iface.Methods {
		f.Add(exchange(requestStructName(signature), signature.Args)).Line()
		f.Add(exchange(responseStructName(signature), signature.Results)).Line()
	}

	return &f
}

func (ExchangeTemplate) DefaultPath() string {
	return "./exchanges.go"
}

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func exchange(name string, params []types.Variable) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, param := range removeContextIfFirst(params) {
			g.Add(structField(&param))
		}
	}).Line()
}
