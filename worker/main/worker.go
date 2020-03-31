package main

import (
	"flag"
	"log"
	"runtime"
	"time"

	"github.com/s1m0n21/go-crontab/worker"
)

var configFile string

func initArgs() {
	flag.StringVar(&configFile, "c", "./worker.json", "Specifying a configuration file.")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	initEnv()

	initArgs()

	if err := worker.InitConfig(configFile); err != nil {
		log.Fatal(err)
	}

	if err := worker.InitScheduler(); err != nil {
		log.Fatal(err)
	}

	if err := worker.InitJobMgr(); err != nil {
		log.Fatal(err)
	}

	if err := worker.InitExecutor(); err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(10 * time.Second)
	}
}
