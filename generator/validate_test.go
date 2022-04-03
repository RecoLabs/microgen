package generator

import (
	"errors"
	"reflect"
	"testing"

	"github.com/vetcher/go-astra/types"
)

var (
	contextVar = types.Variable{
		Base: types.Base{
			Name: "ctx",
		},
		Type: types.TImport{
			Import: &types.Import{
				Package: "context",
			},
			Next: types.TName{
				TypeName: "Context",
			},
		},
	}

	complexPBFile = &types.File{
		Structures: []types.Struct{
			{
				Base: types.Base{
					Name: "FooRequest",
				},
				Fields: []types.StructField{
					{
						Variable: types.Variable{
							Base: types.Base{
								Name: "Bar",
							},
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
				},
			},
			{
				Base: types.Base{
					Name: "FooResponse",
				},
			},
		},
	}
)

func TestValidateInterface(t *testing.T) {
	type args struct {
		iface    *types.Interface
		pbGoFile *types.File
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil interface",
			args: args{
				iface:    nil,
				pbGoFile: &types.File{},
			},
			wantErr: true,
		},
		{
			name: "no methods",
			args: args{
				iface: &types.Interface{
					Methods: []*types.Function{},
				},
				pbGoFile: &types.File{},
			},
			wantErr: true,
		},
		{
			name: "success context",
			args: args{
				iface: &types.Interface{
					Methods: []*types.Function{
						{
							Base: types.Base{
								Name: "Foo",
							},
							Args: []types.Variable{
								contextVar,
								{
									Base: types.Base{
										Name: "bar",
									},
									Type: types.TName{
										TypeName: "int64",
									},
								},
							},
							Results: []types.Variable{
								{
									Base: types.Base{
										Name: "err",
									},
									Type: types.TName{
										TypeName: "error",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateInterface(tt.args.iface, tt.args.pbGoFile); (err != nil) != tt.wantErr {
				t.Errorf("ValidateInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateFunction(t *testing.T) {
	type args struct {
		fn       *types.Function
		pbGoFile *types.File
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []error
	}{
		{
			name: "nil function",
			args: args{
				fn:       nil,
				pbGoFile: &types.File{},
			},
			wantErrs: []error{
				errors.New("nil function"),
			},
		},
		{
			name: "ignore -",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Docs: []string{
							TagMark + MicrogenMainTag + "-",
						},
					},
				},
				pbGoFile: &types.File{},
			},
			wantErrs: nil,
		},
		{
			name: "unnamed var in stream function",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
						Docs: []string{
							TagMark + MicrogenMainTag + "one-to-many",
						},
					},
					Args: []types.Variable{
						{
							Type: types.TName{
								TypeName: "Bar",
							},
						},
					},
				},
				pbGoFile: &types.File{},
			},
			wantErrs: []error{
				errors.New("Foo: unnamed parameter of type Bar"),
			},
		},
		{
			name: "context not first argument",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						{
							Base: types.Base{
								Name: "bar",
							},
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
					Results: []types.Variable{
						{
							Base: types.Base{
								Name: "err",
							},
							Type: types.TName{
								TypeName: "error",
							},
						},
					},
				},
				pbGoFile: complexPBFile,
			},
			wantErrs: []error{
				errors.New("Foo: first argument should be of type context.Context"),
			},
		},
		{
			name: "error not last result",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						contextVar,
						{
							Base: types.Base{
								Name: "bar",
							},
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
					Results: []types.Variable{},
				},
				pbGoFile: complexPBFile,
			},
			wantErrs: []error{
				errors.New("Foo: last result should be of type error"),
			},
		},
		{
			name: "unnamed var",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						contextVar,
						{
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
					Results: []types.Variable{
						{
							Base: types.Base{
								Name: "err",
							},
							Type: types.TName{
								TypeName: "error",
							},
						},
					},
				},
			},
			wantErrs: []error{
				errors.New("Foo: unnamed parameter of type int64"),
			},
		},
		{
			name: "success",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						contextVar,
						{
							Base: types.Base{
								Name: "bar",
							},
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
					Results: []types.Variable{
						{
							Base: types.Base{
								Name: "err",
							},
							Type: types.TName{
								TypeName: "error",
							},
						},
					},
				},
			},
			wantErrs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotErrs := validateFunction(tt.args.fn, tt.args.pbGoFile); !reflect.DeepEqual(gotErrs, tt.wantErrs) {
				t.Errorf("validateFunction() = %v, want %v", gotErrs, tt.wantErrs)
			}
		})
	}
}

func Test_requestStructName(t *testing.T) {
	type args struct {
		signature *types.Function
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success case",
			args: args{
				signature: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
				},
			},
			want: "FooRequest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestStructName(tt.args.signature); got != tt.want {
				t.Errorf("requestStructName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_responseStructName(t *testing.T) {
	type args struct {
		signature *types.Function
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success case",
			args: args{
				signature: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
				},
			},
			want: "FooResponse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := responseStructName(tt.args.signature); got != tt.want {
				t.Errorf("responseStructName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findStruct(t *testing.T) {
	type args struct {
		name   string
		grpcPb *types.File
	}
	tests := []struct {
		name string
		args args
		want *types.Struct
	}{
		{
			name: "nil pb file",
			args: args{
				name:   "Foo",
				grpcPb: nil,
			},
			want: nil,
		},
		{
			name: "no structures in pb file",
			args: args{
				name:   "Foo",
				grpcPb: &types.File{},
			},
			want: nil,
		},
		{
			name: "no matching structures in pb file",
			args: args{
				name: "Foo",
				grpcPb: &types.File{
					Structures: []types.Struct{
						{
							Base: types.Base{
								Name: "Bar",
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "matching structures in pb file",
			args: args{
				name: "Foo",
				grpcPb: &types.File{
					Structures: []types.Struct{
						{
							Base: types.Base{
								Name: "Foo",
							},
						},
					},
				},
			},
			want: &types.Struct{
				Base: types.Base{
					Name: "Foo",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findStruct(tt.args.name, tt.args.grpcPb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findField(t *testing.T) {
	type args struct {
		name string
		s    *types.Struct
	}
	tests := []struct {
		name string
		args args
		want *types.StructField
	}{
		{
			name: "nil struct",
			args: args{
				name: "Foo",
				s:    nil,
			},
			want: nil,
		},
		{
			name: "no matching fields in struct",
			args: args{
				name: "Foo",
				s: &types.Struct{
					Fields: []types.StructField{
						{
							Variable: types.Variable{
								Base: types.Base{
									Name: "Bar",
								},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "matching fields in struct",
			args: args{
				name: "Foo",
				s: &types.Struct{
					Fields: []types.StructField{
						{
							Variable: types.Variable{
								Base: types.Base{
									Name: "Foo",
								},
							},
						},
					},
				},
			},
			want: &types.StructField{
				Variable: types.Variable{
					Base: types.Base{
						Name: "Foo",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findField(tt.args.name, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_typeWithNoImport(t *testing.T) {
	type args struct {
		field types.Type
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil type",
			args: args{
				field: nil,
			},
			want: "",
		},
		{
			name: "basic type",
			args: args{
				field: types.TName{
					TypeName: "Foo",
				},
			},
			want: "Foo",
		},
		{
			name: "import type",
			args: args{
				field: types.TImport{
					Next: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: "Foo",
		},
		{
			name: "array type - IsEllipsis",
			args: args{
				field: types.TArray{
					IsEllipsis: true,
					Next: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: "...Foo",
		},
		{
			name: "array type - IsSlice",
			args: args{
				field: types.TArray{
					IsSlice: true,
					Next: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: "[]Foo",
		},
		{
			name: "array type - not ellipsis and slice",
			args: args{
				field: types.TArray{
					ArrayLen: 5,
					Next: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: "[5]Foo",
		},
		{
			name: "map type",
			args: args{
				field: types.TMap{
					Key: types.TName{
						TypeName: "Foo",
					},
					Value: types.TName{
						TypeName: "Bar",
					},
				},
			},
			want: "map[Foo]Bar",
		},
		{
			name: "pointer type",
			args: args{
				field: types.TPointer{
					NumberOfPointers: 2,
					Next: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: "**Foo",
		},
		{
			name: "interface type",
			args: args{
				field: types.TInterface{
					Interface: &types.Interface{
						Base: types.Base{
							Name: "Foo",
						},
					},
				},
			},
			want: "type Foo interface {\n\t\n}",
		},
		{
			name: "ellipsis type",
			args: args{
				field: types.TEllipsis{
					Next: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: "...Foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeWithNoImport(tt.args.field); got != tt.want {
				t.Errorf("typeWithNoImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateFuncionInPbGoFile(t *testing.T) {
	type args struct {
		fn       *types.Function
		pbGoFile *types.File
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []error
	}{
		{
			name: "function is nil",
			args: args{
				fn:       nil,
				pbGoFile: &types.File{},
			},
			wantErrs: []error{
				errors.New("function is nil"),
			},
		},
		{
			name: "pbGoFile is nil",
			args: args{
				fn:       &types.Function{},
				pbGoFile: nil,
			},
			wantErrs: []error{
				errors.New("pbGoFile is nil"),
			},
		},
		{
			name: "request struct is not in pbGoFile",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
				},
				pbGoFile: &types.File{
					Structures: []types.Struct{},
				},
			},
			wantErrs: []error{
				errors.New("did not find struct FooRequest in grpc pb file"),
			},
		},
		{
			name: "field name mismatch",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						contextVar,
						{
							Base: types.Base{
								Name: "bar",
							},
						},
					},
				},
				pbGoFile: &types.File{
					Structures: []types.Struct{
						{
							Base: types.Base{
								Name: "FooRequest",
							},
						},
						{
							Base: types.Base{
								Name: "FooResponse",
							},
						},
					},
				},
			},
			wantErrs: []error{
				errors.New("did not find field Bar in struct FooRequest in grpc pb file"),
			},
		},
		{
			name: "field type mismatch",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						contextVar,
						{
							Base: types.Base{
								Name: "bar",
							},
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
				},
				pbGoFile: &types.File{
					Structures: []types.Struct{
						{
							Base: types.Base{
								Name: "FooRequest",
							},
							Fields: []types.StructField{
								{
									Variable: types.Variable{
										Base: types.Base{
											Name: "Bar",
										},
										Type: types.TName{
											TypeName: "int32",
										},
									},
								},
							},
						},
						{
							Base: types.Base{
								Name: "FooResponse",
							},
						},
					},
				},
			},
			wantErrs: []error{
				errors.New("argument bar in function Foo has different type in pb.go file. expected int64 got int32"),
			},
		},
		{
			name: "response struct is not in pbGoFile",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
				},
				pbGoFile: &types.File{
					Structures: []types.Struct{
						{
							Base: types.Base{
								Name: "FooRequest",
							},
						},
					},
				},
			},
			wantErrs: []error{
				errors.New("did not find struct FooResponse in grpc pb file"),
			},
		},
		{
			name: "success simple",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
				},
				pbGoFile: &types.File{
					Structures: []types.Struct{
						{
							Base: types.Base{
								Name: "FooRequest",
							},
						},
						{
							Base: types.Base{
								Name: "FooResponse",
							},
						},
					},
				},
			},
			wantErrs: nil,
		},
		{
			name: "success",
			args: args{
				fn: &types.Function{
					Base: types.Base{
						Name: "Foo",
					},
					Args: []types.Variable{
						contextVar,
						{
							Base: types.Base{
								Name: "bar",
							},
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
				},
				pbGoFile: complexPBFile,
			},
			wantErrs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotErrs := validateFuncionInPbGoFile(tt.args.fn, tt.args.pbGoFile); !reflect.DeepEqual(gotErrs, tt.wantErrs) {
				t.Errorf("validateFuncionInPbGoFile() = %v, want %v", gotErrs, tt.wantErrs)
			}
		})
	}
}

func Test_isArgumentsAllowSmartPath(t *testing.T) {
	type args struct {
		fn *types.Function
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "func is nil",
			args: args{
				fn: nil,
			},
			want: false,
		},
		{
			name: "func has no arguments",
			args: args{
				fn: &types.Function{
					Args: []types.Variable{},
				},
			},
			want: true,
		},
		{
			name: "func has only ctx argument",
			args: args{
				fn: &types.Function{
					Args: []types.Variable{
						contextVar,
					},
				},
			},
			want: true,
		},
		{
			name: "func has ctx argument - allowed",
			args: args{
				fn: &types.Function{
					Args: []types.Variable{
						contextVar,
						{
							Type: types.TName{
								TypeName: "int64",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "func has ctx argument - not allowed",
			args: args{
				fn: &types.Function{
					Args: []types.Variable{
						contextVar,
						{
							Type: types.TName{
								TypeName: "Foo",
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isArgumentsAllowSmartPath(tt.args.fn); got != tt.want {
				t.Errorf("isArgumentsAllowSmartPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_canInsertToPath(t *testing.T) {
	type args struct {
		p *types.Variable
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "var is nil",
			args: args{
				p: nil,
			},
			want: false,
		},
		{
			name: "can't insert to path",
			args: args{
				p: &types.Variable{
					Type: types.TName{
						TypeName: "Foo",
					},
				},
			},
			want: false,
		},
		{
			name: "can insert to path",
			args: args{
				p: &types.Variable{
					Type: types.TName{
						TypeName: "int64",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canInsertToPath(tt.args.p); got != tt.want {
				t.Errorf("canInsertToPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_composeErrors(t *testing.T) {
	type args struct {
		errs []error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "nil errors",
			args: args{
				errs: nil,
			},
			want: nil,
		},
		{
			name: "no errors",
			args: args{
				errs: []error{},
			},
			want: nil,
		},
		{
			name: "one error",
			args: args{
				errs: []error{
					errors.New("foo"),
				},
			},
			want: errors.New("foo"),
		},
		{
			name: "many errors",
			args: args{
				errs: []error{
					errors.New("foo"),
					errors.New("bar"),
				},
			},
			want: errors.New("many errors:\nfoo\nbar"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := composeErrors(tt.args.errs...); !reflect.DeepEqual(err, tt.want) {
				t.Errorf("composeErrors() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}
