package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/janicduplessis/resultscrawler/pkg/crawler"
	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/store/mongo"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
	"github.com/janicduplessis/resultscrawler/pkg/webserver"
)

const (
	configFile = "config.json"
)

type config struct {
	ServerPort           string
	ServerTLSPort        string
	Database             *tools.MongoConfig
	AESSecretKey         string // 16 bytes
	RSAPublic            string
	RSAPrivate           string
	TLSCert              string
	TLSPriv              string
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

	crawlerClient := crawler.NewClient(config.CrawlerWebserviceURL)

	server := webserver.NewWebserver(&webserver.Config{
		UserStore:          userStore,
		CrawlerConfigStore: crawlerConfigStore,
		UserResultsStore:   userResultsStore,
		RSAPublic:          []byte(config.RSAPublic),
		RSAPrivate:         []byte(config.RSAPrivate),
		CrawlerClient:      crawlerClient,
	})

	log.Println("Server started")
	server.Start(config.ServerPort, config.ServerTLSPort, config.TLSCert, config.TLSPriv)
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
	val := os.Getenv("RC_SERVER_PORT")
	if len(val) > 0 {
		config.ServerPort = val
	}
	val = os.Getenv("RC_SERVER_TLS_PORT")
	if len(val) > 0 {
		config.ServerTLSPort = val
	}
	// DB
	val = os.Getenv("RC_DB_SERVICE_HOST")
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
	// AES
	val = os.Getenv("RC_AES_SECRET_KEY")
	if len(val) > 0 {
		config.AESSecretKey = val
	}
	//RSA
	val = os.Getenv("RC_RSA_PUBLIC")
	if len(val) > 0 {
		config.RSAPublic = val
	}
	val = os.Getenv("RC_RSA_PRIVATE")
	if len(val) > 0 {
		config.RSAPrivate = val
	}
	val = os.Getenv("RC_TLS_CERT")
	if len(val) > 0 {
		config.TLSCert = val
	}
	val = os.Getenv("RC_TLS_PRIV")
	if len(val) > 0 {
		config.TLSPriv = val
	}
	// Crawler webservice
	val = os.Getenv("RC_CRAWLER_SERVICE_HOST")
	val2 = os.Getenv("RC_CRAWLER_SERVICE_PORT")
	if len(val) > 0 && len(val2) > 0 {
		config.CrawlerWebserviceURL = fmt.Sprintf("%s:%s", val, val2)
	}
}

func validateConfig(config *config) {
	// TODO: actually validate the config.
	// for now it will just get printed.
	log.Printf("AES: %v", config.AESSecretKey)
	log.Printf("db: %+v", config.Database)
	log.Printf("server port: %v", config.ServerPort)
	log.Printf("server tls port: %v", config.ServerTLSPort)
	log.Printf("crawler webservice url: %v", config.CrawlerWebserviceURL)
}
