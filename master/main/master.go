package main

import (
	"flag"
	"log"
	"runtime"

	"github.com/s1m0n21/go-crontab/master"
)

var configFile string

func initArgs() {
	flag.StringVar(&configFile, "c", "./master.json", "Specifying a configuration file.")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	initEnv()

	initArgs()

	if err := master.InitConfig(configFile); err != nil {
		log.Fatal(err)
	}

	if err := master.InitJobMgr(); err != nil {
		log.Fatal(err)
	}

	if err := master.InitAPI(); err != nil {
		log.Fatal(err)
	}
}
