package main

import (
	"log"

	"github.com/janicduplessis/resultscrawler/config"
	"github.com/janicduplessis/resultscrawler/crawler"
	"github.com/janicduplessis/resultscrawler/lib"
)

func main() {
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

	config.ReadFile("crawler.json", conf)
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
