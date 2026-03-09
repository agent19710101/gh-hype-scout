package config

import "testing"

func TestValidate_Table(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Run
		sec     int
		wantErr bool
	}{
		{name: "valid webhook", cfg: Run{DescWidth: 20, WatchWebhook: "https://example.com/hook"}, sec: 300, wantErr: false},
		{name: "invalid webhook", cfg: Run{DescWidth: 20, WatchWebhook: "://bad"}, sec: 300, wantErr: true},
		{name: "watch json conflict", cfg: Run{DescWidth: 20, Watch: true, JSON: true}, sec: 300, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg, tt.sec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v err=%v", tt.wantErr, err)
			}
		})
	}
}
