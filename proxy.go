package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
)

//ServiceConfig ...
type ServiceConfig struct {
	Backend string
	Path    string
	Methods []string
	UseJWT  bool
}

//CreateProxy ...
func (sc *ServiceConfig) CreateProxy(r *chi.Mux) {

	matcher := NewListenPathMatcher()
	proxy := NewReverseProxy(sc)

	for _, method := range sc.Methods {

		if strings.HasSuffix(sc.Path, "/*") {
			r.Method(method, matcher.Extract(sc.Path), proxy)
		}

		r.Method(method, sc.Path, proxy)
	}
}

//NewReverseProxy ...
func NewReverseProxy(conf *ServiceConfig) http.HandlerFunc {

	target, _ := url.Parse(conf.Backend)
	targetQuery := target.RawQuery
	matcher := NewListenPathMatcher()

	return func(w http.ResponseWriter, r *http.Request) {

		d := func(req *http.Request) {

			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path + strings.Replace(r.URL.Path, matcher.Extract(conf.Path), "", 1)

			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}

			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}

			req.Header.Set("X-Proxied-By", "Stuwilli")
		}

		p := &httputil.ReverseProxy{Director: d}
		p.ServeHTTP(w, r)
	}
}
