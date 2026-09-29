package gateway

import (
	"github.com/italo/go-api-gateway/internal/config"
	"sort"
	"strings"
)

type Router struct{ routes []config.RouteConfig }

func NewRouter(routes []config.RouteConfig) *Router {
	copyRoutes := append([]config.RouteConfig(nil), routes...)
	sort.Slice(copyRoutes, func(i, j int) bool { return len(copyRoutes[i].Match.PathPrefix) > len(copyRoutes[j].Match.PathPrefix) })
	return &Router{routes: copyRoutes}
}
func (r *Router) Match(path string) (config.RouteConfig, string, bool) {
	for _, route := range r.routes {
		p := route.Match.PathPrefix
		if path == p || strings.HasPrefix(path, p+"/") {
			target := path
			if route.Match.StripPrefix {
				target = strings.TrimPrefix(path, p)
				if target == "" {
					target = "/"
				}
				if !strings.HasPrefix(target, "/") {
					target = "/" + target
				}
			}
			return route, target, true
		}
	}
	return config.RouteConfig{}, "", false
}
