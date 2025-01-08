package ants

import (
	"fmt"
	"testing"
)

func TestQueenConfig_RangesOverlap(t *testing.T) {
	tests := []struct {
		FirstPort    int
		FirstPortWSS int
		NPorts       int
		want         bool
	}{
		{FirstPort: 6000, FirstPortWSS: 6128, NPorts: 128, want: false},
		{FirstPort: 6000, FirstPortWSS: 6064, NPorts: 128, want: true},
		{FirstPort: 6064, FirstPortWSS: 6000, NPorts: 128, want: true},
		{FirstPort: 6128, FirstPortWSS: 6000, NPorts: 128, want: false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("Standard %d, wss %d, nports %d", tt.FirstPort, tt.FirstPortWSS, tt.NPorts), func(t *testing.T) {
			cfg := &QueenConfig{
				NPorts:       tt.NPorts,
				FirstPort:    tt.FirstPort,
				FirstPortWSS: tt.FirstPortWSS,
			}
			if got := cfg.RangesOverlap(); got != tt.want {
				t.Errorf("RangesOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}
