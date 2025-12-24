package config

import (
	"context"
	"net/url"

	"github.com/m0rjc/goconfig"
)

type Config struct {
	// MyBaseURL is the base URL for my web interface. This is used when sharing links in messages to direct the user
	// to my web interface.
	MyBaseURL *url.URL `key:"MY_BASE_URL" required:"false" default:"http://localhost:8080"`
	// WhatsAppPhoneId is the phone number ID for my phone number
	WhatsAppPhoneId string `key:"WHATSAPP_PHONE_ID" required:"true" pattern:"^[0-9]+$"`
	// WhatsAppServerUrl is the URL of the WhatsApp Business API server. This uses the custom "secure" tag defined by this example
	WhatsAppServerUrl *url.URL `key:"WHATSAPP_SERVER_URL" required:"true" secure:"true" default:"https://api.whatsapp.com"`
	// WhatsAppAuthToken is the authentication token for the WhatsApp Business API.
	WhatsAppAuthToken string `key:"WHATSAPP_AUTH_TOKEN" required:"true"`
	// WhatsAppChallenge is the challenge token sent by the WhatsApp Business API.
	WhatsAppChallenge string `key:"WHATSAPP_CHALLENGE"`
	// ServerPort is the port on which the server will listen for incoming public requests.
	ServerPort int `key:"SERVER_PORT" required:"true" default:"8080" min:"1024" max:"65535"`
	// HealthPort is the port on which the server will listen for health checks.
	// This can be forced off by setting the key to the empty string.
	HealthPort *int `key:"HEALTH_PORT" default:"8081" min:"1024" max:"65535"`
}

func LoadConfig() (*Config, error) {
	var config Config
	keystore := goconfig.CompositeStore(
		fakeSecretsKeyStore,
		goconfig.EnvironmentKeyStore,
		goconfig.NewEnvFileKeyStore("env.example"))
	if err := goconfig.Load(context.Background(), &config, goconfig.WithKeyStore(keystore)); err != nil {
		return nil, err
	}
	return &config, nil
}
