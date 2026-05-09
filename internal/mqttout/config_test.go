package mqttout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigReadsRequiredMQTTSettings(t *testing.T) {
	path := writeEnv(t, `
MQTT_BROKER_URL=tcp://localhost:1883
MQTT_TOPIC=rss/approved
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.BrokerURL != "tcp://localhost:1883" {
		t.Fatalf("BrokerURL = %q", cfg.BrokerURL)
	}
	if cfg.Topic != "rss/approved" {
		t.Fatalf("Topic = %q", cfg.Topic)
	}
	if cfg.ClientID == "" {
		t.Fatal("ClientID is empty, want default client id")
	}
}

func TestLoadConfigIgnoresCommentsAndBlankLines(t *testing.T) {
	path := writeEnv(t, `
# local MQTT settings

MQTT_BROKER_URL=tcp://localhost:1883
MQTT_TOPIC=rss/approved
`)

	_, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
}

func TestLoadConfigRejectsMissingRequiredSettings(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "missing broker",
			content: `
MQTT_TOPIC=rss/approved
`,
			wantErr: "MQTT_BROKER_URL",
		},
		{
			name: "missing topic",
			content: `
MQTT_BROKER_URL=tcp://localhost:1883
`,
			wantErr: "MQTT_TOPIC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeEnv(t, tt.content)

			_, err := LoadConfig(path)
			if err == nil {
				t.Fatal("LoadConfig() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("LoadConfig() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigReportsMissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), ".env"))
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want read error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "read") {
		t.Fatalf("LoadConfig() error = %q, want read context", err)
	}
}

func writeEnv(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	return path
}
