package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndAggregateValidation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(p, []byte("server: {address: ':1'}\nadmin: {address: ':1'}\nroutes:\n- name: x\n  match: {path_prefix: bad}\n  upstreams: []\n  timeout: 0s\n  retries: {attempts: 99}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, e := Load(p)
	if e == nil {
		t.Fatal("expected error")
	}
	s := e.Error()
	for _, want := range []string{"addresses must differ", "path_prefix", "at least one upstream", "retry attempts"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q: %s", want, s)
		}
	}
}

func TestLoadValidConfiguration(t *testing.T) {
	p := filepath.Join(t.TempDir(), "good.yaml")
	data := "server: {address: ':8080'}\nadmin: {address: ':9090'}\nroutes:\n- name: api\n  match: {path_prefix: /api}\n  upstreams: [{id: upstream, url: http://example.com}]\n"
	if err := os.WriteFile(p, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Routes[0].Timeout <= 0 || c.Routes[0].Retries.Attempts != 1 {
		t.Fatalf("defaults were not applied")
	}
}
