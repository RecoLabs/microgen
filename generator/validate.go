package generator

import (
	"fmt"
	"strconv"
	"strings"

	mstrings "github.com/recolabs/microgen/generator/strings"
	"github.com/recolabs/microgen/generator/template"
	"github.com/vetcher/go-astra/types"
)

func ValidateInterface(iface *types.Interface, pbGoFile *types.File) error {
	if iface == nil {
		return fmt.Errorf("nil interface")
	}

	var errs []error
	if len(iface.Methods) == 0 {
		errs = append(errs, fmt.Errorf("%s does not have any methods", iface.Name))
	}
	for _, m := range iface.Methods {
		errs = append(errs, validateFunction(m, pbGoFile)...)
	}
	return composeErrors(errs...)
}

// Rules:
// * First argument is context.Context.
// * Last result is error.
// * All params have names.
func validateFunction(fn *types.Function, pbGoFile *types.File) (errs []error) {
	if fn == nil {
		errs = append(errs, fmt.Errorf("nil function"))
		return
	}
	microgenTags := mstrings.FetchTags(fn.Docs, TagMark+MicrogenMainTag)

	// don't validate when `@microgen -` provided
	if mstrings.ContainTag(microgenTags, "-") {
		return
	}
	if mstrings.ContainTag(microgenTags, "one-to-many") ||
		mstrings.ContainTag(microgenTags, "many-to-many") ||
		mstrings.ContainTag(microgenTags, "many-to-one") {
		for _, param := range append(fn.Args, fn.Results...) {
			if param.Name == "" {
				errs = append(errs, fmt.Errorf("%s: unnamed parameter of type %s", fn.Name, param.Type.String()))
			}
		}
		return
	}

	if !template.IsContextFirst(fn.Args) {
		errs = append(errs, fmt.Errorf("%s: first argument should be of type context.Context", fn.Name))
	}
	if !template.IsErrorLast(fn.Results) {
		errs = append(errs, fmt.Errorf("%s: last result should be of type error", fn.Name))
	}
	for _, param := range append(fn.Args, fn.Results...) {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed parameter of type %s", fn.Name, param.Type.String()))
		}
		if iface := types.TypeInterface(param.Type); iface != nil && !iface.(types.TInterface).Interface.IsEmpty() {
			errs = append(errs, fmt.Errorf("%s: non empty interface %s is not allowed, delcare it outside", fn.Name, param.String()))
		}
		if strct := types.TypeStruct(param.Type); strct != nil {
			errs = append(errs, fmt.Errorf("%s: raw struct %s is not allowed, declare it outside", fn.Name, param.Name))
		}
		if f := types.TypeFunction(param.Type); f != nil {
			errs = append(errs, fmt.Errorf("%s: raw function %s is not allowed, declare it outside", fn.Name, param.Name))
		}
	}
	if template.FetchHttpMethodTag(fn.Docs) == "GET" && !isArgumentsAllowSmartPath(fn) {
		errs = append(errs, fmt.Errorf("%s: can't use GET method with provided arguments", fn.Name))
	}
	if pbGoFile != nil {
		errs = append(errs, validateFuncionInPbGoFile(fn, pbGoFile)...)
	}
	return
}

func requestStructName(signature *types.Function) string {
	return signature.Name + "Request"
}

func responseStructName(signature *types.Function) string {
	return signature.Name + "Response"
}

