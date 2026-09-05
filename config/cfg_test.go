package config

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	data := strings.NewReader(`services:
  - name: service1
    replicas:
      - http://localhost:8081
      - http://localhost:8082
    strategy: round-robin
  - name: service2
    replicas:
      - http://localhost:9091
      - http://localhost:9092
    strategy: least-connections`)

	_, err := LoadConfig(data)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}


}
