package template

import (
	goparser "go/parser"
	"go/token"
	"testing"

	"bytes"

	"github.com/devimteam/microgen/generator"
	parser "github.com/devimteam/microgen/parser"
)

func TestMiddlewareForCountSvc(t *testing.T) {
	src := `package stringsvc

	import (
		"context"
	)

	type StringService interface {
		Count(ctx context.Context, text string, symbol string) (count int, positions []int)
	}`

	out := `// This file was automatically generated by "microgen" utility.
// Please, do not edit.
package stringsvc

import svc "github.com/devimteam/microgen/test/svc"

type Middleware func(svc.StringService) svc.StringService
` // Blank line!
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Errorf("unable to parse file: %v", err)
	}
	fs, err := parser.ParseInterface(f, "StringService")
	if err != nil {
		t.Errorf("could not get interface func signatures: %v", err)
	}
	buf := bytes.NewBuffer([]byte{})
	gen := generator.NewGenerator([]generator.Template{
		&MiddlewareTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
	}, fs, generator.NewWriterStrategy(buf))
	err = gen.Generate()
	if err != nil {
		t.Errorf("unable to generate: %v", err)
	}
	if buf.String() != out {
		t.Errorf("Got:\n\n%v\n\nShould be:\n\n%v", buf.String(), out)
	}
}
