package api

import (
	"testing"
)

func TestRewritePageAssetReferences(t *testing.T) {
	input := `<html><head><link rel="icon" href="favicon.ico"><style>body { background: url("bg.png"); }</style></head><body><img src="logo.png"><img src="https://example.com/logo.png"></body></html>`
	got := rewritePageAssetReferences(input, "page-test", "index.html")

	want := `<html><head><link rel="icon" href="/static/page-assets/page-test/favicon.ico"><style>body { background: url("/static/page-assets/page-test/bg.png"); }</style></head><body><img src="/static/page-assets/page-test/logo.png"><img src="https://example.com/logo.png"></body></html>`
	if got != want {
		t.Fatalf("unexpected rewrite\n got: %s\nwant: %s", got, want)
	}
}
