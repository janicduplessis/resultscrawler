package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/janicduplessis/resultscrawler/crawler/crawler"
	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	configFile = "crawler.config.json"
)

type config struct {
	Database     *lib.DBConfig
	Email        *lib.EmailConfig
	AESSecretKey string // 16 bytes
}

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()
	config := readConfig()

	// Inject dependencies
	crypto := lib.NewCryptoHandler(config.AESSecretKey)
	emailSender := lib.NewEmailSender(config.Email)
	store := lib.NewMongoStore(config.Database)

	userStore := lib.NewUserStoreHandler(store)

	crawlers := []*crawler.Crawler{
		crawler.NewCrawler(crypto),
	}
	scheduler := crawler.NewScheduler(crawlers, userStore, emailSender)

	log.Println("Server started")
	scheduler.Start()
	log.Println("Server stopped")
}

func readConfig() *config {
	conf := &config{
		Database: new(lib.DBConfig),
		Email:    new(lib.EmailConfig),
	}

	readFileConfig(conf)
	readEnvConfig(conf)

	return conf
}

func readFileConfig(config *config) {
	// Get server config
	file, err := ioutil.ReadFile(configFile)

	// return if no config files
	if err != nil {
		return
	}

	if err = json.Unmarshal(file, config); err != nil {
		log.Fatal(err)
	}
}

func readEnvConfig(config *config) {
	// DB
	val := os.Getenv("CRAWLER_DB_HOST")
	if len(val) > 0 {
		config.Database.Host = val
	}
	val = os.Getenv("CRAWLER_DB_PORT")
	if len(val) > 0 {
		config.Database.Port = val
	}
	val = os.Getenv("CRAWLER_DB_USER")
	if len(val) > 0 {
		config.Database.User = val
	}
	val = os.Getenv("CRAWLER_DB_PASSWORD")
	if len(val) > 0 {
		config.Database.Password = val
	}
	val = os.Getenv("CRAWLER_DB_NAME")
	if len(val) > 0 {
		config.Database.Name = val
	}
	// Email
	val = os.Getenv("CRAWLER_EMAIL_HOST")
	if len(val) > 0 {
		config.Email.Host = val
	}
	val = os.Getenv("CRAWLER_EMAIL_PORT")
	if len(val) > 0 {
		config.Email.Port = val
	}
	val = os.Getenv("CRAWLER_EMAIL_USER")
	if len(val) > 0 {
		config.Email.User = val
	}
	val = os.Getenv("CRAWLER_EMAIL_PASSWORD")
	if len(val) > 0 {
		config.Email.Password = val
	}
	// AES
	val = os.Getenv("CRAWLER_AES_SECRET_KEY")
	if len(val) > 0 {
		config.AESSecretKey = val
	}
}
