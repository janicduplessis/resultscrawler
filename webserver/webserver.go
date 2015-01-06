package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/mongo"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
	"github.com/janicduplessis/resultscrawler/pkg/webserver"
)

const (
	configFile = "config.json"
)

var (
	envPort = flag.String("port", "8080", "Webserver port")
)

type config struct {
	ServerPort           string
	Database             *tools.MongoConfig
	AESSecretKey         string // 16 bytes
	SessionKey           string
	CrawlerWebserviceURL string
}

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()
	config := readConfig()
	crypto.Init(config.AESSecretKey)

	// Inject dependencies
	mongoHelper := tools.NewMongoHelper(config.Database)

	userStore := mongo.New(mongoHelper)
	crawlerConfigStore := mongo.New(mongoHelper)
	userResultsStore := mongo.New(mongoHelper)

	server := webserver.NewWebserver(&webserver.Config{
		UserStore:            userStore,
		CrawlerConfigStore:   crawlerConfigStore,
		UserResultsStore:     userResultsStore,
		SessionKey:           config.SessionKey,
		CrawlerWebserviceURL: config.CrawlerWebserviceURL,
	})

	log.Println("Server started")
	log.Fatal(server.Start(fmt.Sprintf(":%s", config.ServerPort)))
	log.Println("Server stopped")
}

func readConfig() *config {
	conf := &config{
		Database: new(tools.MongoConfig),
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
	val := os.Getenv("CRAWLERSERVER_SERVER_PORT")
	if len(val) > 0 {
		config.ServerPort = val
	}
	// DB
	val = os.Getenv("CRAWLERSERVER_DB_URL")
	if len(val) > 0 {
		config.Database.URL = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_USER")
	if len(val) > 0 {
		config.Database.User = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_PASSWORD")
	if len(val) > 0 {
		config.Database.Password = val
	}
	val = os.Getenv("CRAWLERSERVER_DB_NAME")
	if len(val) > 0 {
		config.Database.Name = val
	}
	// AES
	val = os.Getenv("CRAWLERSERVER_AES_SECRET_KEY")
	if len(val) > 0 {
		config.AESSecretKey = val
	}
	val = os.Getenv("CRAWLERSERVER_SESSION_KEY")
	if len(val) > 0 {
		config.SessionKey = val
	}
	// Crawler webservice
	val = os.Getenv("CRAWLERSERVER_CRAWLER_WEBSERVICE_URL")
	if len(val) > 0 {
		config.CrawlerWebserviceURL = val
	}
}

func validateConfig(config *config) {
	// TODO: actually validate the config.
	// for now it will just get printed.
	log.Printf("AES: %v", config.AESSecretKey)
	log.Printf("db: %+v", config.Database)
	log.Printf("server port: %v", config.ServerPort)
	log.Printf("crawler webservice url: %v", config.CrawlerWebserviceURL)
	log.Printf("session key: %v", config.SessionKey)
}
