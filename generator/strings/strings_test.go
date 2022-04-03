package strings

import "testing"

func TestFetchMetaInfo(t *testing.T) {
	type args struct {
		tag      string
		comments []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "tag found",
			args: args{
				tag:      "// tag",
				comments: []string{"// tag:wow"},
			},
			want: "wow",
		},
		{
			name: "tag not found",
			args: args{
				tag:      "// tag",
				comments: []string{"// bad-tag:wow"},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FetchMetaInfo(tt.args.tag, tt.args.comments); got != tt.want {
				t.Errorf("FetchMetaInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainTag(t *testing.T) {
	type args struct {
		strs   []string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "tag found",
			args: args{
				strs:   []string{"// tag:wow"},
				prefix: "// tag",
			},
			want: true,
		},
		{
			name: "tag not found",
			args: args{
				strs:   []string{"// tag:wow"},
				prefix: "// bad-tag",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainTag(tt.args.strs, tt.args.prefix); got != tt.want {
				t.Errorf("ContainTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastWordFromName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no upper case rune in string",
			args: args{
				name: "test",
			},
			want: "test",
		},
		{
			name: "one upper case rune in string",
			args: args{
				name: "Test",
			},
			want: "test",
		},
		{
			name: "two upper case rune in string",
			args: args{
				name: "MyTest",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastWordFromName(tt.args.name); got != tt.want {
				t.Errorf("LastWordFromName() = %v, want %v", got, tt.want)
			}
		})
	}
}
