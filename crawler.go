package main

import (
	"flag"
	"log"

	"github.com/janicduplessis/resultscrawler/config"
	"github.com/janicduplessis/resultscrawler/crawler"
	"github.com/janicduplessis/resultscrawler/lib"
)

func main() {
	log.SetFlags(log.Lshortfile)

	envConfig := flag.Bool("useenv", false, "Use environnement variables config")
	flag.Parse()

	// Default config
	conf := &config.ServerConfig{
		ServerURL:  "localhost",
		ServerPort: "9898",
		DbName:     "resultscrawler",
		DbUser:     "resultscrawler",
		DbPassword: "***",
		DbURL:      "localhost",
		DbPort:     "7777",
	}

	if *envConfig {
		config.ReadEnv(conf)
	} else {
		config.ReadFile("crawler.json", conf)
	}

	config.Print(conf)

	// Inject dependencies
	crypto := lib.NewCryptoHandler(config.Config.AESSecretKey)
	emailSender := &lib.EmailSender{}
	store := lib.NewMongoStore()

	userStore := lib.NewUserStoreHandler(store)

	crawlers := []*crawler.Crawler{
		crawler.NewCrawler(crypto),
	}
	scheduler := crawler.NewScheduler(crawlers, userStore, emailSender)

	log.Println("Server started")
	scheduler.Start()
	log.Println("Server stopped")
}
