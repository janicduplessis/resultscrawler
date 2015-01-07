package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/janicduplessis/resultscrawler/pkg/crawler"
	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/mongo"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
)

type config struct {
	Database       *tools.MongoConfig
	Email          *tools.EmailConfig
	AESSecretKey   string // 16 bytes
	WebservicePort string
}

const (
	configFile  = "config.json"
	numCrawlers = 10
)

var (
	webservicePort = flag.String("port", "", "Webservice port")
	dbURL          = flag.String("db-url", "", "DB url")
	dbUser         = flag.String("db-user", "", "DB user")
	dbPassword     = flag.String("db-password", "", "DB password")
	dbName         = flag.String("db-name", "", "DB name")
	emailURL       = flag.String("email-url", "", "Email url")
	emailUser      = flag.String("email-user", "", "Email user")
	emailPassword  = flag.String("email-password", "", "Email password")
	aesSecret      = flag.String("aes-secret", "", "AES secret key")
)

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
	readFlagConfig(conf)
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
	val := os.Getenv("RC_DB_SERVICE_HOST")
	val2 := os.Getenv("RC_DB_SERVICE_PORT")
	if len(val) > 0 && len(val2) > 0 {
		config.Database.URL = fmt.Sprintf("%s:%s", val, val2)
	}
	val = os.Getenv("RC_DB_USER")
	if len(val) > 0 {
		config.Database.User = val
	}
	val = os.Getenv("RC_DB_PASSWORD")
	if len(val) > 0 {
		config.Database.Password = val
	}
	val = os.Getenv("RC_DB_NAME")
	if len(val) > 0 {
		config.Database.Name = val
	}
	// Email
	val = os.Getenv("RC_EMAIL_SERVICE_HOST")
	val2 = os.Getenv("RC_EMAIL_SERVICE_PORT")
	if len(val) > 0 && len(val2) > 0 {
		config.Email.URL = fmt.Sprintf("%s:%s", val, val2)
	}
	val = os.Getenv("RC_EMAIL_USER")
	if len(val) > 0 {
		config.Email.User = val
	}
	val = os.Getenv("RC_EMAIL_PASSWORD")
	if len(val) > 0 {
		config.Email.Password = val
	}
	// AES
	val = os.Getenv("RC_AES_SECRET_KEY")
	if len(val) > 0 {
		config.AESSecretKey = val
	}
	// Webservice
	val = os.Getenv("RC_CRAWLER_PORT")
	if len(val) > 0 {
		config.WebservicePort = val
	}
}

func readFlagConfig(config *config) {
	// DB
	val := *dbURL
	if len(val) > 0 {
		config.Database.URL = val
	}
	val = *dbUser
	if len(val) > 0 {
		config.Database.User = val
	}
	val = *dbPassword
	if len(val) > 0 {
		config.Database.Password = val
	}
	val = *dbName
	if len(val) > 0 {
		config.Database.Name = val
	}
	// Email
	val = *emailURL
	if len(val) > 0 {
		config.Email.URL = val
	}
	val = *emailUser
	if len(val) > 0 {
		config.Email.User = val
	}
	val = *emailPassword
	if len(val) > 0 {
		config.Email.Password = val
	}
	// AES
	val = *aesSecret
	if len(val) > 0 {
		config.AESSecretKey = val
	}
	// Webservice
	val = *webservicePort
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
