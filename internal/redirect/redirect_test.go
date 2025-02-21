// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type redirectResult struct {
	status int
	path   string
}

func errorResult(status int) redirectResult {
	return redirectResult{status, ""}
}

func TestRedirects(t *testing.T) {
	var tests = map[string]redirectResult{
		"/build":                         {301, "https://build.golang.org"},
		"/ref":                           {301, "/doc/#references"},
		"/doc/mem":                       {301, "/ref/mem"},
		"/doc/spec":                      {301, "/ref/spec"},
		"/foo":                           errorResult(404),
		"/blog/2011/01/json-and-go.html": {301, "/blog/json-and-go"},

		"/pkg/asn1":           {301, "/pkg/encoding/asn1/"},
		"/pkg/template/parse": {301, "/pkg/text/template/parse/"},
		"/pkg/C":              {301, "/pkg/C/"},
		"/pkg/C/":             {301, "/cmd/cgo/"},
		"/pkg/C/foo":          {301, "/cmd/cgo/"}, // historical

		"/src/pkg/foo": {301, "/src/foo"},

		"/cmd/gofix": {301, "/cmd/fix/"},

		// git commits (/change)
		// TODO: mercurial tags and LoadChangeMap.
		"/change":   {301, "https://go.googlesource.com/go"},
		"/change/a": {302, "https://go.googlesource.com/go/+/a"},

		"/issue":                    {301, "https://github.com/golang/go/issues"},
		"/issue?":                   {301, "https://github.com/golang/go/issues"},
		"/issue/1":                  {302, "https://github.com/golang/go/issues/1"},
		"/issue/new":                {302, "https://github.com/golang/go/issues/new/choose"},
		"/issue/new/":               {302, "https://github.com/golang/go/issues/new/choose"},
		"/issue/new?a=b&c=d%20&e=f": {302, "https://github.com/golang/go/issues/new?a=b&c=d%20&e=f"},
		"/issue/new/choose":         errorResult(404),
		"/issues":                   {301, "https://github.com/golang/go/issues"},
		"/issues/1":                 {302, "https://github.com/golang/go/issues/1"},
		"/issues/new":               {302, "https://github.com/golang/go/issues/new/choose"},
		"/issues/new/":              {302, "https://github.com/golang/go/issues/new/choose"},
		"/issues/new?title=pkg":     {302, "https://github.com/golang/go/issues/new?title=pkg"},
		"/issues/new/choose":        errorResult(404),
		"/issues/1/2/3":             errorResult(404),

		"/wiki/foo":  {302, "https://github.com/golang/go/wiki/foo"},
		"/wiki/foo/": {302, "https://github.com/golang/go/wiki/foo/"},

		"/design":              {301, "https://go.googlesource.com/proposal/+/master/design"},
		"/design/":             {302, "/design"},
		"/design/123-foo":      {302, "https://go.googlesource.com/proposal/+/master/design/123-foo.md"},
		"/design/text/123-foo": {302, "https://go.googlesource.com/proposal/+/master/design/text/123-foo.md"},

		"/cl/1":          {302, "https://go-review.googlesource.com/1"},
		"/cl/1/":         {302, "https://go-review.googlesource.com/1"},
		"/cl/267120043":  {302, "https://codereview.appspot.com/267120043"},
		"/cl/267120043/": {302, "https://codereview.appspot.com/267120043"},

		"/cl/1/3":   {302, "https://go-review.googlesource.com/c/1/3"},
		"/cl/c/1/3": errorResult(404),

		// Verify that we're using the Rietveld CL table:
		"/cl/152046": {302, "https://codereview.appspot.com/152046"},
		"/cl/152047": {302, "https://go-review.googlesource.com/152047"},
		"/cl/152048": {302, "https://codereview.appspot.com/152048"},

		// And verify we're using the "bigEnoughAssumeRietveld" value:
		"/cl/3999999": {302, "https://go-review.googlesource.com/3999999"},
		"/cl/4000000": {302, "https://codereview.appspot.com/4000000"},
	}

	mux := http.NewServeMux()
	Register(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	for path, want := range tests {
		if want.path != "" && want.path[0] == '/' {
			// All redirects are absolute.
			want.path = ts.URL + want.path
		}

		req, err := http.NewRequest("GET", ts.URL+path, nil)
		if err != nil {
			t.Errorf("(path: %q) unexpected error: %v", path, err)
			continue
		}

		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			t.Errorf("(path: %q) unexpected error: %v", path, err)
			continue
		}

		if resp.StatusCode != want.status {
			t.Errorf("(path: %q) got status %d, want %d", path, resp.StatusCode, want.status)
			continue
		}

		if want.status != 301 && want.status != 302 {
			// Not a redirect. Just check status.
			continue
		}

		out, _ := resp.Location()
		if got := out.String(); got != want.path {
			t.Errorf("(path: %q) got %s, want %s", path, got, want.path)
		}
	}
}
