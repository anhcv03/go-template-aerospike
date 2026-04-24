package db

import (
	"testing"

	as "github.com/aerospike/aerospike-client-go/v8"
	astypes "github.com/aerospike/aerospike-client-go/v8/types"
)

func TestBuildCreateSecondaryIndexCommand(t *testing.T) {
	command := buildCreateSecondaryIndexCommand("track_manager", "tracks", "tracks_tid_idx", "track_id", as.NUMERIC)

	expected := "sindex-create:ns=track_manager;set=tracks;indexname=tracks_tid_idx;indexdata=track_id,NUMERIC"
	if command != expected {
		t.Fatalf("unexpected command: got %q want %q", command, expected)
	}
}

func TestSecondaryIndexReadyResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		ready    bool
		wantErr  bool
	}{
		{
			name:     "ready when load complete",
			response: "load_pct=100",
			ready:    true,
		},
		{
			name:     "not ready when load incomplete",
			response: "load_pct=42",
			ready:    false,
		},
		{
			name:     "not ready when index not found",
			response: "FAIL:201",
			ready:    false,
		},
		{
			name:     "not ready when index not readable",
			response: "FAIL:203",
			ready:    false,
		},
		{
			name:     "error on generic failure",
			response: "FAIL:4:bad parameter",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready, err := secondaryIndexReadyResponse(tt.response)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for response %q", tt.response)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ready != tt.ready {
				t.Fatalf("unexpected ready state: got %v want %v", ready, tt.ready)
			}
		})
	}
}

func TestInfoResponseMatchesCode(t *testing.T) {
	tests := []struct {
		name     string
		response string
		code     astypes.ResultCode
		want     bool
	}{
		{
			name:     "matches numeric code",
			response: "FAIL:200",
			code:     astypes.INDEX_FOUND,
			want:     true,
		},
		{
			name:     "matches symbolic code",
			response: "error=INDEX_NOTREADABLE",
			code:     astypes.INDEX_NOTREADABLE,
			want:     true,
		},
		{
			name:     "does not match different code",
			response: "FAIL:201",
			code:     astypes.INDEX_FOUND,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := infoResponseMatchesCode(tt.response, tt.code); got != tt.want {
				t.Fatalf("unexpected match result: got %v want %v", got, tt.want)
			}
		})
	}
}
