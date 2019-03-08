package main

import (
	"flag"
	"fmt"
	"os"
)

// AppName exposes application name globally
var AppName = "patroniglue"

// AppVersion stores application version at compilation time
var AppVersion string

func main() {

	var err error
	config := NewConfig()

	// Argument handling
	quiet := flag.Bool("quiet", false, "Quiet mode")
	verbose := flag.Bool("verbose", false, "Verbose mode")
	debug := flag.Bool("debug", false, "Debug mode")
	version := flag.Bool("version", false, "Print version")
	flag.StringVar(&config.File, "config", os.Getenv("HOME")+"/."+AppName+".yml", "Configuration file")
	flag.Parse()

	// Print version and exit
	if *version {
		if AppVersion == "" {
			AppVersion = "unknown"
		}
		fmt.Println(AppVersion)
		return
	}

	// Log level management
	if *debug {
		err = SetLogLevel("DEBUG")
	}
	if *verbose {
		err = SetLogLevel("INFO")
	}
	if *quiet {
		err = SetLogLevel("ERROR")
	}

	if err != nil {
		Fatal("could not set log level: %v", err)
	}

	// Read configuration file
	Debug("reading configuration file")
	err = config.ReadFile(config.File)
	if err != nil {
		Fatal("could not read configuration file: %v", err)
	}

	// Cache management
	cache, err := NewCache(config.Cache)
	if err != nil {
		Fatal("could not create cache: %v", err)
	}
	cache.Startup()

	// Backend management
	backend := NewBackend(config.Backend, cache)

	// Frontend management
	frontend, err := NewFrontend(config.Frontend, backend)
	if err != nil {
		Fatal("could not create frontend: %v", err)
	}
	err = frontend.Start()
	if err != nil {
		Fatal("could not start frontend: %v", err)
	}
}
