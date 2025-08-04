package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func callerHelper() string {
	return internal.GetCaller()
}

func anotherHelper(helperFunc func() string) string {
	return helperFunc()
}

func TestGetCaller(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "direct_call",
			fn: func() string {
				return internal.GetCaller()
			},
			want: "func5",
		},
		{
			name: "nested_call",
			fn: func() string {
				return callerHelper()
			},
			want: "func2",
		},
		{
			name: "anonymous_function",
			fn: func() string {
				anonymous := func() string {
					return internal.GetCaller()
				}
				return anonymous()
			},
			want: "func3",
		},
		{
			name: "another_helper_function",
			fn: func() string {
				return anotherHelper(callerHelper)
			},
			want: "anotherHelper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fn())
		})
	}
}

func TestCallerFuncName(t *testing.T) {
	tests := []struct {
		name string
		f    *runtime.Func
		want string
	}{
		{
			name: "nil_function",
			f:    nil,
			want: "",
		},
		{
			name: "valid_function",
			f:    runtime.FuncForPC((uintptr)(1)), // Replace with actual function's PC for testing, if necessary
			want: "",                              // Replace with the expected function name
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, internal.CallerFuncName(tt.f))
		})
	}
}

func TestCallerName(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     string
	}{
		{
			name:     "empty_input",
			funcName: "",
			want:     "",
		},
		{
			name:     "simple_function_name",
			funcName: "main",
			want:     "main",
		},
		{
			name:     "nested_function",
			funcName: "pkg.subpkg.funcName",
			want:     "funcName",
		},
		{
			name:     "single_dot_input",
			funcName: ".",
			want:     "",
		},
		{
			name:     "no_dot_in_function_name",
			funcName: "functionName",
			want:     "functionName",
		},
		{
			name:     "trailing_dot",
			funcName: "pkg.funcName.",
			want:     "",
		},
		{
			name:     "multiple_dots",
			funcName: "pkg.subpkg.subsubpkg.funcName",
			want:     "funcName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, internal.CallerName(tt.funcName))
		})
	}
}
