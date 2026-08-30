package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type openAPISpec struct {
	OpenAPI    string                    `json:"openapi"`
	Info       map[string]string         `json:"info"`
	Paths      map[string]map[string]any `json:"paths"`
	Components map[string]any            `json:"components,omitempty"`
}

func withOpenAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/openapi.json" {
			writeOpenAPI(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeOpenAPI(w http.ResponseWriter) {
	spec := openAPISpec{
		OpenAPI: "3.1.0",
		Info: map[string]string{
			"title":   "PVNaive API",
			"version": "r1",
		},
		Paths: make(map[string]map[string]any),
		Components: map[string]any{
			"securitySchemes": map[string]any{
				"sessionCookie": map[string]any{
					"type": "apiKey", "in": "cookie", "name": "__Host-pvnaive_session",
				},
			},
		},
	}
	for _, route := range Routes {
		// Route.Ready is the current implementation gate. system.status is
		// delivered by this WS4 extraction and is included explicitly while
		// the historical route registry still carries Ready=false.
		if !route.Ready && route.Name != "system.status" {
			continue
		}
		methods := spec.Paths[route.Path]
		if methods == nil {
			methods = make(map[string]any)
			spec.Paths[route.Path] = methods
		}
		operation := map[string]any{
			"operationId": route.Name,
			"responses": map[string]any{
				"200": map[string]any{"description": "Success"},
			},
			"x-pvnaive-access": string(route.Access),
		}
		if route.Access != Public {
			operation["security"] = []map[string][]string{{"sessionCookie": {}}}
		}
		methods[strings.ToLower(route.Method)] = operation
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = json.NewEncoder(w).Encode(spec)
}