func findStruct(name string, grpcPb *types.File) *types.Struct {
	if grpcPb == nil {
		return nil
	}

	for _, s := range grpcPb.Structures {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

func findField(name string, s *types.Struct) *types.StructField {
	if s == nil {
		return nil
	}

	for _, f := range s.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func typeWithNoImport(field types.Type) string {
	typeName := ""
	for field != nil {
		switch i := field.(type) {
		case types.TImport:
			field = i.Next
		case types.TName:
			typeName += i.TypeName
			field = nil
		case types.TArray:
			str := ""
			if i.IsEllipsis {
				str += "..."
			} else if i.IsSlice {
				str += "[]"
			} else {
				str += "[" + strconv.Itoa(i.ArrayLen) + "]"
			}
			typeName += str
			field = i.Next
		case types.TMap:
			typeName += "map[" + i.Key.String() + "]" + i.Value.String()
			field = nil
		case types.TPointer:
			typeName += strings.Repeat("*", i.NumberOfPointers)
			field = i.Next
		case types.TInterface:
			typeName += i.Interface.String()
			field = nil
		case types.TEllipsis:
			typeName += "..."
			field = i.Next
		default:
			break
		}
	}

	return typeName
}

// TODO - split this to two functions
func validateFuncionInPbGoFile(fn *types.Function, pbGoFile *types.File) (errs []error) {
	if fn == nil {
		errs = append(errs, fmt.Errorf("function is nil"))
		return
	}
	if pbGoFile == nil {
		errs = append(errs, fmt.Errorf("pbGoFile is nil"))
		return
	}

	requestStructName := requestStructName(fn)
	s := findStruct(requestStructName, pbGoFile)
	if s == nil {
		errs = append(errs, fmt.Errorf("did not find struct %v in grpc pb file", requestStructName))
		return
	}

	// TODO - we should also validate the struct hasn't any fields that are not in the function signature
	for i, arg := range fn.Args {
		if i == 0 {
			// Note - we already now the first argument is 'ctx context.context'
			// TODO - handle stream input and always perform a check
			continue
		}
		protoFieldName := mstrings.ToUpperFirst(arg.Name)
		foundField := findField(protoFieldName, s)
		if foundField == nil {
			errs = append(errs, fmt.Errorf("did not find field %v in struct %v in grpc pb file", protoFieldName, requestStructName))
		} else {
			argType := typeWithNoImport(arg.Type)
			foundType := typeWithNoImport(foundField.Type)
			if argType != foundType {
				errs = append(errs, fmt.Errorf("argument %v in function %v has different type in pb.go file. expected %v got %v", arg.Name, fn.Name, argType, foundType))
			}
		}

	}

	responseStructName := responseStructName(fn)
	s = findStruct(responseStructName, pbGoFile)
	if s == nil {
		errs = append(errs, fmt.Errorf("did not find struct %v in grpc pb file", responseStructName))
		return
	}
	for i, res := range fn.Results {
		if i == len(fn.Results)-1 {
			// Note - we already now the last return is 'err error'
			continue
		}
		protoFieldName := mstrings.ToUpperFirst(res.Name)
		foundField := findField(protoFieldName, s)
		if foundField == nil {
			errs = append(errs, fmt.Errorf("did not find field %v in struct %v in grpc pb file", protoFieldName, responseStructName))
		} else {
			resType := typeWithNoImport(res.Type)
			foundType := typeWithNoImport(foundField.Type)
			if resType != foundType {
				errs = append(errs, fmt.Errorf("result %v in function %v has different type in pb.go file. expected %v got %v", res.Name, fn.Name, resType, foundType))
			}
		}

	}
	return
}

func isArgumentsAllowSmartPath(fn *types.Function) bool {
	if fn == nil {
		return false
	}

	for _, arg := range template.RemoveContextIfFirst(fn.Args) {
		if !canInsertToPath(&arg) {
			return false
		}
	}
	return true
}

var insertableToUrlTypes = []string{"string", "int", "int32", "int64", "uint", "uint32", "uint64"}

// We can make url variable from string, int, int32, int64, uint, uint32, uint64
func canInsertToPath(p *types.Variable) bool {
	if p == nil {
		return false
	}

	name := types.TypeName(p.Type)
	return name != nil && mstrings.IsInStringSlice(*name, insertableToUrlTypes)
}

func composeErrors(errs ...error) error {
	if len(errs) > 0 {
		var strs []string
		for _, err := range errs {
			if err != nil {
				strs = append(strs, err.Error())
			}
		}
		if len(strs) == 1 {
			return fmt.Errorf(strs[0])
		}
		if len(strs) > 0 {
			return fmt.Errorf("many errors:\n%v", strings.Join(strs, "\n"))
		}
	}
	return nil
}
