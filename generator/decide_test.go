package generator

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/recolabs/microgen/generator/template"
	"github.com/vetcher/go-astra/types"
)

func Test_getGenerationInfo(t *testing.T) {
	iface := &types.Interface{
		Methods: []*types.Function{
			{
				Base: types.Base{
					Name: "Ignore",
					Docs: []string{
						TagMark + MicrogenMainTag + "-",
					},
				},
			},
			{
				Base: types.Base{
					Name: "OTMStream",
					Docs: []string{
						TagMark + MicrogenMainTag + "one-to-many",
					},
				},
			},
			{
				Base: types.Base{
					Name: "Regular",
				},
			},
		},
	}

	source, err := filepath.Abs("source")
	if err != nil {
		t.Fatalf("failed to get absolute path for source: %v", err)
		return
	}

	type args struct {
		iface       *types.Interface
		absOutPath  string
		sourcePath  string
		packageName string
	}
	tests := []struct {
		name    string
		args    args
		want    *template.GenerationInfo
		wantErr bool
	}{
		{
			name: "interface is nil",
			args: args{
				iface: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "interface is nil",
			args: args{
				absOutPath:  "out",
				sourcePath:  source,
				packageName: "package",
				iface:       iface,
			},
			want: &template.GenerationInfo{
				Iface:                 iface,
				SourcePackageImport:   "package",
				SourceFilePath:        source,
				OutputPackageImport:   "package",
				OutputFilePath:        "out",
				FileHeader:            defaultFileHeader,
				ProtobufPackageImport: "",
				ProtobufClientAddr:    "",
				AllowedMethods: map[string]bool{
					"Ignore":    false,
					"OTMStream": true,
					"Regular":   true,
				},
				OneToManyStreamMethods: map[string]bool{
					"Ignore":    false,
					"OTMStream": true,
					"Regular":   false,
				},
				ManyToManyStreamMethods: map[string]bool{
					"Ignore":    false,
					"OTMStream": false,
					"Regular":   false,
				},
				ManyToOneStreamMethods: map[string]bool{
					"Ignore":    false,
					"OTMStream": false,
					"Regular":   false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGenerationInfo(
				tt.args.iface,
				tt.args.packageName,
				tt.args.sourcePath,
				tt.args.absOutPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGenerationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getGenerationInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagToTemplate(t *testing.T) {
	info := &template.GenerationInfo{}
	type args struct {
		tag  string
		info *template.GenerationInfo
	}
	tests := []struct {
		name      string
		args      args
		wantTmpls []template.Template
	}{
		{
			name: "empty tag",
			args: args{
				tag:  "",
				info: &template.GenerationInfo{},
			},
			wantTmpls: nil,
		},
		{
			name: "nil info",
			args: args{
				tag:  "tag",
				info: nil,
			},
			wantTmpls: nil,
		},
		{
			name: "middleware",
			args: args{
				tag:  MiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewMiddlewareTemplate(info),
			},
		},
		{
			name: "logging",
			args: args{
				tag:  LoggingMiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewMiddlewareTemplate(info),
				template.NewLoggingTemplate(info),
			},
		},
		{
			name: "grpc",
			args: args{
				tag:  GrpcTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsClientTemplate(info),
				template.NewEndpointsServerTemplate(info),
				template.NewGRPCClientTemplate(info),
				template.NewGRPCServerTemplate(info),
				template.NewGRPCEndpointConverterTemplate(info),
				template.NewStubGRPCTypeConverterTemplate(info),
			},
		},
		{
			name: "GrpcClientTag",
			args: args{
				tag:  GrpcClientTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsClientTemplate(info),
				template.NewGRPCClientTemplate(info),
				template.NewGRPCEndpointConverterTemplate(info),
				template.NewStubGRPCTypeConverterTemplate(info),
			},
		},
		{
			name: "GrpcServerTag",
			args: args{
				tag:  GrpcServerTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsServerTemplate(info),
				template.NewGRPCServerTemplate(info),
				template.NewGRPCEndpointConverterTemplate(info),
				template.NewStubGRPCTypeConverterTemplate(info),
			},
		},
		{
			name: "HttpTag",
			args: args{
				tag:  HttpTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsClientTemplate(info),
				template.NewEndpointsServerTemplate(info),
				template.NewHttpServerTemplate(info),
				template.NewHttpClientTemplate(info),
				template.NewHttpConverterTemplate(info),
			},
		},
		{
			name: "HttpServerTag",
			args: args{
				tag:  HttpServerTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsServerTemplate(info),
				template.NewHttpServerTemplate(info),
				template.NewHttpConverterTemplate(info),
			},
		},
		{
			name: "HttpClientTag",
			args: args{
				tag:  HttpClientTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsClientTemplate(info),
				template.NewHttpClientTemplate(info),
				template.NewHttpConverterTemplate(info),
			},
		},
		{
			name: "RecoveringMiddlewareTag",
			args: args{
				tag:  RecoveringMiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewMiddlewareTemplate(info),
				template.NewRecoverTemplate(info),
			},
		},
		{
			name: "MainTag",
			args: args{
				tag:  MainTag,
				info: info,
			},
			wantTmpls: nil,
		},
		{
			name: "ErrorLoggingMiddlewareTag",
			args: args{
				tag:  ErrorLoggingMiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewMiddlewareTemplate(info),
				template.NewErrorLoggingTemplate(info),
			},
		},
		{
			name: "CachingMiddlewareTag",
			args: args{
				tag:  CachingMiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewMiddlewareTemplate(info),
				template.NewCacheMiddlewareTemplate(info),
			},
		},
		{
			name: "TracingMiddlewareTag",
			args: args{
				tag:  TracingMiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.EmptyTemplate{},
			},
		},
		{
			name: "MetricsMiddlewareTag",
			args: args{
				tag:  MetricsMiddlewareTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.EmptyTemplate{},
			},
		},
		{
			name: "ServiceDiscoveryTag",
			args: args{
				tag:  ServiceDiscoveryTag,
				info: info,
			},
			wantTmpls: []template.Template{
				template.EmptyTemplate{},
			},
		},
		{
			name: "Transport",
			args: args{
				tag:  Transport,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsClientTemplate(info),
				template.NewEndpointsServerTemplate(info),
			},
		},
		{
			name: "TransportClient",
			args: args{
				tag:  TransportClient,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsClientTemplate(info),
			},
		},
		{
			name: "TransportServer",
			args: args{
				tag:  TransportServer,
				info: info,
			},
			wantTmpls: []template.Template{
				template.NewExchangeTemplate(info),
				template.NewEndpointsTemplate(info),
				template.NewEndpointsServerTemplate(info),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTmpls := tagToTemplate(tt.args.tag, tt.args.info); !reflect.DeepEqual(gotTmpls, tt.wantTmpls) {
				t.Errorf("tagToTemplate() = %v, want %v", gotTmpls, tt.wantTmpls)
			}
		})
	}
}
