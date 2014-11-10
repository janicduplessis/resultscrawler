package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/janicduplessis/resultscrawler/config"
	"github.com/janicduplessis/resultscrawler/lib"
	"github.com/janicduplessis/resultscrawler/webserver/webserver"
)

func main() {
	log.SetFlags(log.Lshortfile)

	envConfig := flag.Bool("useenv", false, "Use environnement variables config")
	envPort := flag.String("port", "8080", "Webserver port")
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
		conf.ServerPort = *envPort
	} else {
		config.ReadFile("webserver.config.json", conf)
	}

	config.Print(conf)

	// Inject dependencies
	crypto := lib.NewCryptoHandler(config.Config.AESSecretKey)
	store := lib.NewMongoStore()
	logger := &lib.ConsoleLogger{}

	userStore := lib.NewUserStoreHandler(store)
	ws := webserver.NewWebserver(logger)

	webserver.NewResultsWebserver(ws, userStore, crypto)

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.Config.ServerPort), nil))
	log.Println("Server stopped")
}
