package pprof

import (
	"net/http"
	"net/http/pprof"
)

func InitPprof(pprofPort string) {
	if pprofPort == "" {
		pprofPort = "6060"
	}
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	http.ListenAndServe(":"+pprofPort, mux)
}
