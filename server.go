package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
)

import (
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/testutils"
	"golang.org/x/crypto/acme/autocert"
)

func InitServer() error {
	m := autocert.Manager{
		Cache:  autocert.DirCache(*STORAGE),
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			if _, ok := HOSTS[host]; ok {
				return nil
			}
			return errors.New("Unkown host(" + host + ")")
		},
	}
	s := &http.Server{
		Addr:      ":https",
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   ServeHTTP(),
	}
	return s.ListenAndServeTLS("", "")
}

func ServeHTTP() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if upstreams, ok := HOSTS[req.Host]; ok {
			forwarder, _ := forward.New()
			loadbalancer, _ := roundrobin.New(forwarder)
			for _, upstream := range upstreams {
				loadbalancer.UpsertServer(testutils.ParseURI(upstream))
			}
			loadbalancer.ServeHTTP(res, req)
		}
		http.Error(res, "The request domain couldn't be found here", http.StatusNotImplemented)
	})
}
