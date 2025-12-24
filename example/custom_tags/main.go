package main

import (
	"fmt"
	"log"

	config2 "github.com/m0rjc/goconfig/example/custom_types/config"
)

func main() {
	config, err := config2.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Given this is purely a demo, we don't have to worry about printing real secrets.
	fmt.Printf(`The following configuration has been loaded:
           Base URL: %s
WhatsApp Server URL: %s
WhatsApp Auth Token: %s
 WhatsApp Challenge: %s
     My Server Port: %d
        Health Port: %d
`, config.MyBaseURL, config.WhatsAppServerUrl, config.WhatsAppAuthToken, config.WhatsAppChallenge, config.ServerPort, config.HealthPort)
}
