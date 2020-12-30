package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"log"
	"os"
	"sync"
)

var c Config // global var to hold static configuration

const HIDDEN_FOLDER = ".hidden"

func main() {
	configPath := flag.String("c", "flounder.toml", "path to config file") // doesnt work atm
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("expected 'admin' or 'serve' subcommand")
		os.Exit(1)
	}

	var err error
	c, err = getConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	logFile, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	if c.HttpsEnabled {
		_, err1 := os.Stat(c.TLSCertFile)
		_, err2 := os.Stat(c.TLSKeyFile)
		if os.IsNotExist(err1) || os.IsNotExist(err2) {
			log.Fatal("Keyfile or certfile does not exist.")
		}
	}

	initializeDB()

	cookie := generateCookieKeyIfDNE()
	SessionStore = sessions.NewCookieStore(cookie)

	switch args[0] {
	case "serve":
		wg := new(sync.WaitGroup)
		wg.Add(2)
		go func() {
			runHTTPServer()
			wg.Done()
		}()
		go func() {
			runGeminiServer()
			wg.Done()
		}()
		wg.Wait()
	case "admin":
		runAdminCommand()
	}
}
