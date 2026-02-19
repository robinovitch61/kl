package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // register pprof endpoints
	"os"

	"github.com/robinovitch61/kl/cmd"
)

func main() {
	if pprofServer := os.Getenv("KL_PPROF_SERVER"); pprofServer != "" {
		port := "6060"
		go func() {
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				panic(fmt.Errorf("pprof server failed: %v", err))
			}
		}()
	}

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
