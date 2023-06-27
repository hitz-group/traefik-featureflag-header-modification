package traefik_featureflag_header_modification_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	traefik_featureflag_header_modification "github.com/hitz-group/traefik-featureflag-header-modification"
)

func TestXRequestStart(t *testing.T) {
	cfg := traefik_featureflag_header_modification.CreateConfig()

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := traefik_featureflag_header_modification.New(ctx, next, cfg, "traefik_featureflag_header_modification")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)
}
