package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndAggregateValidation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.yaml")
	os.WriteFile(p, []byte("server: {address: ':1'}\nadmin: {address: ':1'}\nroutes:\n- name: x\n  match: {path_prefix: bad}\n  upstreams: []\n  timeout: 0s\n  retries: {attempts: 99}\n"), 0600)
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
