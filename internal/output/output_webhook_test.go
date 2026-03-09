package output

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/agent19710101/gh-hype-scout/internal/githubapi"
	"github.com/agent19710101/gh-hype-scout/internal/rank"
)

func sampleReport() DeltaReport {
	return DeltaReport{
		NewRepos: []rank.Repo{{Repo: githubapi.Repo{FullName: "org/new"}}},
		Moves:    []RankMove{{FullName: "org/a", FromRank: 2, ToRank: 1, DeltaRank: 1, DeltaStars: 5}},
	}
}

func TestSendDeltaWebhookWithClient_Table(t *testing.T) {
	now := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "success payload",
			handler: func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(b), "org/new") {
					t.Fatalf("payload missing new repo: %s", string(b))
				}
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name: "server failure retries and returns error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()
			client := &http.Client{Timeout: 1 * time.Second}
			err := SendDeltaWebhookWithClient(client, srv.URL, now, sampleReport(), WebhookOptions{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v got err=%v", tt.wantErr, err)
			}
		})
	}
}
