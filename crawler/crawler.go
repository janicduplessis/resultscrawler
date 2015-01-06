package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/janicduplessis/resultscrawler/pkg/crawler"
	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/mongo"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
)

const (
	configFile  = "config.json"
	numCrawlers = 10
)

type config struct {
	Database       *tools.MongoConfig
	Email          *tools.EmailConfig
	AESSecretKey   string // 16 bytes
	WebservicePort string
}

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()
	config := readConfig()

	crypto.Init(config.AESSecretKey)

	// Inject dependencies
	emailSender := tools.NewEmailSender(config.Email)
	mongoHelper := tools.NewMongoHelper(config.Database)
	httpClient := &http.Client{}

	userStore := mongo.New(mongoHelper)
	crawlerConfigStore := mongo.New(mongoHelper)
	userResultsStore := mongo.New(mongoHelper)

	var crawlers []crawler.ResultGetter
	for i := 0; i < numCrawlers; i++ {
		crawlers = append(crawlers, crawler.NewCrawler(httpClient))
	}

	scheduler := crawler.NewScheduler(&crawler.SchedulerConfig{
		ResultGetters:      crawlers,
		UserStore:          userStore,
		CrawlerConfigStore: crawlerConfigStore,
		UserResultsStore:   userResultsStore,
		Sender:             emailSender,
	})

	crawler.StartWebservice(scheduler, userStore, config.WebservicePort)

	log.Println("Crawler started")
	scheduler.Start()
	log.Println("Crawler stopped")
}

func readConfig() *config {
	conf := &config{
		Database: new(tools.MongoConfig),
		Email:    new(tools.EmailConfig),
	}

	readFileConfig(conf)
	readEnvConfig(conf)
	validateConfig(conf)

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
	val := os.Getenv("CRAWLER_DB_URL")
	if len(val) > 0 {
		config.Database.URL = val
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
	val = os.Getenv("CRAWLER_EMAIL_URL")
	if len(val) > 0 {
		config.Email.URL = val
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
	// Webservice
	val = os.Getenv("CRAWLER_WEBSERVICE_PORT")
	if len(val) > 0 {
		config.WebservicePort = val
	}
}

func validateConfig(config *config) {
	// TODO: actually validate the config.
	// for now it will just get printed.
	log.Printf("AES: %v", config.AESSecretKey)
	log.Printf("db: %+v", config.Database)
	log.Printf("email: %+v", config.Email)
	log.Printf("webservice port: %v", config.WebservicePort)
}
