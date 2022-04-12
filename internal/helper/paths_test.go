package helper

import (
	"runtime"
	"testing"
)

func TestIsAbsolutePath(t *testing.T) {
	type args struct {
		pathname string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple root",
			args: args{pathname: "/"},
			want: true,
		},
		{
			name: "Simple drive path",
			args: args{pathname: "e:\\"},
			want: true,
		},
		{
			name: "Windows pathname",
			args: args{pathname: "g:\\a\\b\\c/d"},
			want: true,
		},
		{
			name: "relative path",
			args: args{pathname: ".\\test"},
			want: false,
		},
		{
			name: "relative #2",
			args: args{pathname: "../../.."},
			want: false,
		},
	}
	if runtime.GOOS == "windows" {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := IsAbsolutePath(tt.args.pathname); got != tt.want {
					t.Errorf("IsAbsolutePath() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}
