package ants

import (
	"reflect"
	"testing"
)

func Test_parseAgentVersion(t *testing.T) {
	tests := []struct {
		av   string
		want agentVersionInfo
	}{
		{
			av:   "",
			want: agentVersionInfo{},
		},
		{
			av: "celestia-node/celestia/bridge/v0.17.1/078c291",
			want: agentVersionInfo{
				full:  "celestia-node/celestia/bridge/v0.17.1/078c291",
				typ:   "bridge",
				major: 0,
				minor: 17,
				patch: 1,
				hash:  "078c291",
			},
		},
		{
			av: "celestia-node/celestia/full/v0.17.2/57f8bd8",
			want: agentVersionInfo{
				full:  "celestia-node/celestia/full/v0.17.2/57f8bd8",
				typ:   "full",
				major: 0,
				minor: 17,
				patch: 2,
				hash:  "57f8bd8",
			},
		},
		{
			av: "celestia-node/celestia/random/v4.46.6/57f8bd8",
			want: agentVersionInfo{
				full:  "celestia-node/celestia/random/v4.46.6/57f8bd8",
				typ:   "other",
				major: 4,
				minor: 46,
				patch: 6,
				hash:  "57f8bd8",
			},
		},
		{
			av: "celestia-node/celestia/light/vv0.14.0/13439cc",
			want: agentVersionInfo{
				full:  "celestia-node/celestia/light/vv0.14.0/13439cc",
				typ:   "light",
				major: 0,
				minor: 14,
				patch: 0,
				hash:  "13439cc",
			},
		},
		{
			av: "celestia-node/celestia/light/v0.20.3-15-gbd3105b9/bd3105b",
			want: agentVersionInfo{
				full:  "celestia-node/celestia/light/v0.20.3-15-gbd3105b9/bd3105b",
				typ:   "light",
				major: 0,
				minor: 20,
				patch: 3,
				hash:  "bd3105b",
			},
		},
		{
			av: "celestia-node/celestia/full/v0.18.0-refs-tags-v0-20-1-mocha.0/353141f",
			want: agentVersionInfo{
				full:  "celestia-node/celestia/full/v0.18.0-refs-tags-v0-20-1-mocha.0/353141f",
				typ:   "full",
				major: 0,
				minor: 18,
				patch: 0,
				hash:  "353141f",
			},
		},
		{
			av: "celestia-node/celestia/light/unknown/unknown",
			want: agentVersionInfo{
				full: "celestia-node/celestia/light/unknown/unknown",
				typ:  "light",
			},
		},
		{
			av: "celestia-celestia",
			want: agentVersionInfo{
				full: "celestia-celestia",
				typ:  "celestia-celestia",
			},
		},
		{
			av: "celestiant",
			want: agentVersionInfo{
				full: "celestiant",
				typ:  "other",
			},
		},
		{
			av: "lumina/mainnet/0.3.5",
			want: agentVersionInfo{
				full:  "lumina/mainnet/0.3.5",
				typ:   "lumina",
				major: 0,
				minor: 3,
				patch: 5,
				hash:  "",
			},
		},
		{
			av: "lumina/mainnet/",
			want: agentVersionInfo{
				full: "lumina/mainnet/",
				typ:  "lumina",
			},
		},
		{
			av: "lumina/mainnet/unknown",
			want: agentVersionInfo{
				full: "lumina/mainnet/unknown",
				typ:  "lumina",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.av, func(t *testing.T) {
			if got := parseAgentVersion(tt.av); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAgentVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
