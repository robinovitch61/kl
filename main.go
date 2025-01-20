package main

import (
	"github.com/robinovitch61/kl/cmd"
	"log"
	"os"
	"runtime/pprof"
)

func main() {
	if cpuProfile := os.Getenv("KL_CPU_PROFILE"); cpuProfile != "" {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
