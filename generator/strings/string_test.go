package strings

import (
	"reflect"
	"testing"
	"unicode"
)

func TestToUpperFirst(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "success case",
			args: args{
				s: "hello",
			},
			want: "Hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUpperFirst(tt.args.s); got != tt.want {
				t.Errorf("ToUpperFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSomeCaseWithSep(t *testing.T) {
	type args struct {
		sep      rune
		runeConv func(rune) rune
		input    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				sep:      '_',
				runeConv: func(r rune) rune { return r },
				input:    "",
			},
			want: "",
		},
		{
			name: "success case with extended space",
			args: args{
				sep:      '_',
				runeConv: unicode.ToLower,
				input:    "Hello_World",
			},
			want: "hello_world",
		},
		{
			name: "success case without extended space",
			args: args{
				sep:      '*',
				runeConv: unicode.ToLower,
				input:    "Hello*World",
			},
			want: "hello*world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFunc := ToSomeCaseWithSep(tt.args.sep, tt.args.runeConv)
			if got := gotFunc(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToSomeCaseWithSep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isExtendedSpace(t *testing.T) {
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "space",
			args: args{
				r: ' ',
			},
			want: true,
		},
		{
			name: "tab",
			args: args{
				r: '\t',
			},
			want: true,
		},
		{
			name: "newline",
			args: args{
				r: '\n',
			},
			want: true,
		},
		{
			name: "carriage return",
			args: args{
				r: '\r',
			},
			want: true,
		},
		{
			name: "form feed",
			args: args{
				r: '\f',
			},
			want: true,
		},
		{
			name: "vertical tab",
			args: args{
				r: '\v',
			},
			want: true,
		},
		{
			name: "non-extended space",
			args: args{
				r: 'a',
			},
			want: false,
		},
		{
			name: "underline",
			args: args{
				r: '_',
			},
			want: true,
		},
		{
			name: "dot",
			args: args{
				r: '.',
			},
			want: true,
		},
		{
			name: "middleline",
			args: args{
				r: '-',
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExtendedSpace(tt.args.r); got != tt.want {
				t.Errorf("isExtendedSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToLowerFirst(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "success case",
			args: args{
				s: "Hello",
			},
			want: "hello",
		},
		{
			name: "success case - no change",
			args: args{
				s: "hello",
			},
			want: "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLowerFirst(tt.args.s); got != tt.want {
				t.Errorf("ToLowerFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInStringSlice(t *testing.T) {
	type args struct {
		what  string
		where []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty string",
			args: args{
				what:  "",
				where: []string{""},
			},
			want: true,
		},
		{
			name: "empty string slice",
			args: args{
				what:  "hello",
				where: []string{},
			},
			want: false,
		},
		{
			name: "success case",
			args: args{
				what:  "hello",
				where: []string{"hello", "world"},
			},
			want: true,
		},
		{
			name: "success case - case insensitive",
			args: args{
				what:  "Hello",
				where: []string{"hello", "world"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInStringSlice(tt.args.what, tt.args.where); got != tt.want {
				t.Errorf("IsInStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchTags(t *testing.T) {
	type args struct {
		strs   []string
		prefix string
	}
	tests := []struct {
		name     string
		args     args
		wantTags []string
	}{
		{
			name: "empty list",
			args: args{
				strs:   []string{},
				prefix: "",
			},
			wantTags: nil,
		},
		{
			name: "empty prefix",
			args: args{
				strs:   []string{"hello", "world"},
				prefix: "",
			},
			wantTags: []string{"hello", "world"},
		},
		{
			name: "single tag",
			args: args{
				strs:   []string{"pre:hello", "world"},
				prefix: "pre:",
			},
			wantTags: []string{"hello"},
		},
		{
			name: "multiple tags",
			args: args{
				strs:   []string{"pre:hello", "pre:world"},
				prefix: "pre:",
			},
			wantTags: []string{"hello", "world"},
		},
		{
			name: "no tags found",
			args: args{
				strs:   []string{"hello", "world"},
				prefix: "pre:",
			},
			wantTags: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTags := FetchTags(tt.args.strs, tt.args.prefix); !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("FetchTags() = %v, want %v", gotTags, tt.wantTags)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				str: "",
			},
			want: "",
		},
		{
			name: "only first letter is upper case",
			args: args{
				str: "Hello",
			},
			want: "hello",
		},
		{
			name: "multiple letters are upper case",
			args: args{
				str: "HelloWorld",
			},
			want: "helloWorld",
		},
		{
			name: "multiple letters are upper case in the start",
			args: args{
				str: "HHHelloWorld",
			},
			want: "hhHelloWorld",
		},
		{
			name: "no change",
			args: args{
				str: "helloWorld",
			},
			want: "helloWorld",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLower(tt.args.str); got != tt.want {
				t.Errorf("ToLower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastUpperOrFirst(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				str: "",
			},
			want: "",
		},
		{
			name: "only first letter is upper case",
			args: args{
				str: "Hello",
			},
			want: "H",
		},
		{
			name: "multiple letters are upper case",
			args: args{
				str: "HelloWorld",
			},
			want: "W",
		},
		{
			name: "no upper case letters",
			args: args{
				str: "helloworld",
			},
			want: "h",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastUpperOrFirst(tt.args.str); got != tt.want {
				t.Errorf("LastUpperOrFirst() = %v, want %v", got, tt.want)
			}
		})
	}
}
