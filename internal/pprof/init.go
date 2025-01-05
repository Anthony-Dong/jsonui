package pprof

import (
	"net/http"
	_ "net/http/pprof"
)

// InitPProf
// go InitPProf()
func InitPProf() {
	err := http.ListenAndServe(":12345", http.DefaultServeMux)
	if err != nil {
		panic(err)
	}
}
