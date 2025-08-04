package internal_test

import (
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyResponseWriter_Header(t *testing.T) {
	tests := []struct {
		name string
		want http.Header
	}{
		{
			name: "Header always returns nil",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := internal.DummyResponseWriter{}
			got := rw.Header()
			assert.Equal(t, tt.want, got)
		})
	}
}
