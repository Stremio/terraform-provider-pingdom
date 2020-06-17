package pingdom

import (
	"log"
	"os"
	"github.com/Stremio/go-pingdom/pingdom"
)

type Config struct {
	APIToken       string `mapstructure:"pingdom_api_token"`
}

// Client() returns a new client for accessing pingdom.
//
func (c *Config) Client() (*pingdom.Client, error) {

	if v := os.Getenv("PINGDOM_API_TOKEN"); v != "" {
		c.APIToken = v
	}

	client, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
			APIToken: c.APIToken,
		})

	log.Printf("[INFO] Pingdom Client configured for user: %v", err)

	return client, err
}
