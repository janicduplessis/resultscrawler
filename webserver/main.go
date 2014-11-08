package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/context"

	"github.com/janicduplessis/resultscrawler/config"
	"github.com/janicduplessis/resultscrawler/lib"
	"github.com/janicduplessis/resultscrawler/webserver/webserver"
)

func main() {
	log.SetFlags(log.Lshortfile)

	envConfig := flag.Bool("useenv", false, "Use environnement variables config")
	flag.Parse()

	// Default config
	conf := &config.ServerConfig{
		ServerURL:  "localhost",
		ServerPort: "8080",
		DbName:     "resultscrawler",
		DbUser:     "resultscrawler",
		DbPassword: "***",
		DbURL:      "localhost",
		DbPort:     "7777",
	}

	if *envConfig {
		config.ReadEnv(conf)
	} else {
		config.ReadFile("webserver.config.json", conf)
	}

	config.Print(conf)

	// Inject dependencies
	crypto := lib.NewCryptoHandler(config.Config.AESSecretKey)
	store := lib.NewMongoStore()
	logger := &lib.ConsoleLogger{}

	userStore := lib.NewUserStoreHandler(store)
	ws := webserver.NewWebserverHandler(logger)

	webserver.NewResultsWebserver(ws, userStore, crypto)

	http.Handle("/", http.FileServer(http.Dir("webserver/public")))

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.Config.ServerPort), context.ClearHandler(http.DefaultServeMux)))
	log.Println("Server stopped")
}
