package generator

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/recolabs/microgen/generator/template"
	"github.com/recolabs/microgen/generator/write_strategy"
	"github.com/recolabs/microgen/mocks"
	"github.com/stretchr/testify/mock"
)

func TestNewGenUnit(t *testing.T) {
	failedToPrepareTemplate := new(mocks.Template)
	failedToPrepareTemplate.
		On("DefaultPath").
		Return("")
	failedToPrepareTemplate.
		On("Prepare", mock.Anything).
		Return(fmt.Errorf("failed to prepare template"))

	failedToChooseStrategyTemplate := new(mocks.Template)
	failedToChooseStrategyTemplate.
		On("DefaultPath").
		Return("")
	failedToChooseStrategyTemplate.
		On("Prepare", mock.Anything).
		Return(nil)
	failedToChooseStrategyTemplate.
		On("ChooseStrategy", mock.Anything).
		Return(nil, fmt.Errorf("failed to choose strategy"))

	validStrategy := new(mocks.Strategy)
	validTemplate := new(mocks.Template)
	validTemplate.
		On("Prepare", mock.Anything).
		Return(nil)
	validTemplate.
		On("ChooseStrategy", mock.Anything).
		Return(validStrategy, nil)

	type args struct {
		ctx     context.Context
		tmpl    template.Template
		outPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *GenerationUnit
		wantErr bool
	}{
		{
			name: "failed to prepare template",
			args: args{
				ctx:     context.Background(),
				tmpl:    failedToPrepareTemplate,
				outPath: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "failed to choose strategy",
			args: args{
				ctx:     context.Background(),
				tmpl:    failedToChooseStrategyTemplate,
				outPath: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success case",
			args: args{
				ctx:     context.Background(),
				tmpl:    validTemplate,
				outPath: "path",
			},
			want: &GenerationUnit{
				template:      validTemplate,
				writeStrategy: validStrategy,
				absOutPath:    "path",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGenUnit(tt.args.ctx, tt.args.tmpl, tt.args.outPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenUnit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGenUnit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerationUnit_Generate(t *testing.T) {
	renderer := new(mocks.Renderer)
	successTemplate := new(mocks.Template)
	successTemplate.
		On("Render", mock.Anything).
		Return(renderer)

	failedWriteStrategy := new(mocks.Strategy)
	failedWriteStrategy.
		On("Write", mock.Anything).
		Return(fmt.Errorf("failed to write"))

	successWriteStrategy := new(mocks.Strategy)
	successWriteStrategy.
		On("Write", mock.Anything).
		Return(nil)

	type fields struct {
		template      template.Template
		writeStrategy write_strategy.Strategy
		absOutPath    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no template",
			fields: fields{
				template:      nil,
				writeStrategy: successWriteStrategy,
				absOutPath:    "",
			},
			wantErr: true,
		},
		{
			name: "no strategy",
			fields: fields{
				template:      successTemplate,
				writeStrategy: nil,
				absOutPath:    "",
			},
			wantErr: true,
		},
		{
			name: "write fail",
			fields: fields{
				template:      successTemplate,
				writeStrategy: failedWriteStrategy,
				absOutPath:    "",
			},
			wantErr: true,
		},
		{
			name: "success case",
			fields: fields{
				template:      successTemplate,
				writeStrategy: successWriteStrategy,
				absOutPath:    "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GenerationUnit{
				template:      tt.fields.template,
				writeStrategy: tt.fields.writeStrategy,
				absOutPath:    tt.fields.absOutPath,
			}
			if err := g.Generate(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("GenerationUnit.Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerationUnit_Path(t *testing.T) {
	type fields struct {
		template      template.Template
		writeStrategy write_strategy.Strategy
		absOutPath    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty string",
			fields: fields{
				template:      nil,
				writeStrategy: nil,
				absOutPath:    "",
			},
			want: "",
		},
		{
			name: "success case",
			fields: fields{
				template:      nil,
				writeStrategy: nil,
				absOutPath:    "hello",
			},
			want: "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GenerationUnit{
				template:      tt.fields.template,
				writeStrategy: tt.fields.writeStrategy,
				absOutPath:    tt.fields.absOutPath,
			}
			if got := g.Path(); got != tt.want {
				t.Errorf("GenerationUnit.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}
