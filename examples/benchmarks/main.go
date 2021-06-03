package main

import (
	"flag"
	"log"
)

var engines = map[string]Engine{}

func register(engine Engine) { engines[engine.Name()] = engine }

var (
	cmdEngineFlag   = flag.String("e", "sha", "engine name, default is `sha`")
	cmdTestTypeFlag = flag.String("t", "helloworld", "test type, default is `helloworld`")
	cmdAddressFlag  = flag.String("a", "127.0.0.1:8080", "listen address, default is `127.0.0.1:8080`")
)

func main() {
	flag.Parse()

	engine := engines[*cmdEngineFlag]
	if engine == nil {
		log.Fatalf("unknown engine: `%s`\r\n", *cmdEngineFlag)
	}

	switch *cmdTestTypeFlag {
	case "helloworld":
		engine.HelloWorld(*cmdAddressFlag)
	}
}
