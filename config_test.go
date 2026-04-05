package main

import "testing"

func TestApplyListenPort(t *testing.T) {
	got := applyListenPort("0.0.0.0:9165", 9200)
	if got != "0.0.0.0:9200" {
		t.Fatalf("unexpected listen address: %s", got)
	}
}

func TestNormalizeConfig(t *testing.T) {
	cfg := defaultConfig()
	cfg.ListenAddress = "9165"
	cfg.MetricsPath = "metrics"
	cfg.APIPassword = "secret"
	if err := cfg.normalize(); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if cfg.ListenAddress != ":9165" {
		t.Fatalf("unexpected listen address: %s", cfg.ListenAddress)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Fatalf("unexpected metrics path: %s", cfg.MetricsPath)
	}
}
