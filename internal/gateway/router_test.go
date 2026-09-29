package gateway

import (
	"github.com/italo/go-api-gateway/internal/config"
	"testing"
)

func TestRouterPrefersLongestPrefixAndStrips(t *testing.T) {
	r := NewRouter([]config.RouteConfig{{Name: "root", Match: config.MatchConfig{PathPrefix: "/orders"}}, {Name: "admin", Match: config.MatchConfig{PathPrefix: "/orders/admin", StripPrefix: true}}})
	got, path, ok := r.Match("/orders/admin/users")
	if !ok || got.Name != "admin" || path != "/users" {
		t.Fatalf("got %#v %q %v", got, path, ok)
	}
}
func BenchmarkRouterMatch(b *testing.B) {
	r := NewRouter([]config.RouteConfig{{Name: "a", Match: config.MatchConfig{PathPrefix: "/a"}}, {Name: "b", Match: config.MatchConfig{PathPrefix: "/a/b/c"}}})
	for b.Loop() {
		r.Match("/a/b/c/123")
	}
}
