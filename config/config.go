// Package config handles an application configuration and provides
// a global object to read it. The configuration can be provided by
// a json config file or environment variables.
package config

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"os"
	"reflect"
)

// Config the public global config object
var Config *ServerConfig

// ServerConfig the config object
type ServerConfig struct {
	ServerURL  string
	ServerPort string

	DbUser     string
	DbPassword string
	DbName     string
	DbURL      string
	DbPort     string

	EmailUser     string
	EmailPassword string
	EmailHost     string
	EmailPort     string

	AESSecretKey string // 16 bytes
}

// ReadEnv sets the values of fields in obj using the env variables with the same name
func ReadEnv(obj *ServerConfig) {
	s := reflect.ValueOf(obj).Elem()
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		config := os.Getenv(typeOf.Field(i).Name)
		if len(config) > 0 {
			configVal := reflect.ValueOf(config)
			if configVal.Type().ConvertibleTo(f.Type()) {
				f.Set(configVal.Convert(f.Type()))
			}
		}
	}

	Config = obj
}

// ReadFile sets the values of fields in obj using a json formatted config file
func ReadFile(configFile string, obj *ServerConfig) {
	// Get server config
	file, err := ioutil.ReadFile(configFile)

	// return if no config files
	if err != nil {
		return
	}

	if err = json.Unmarshal(file, &obj); err != nil {
		log.Fatal(err)
	}

	Config = obj
}

func ValidateConfig(obj *ServerConfig) {

}

// Print prints the config to std output
func Print(config *ServerConfig) {
	// Print the config
	log.Println("---------------------")
	log.Println("-     Config        -")
	log.Println("---------------------")

	s := reflect.ValueOf(config).Elem()
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		log.Println(fmt.Sprintf("%s: %v", typeOf.Field(i).Name, f.Interface()))
	}

	log.Println("---------------------")
}
