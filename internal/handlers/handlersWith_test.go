package handlers

import (
	"github.com/sergeysynergy/metricser/internal/data/repository/memory"
	"github.com/sergeysynergy/metricser/internal/storage"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestWithTrustedSubnet(t *testing.T) {
	getNet := func(s string) *net.IPNet {
		_, ipNet, _ := net.ParseCIDR(s)
		return ipNet
	}

	type want struct {
		net *net.IPNet
	}
	tests := []struct {
		name string
		cidr string
		want want
	}{
		{
			name: "Ok",
			cidr: "127.0.0.0/24",
			want: want{
				net: getNet("127.0.0.0/24"),
			},
		},
		{
			name: "bad ip",
			cidr: "bad bad ip/24",
			want: want{
				net: &net.IPNet{},
			},
		},
		{
			name: "bad mask",
			cidr: "127.0.0.0/bad bad mask",
			want: want{
				net: &net.IPNet{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := New(storage.New(memory.New(), nil),
				WithTrustedSubnet(tt.cidr),
			)
			assert.Equal(t, tt.want.net, handler.trustedSubnet)
		})
	}
}
