package util

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/joho/godotenv"
)

type Config struct {
	DISCORD_TOKEN        string
	REDDIT_USERNAME      string
	REDDIT_PASSWORD      string
	REDDIT_CLIENT_ID     string
	REDDIT_CLIENT_SECRET string

	AWS_REGION string

	AWS_PARAMETER_DISCORD_TOKEN        string
	AWS_PARAMETER_REDDIT_USERNAME      string
	AWS_PARAMETER_REDDIT_PASSWORD      string
	AWS_PARAMETER_REDDIT_CLIENT_ID     string
	AWS_PARAMETER_REDDIT_CLIENT_SECRET string

	TERMINAL_REGEX string
	STATISTICS_BOT string
}

var (
	ssmClient  *ssm.Client
	ConfigFile *Config
)

// initializing the main config file
func init() {
	ConfigFile = new(Config)
	_, err := os.Stat(".env")
	if err == nil {
		err = godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading environment variables")
		}
	}

	err = nil

	ConfigFile = &Config{
		DISCORD_TOKEN:                      os.Getenv("DISCORD_TOKEN"),
		AWS_REGION:                         os.Getenv("AWS_REGION"),
		AWS_PARAMETER_DISCORD_TOKEN:        os.Getenv("AWS_PARAMETER_DISCORD_TOKEN"),
		AWS_PARAMETER_REDDIT_USERNAME:      os.Getenv("AWS_PARAMETER_REDDIT_USERNAME"),
		AWS_PARAMETER_REDDIT_PASSWORD:      os.Getenv("AWS_PARAMETER_REDDIT_PASSWORD"),
		AWS_PARAMETER_REDDIT_CLIENT_ID:     os.Getenv("AWS_PARAMETER_REDDIT_CLIENT_ID"),
		AWS_PARAMETER_REDDIT_CLIENT_SECRET: os.Getenv("AWS_PARAMETER_REDDIT_CLIENT_SECRET"),
		REDDIT_USERNAME:                    os.Getenv("REDDIT_USERNAME"),
		REDDIT_PASSWORD:                    os.Getenv("REDDIT_PASSWORD"),
		REDDIT_CLIENT_ID:                   os.Getenv("REDDIT_CLIENT_ID"),
		REDDIT_CLIENT_SECRET:               os.Getenv("REDDIT_CLIENT_SECRET"),
		TERMINAL_REGEX:                     os.Getenv("TERMINAL_REGEX"),
		STATISTICS_BOT:                     os.Getenv("STATSBOT_URL"),
	}

	if ConfigFile.TERMINAL_REGEX == "" {
		ConfigFile.TERMINAL_REGEX = `(\.|,|:|;|\?|!)$`
	}

	if ConfigFile.DISCORD_TOKEN == "" && ConfigFile.AWS_PARAMETER_DISCORD_TOKEN == "" {
		err = fmt.Errorf("DISCORD_TOKEN or AWS_PARAMETER_DISCORD_TOKEN is not set. %w", err)
	}
	if ConfigFile.AWS_PARAMETER_REDDIT_CLIENT_ID == "" && ConfigFile.REDDIT_CLIENT_ID == "" {
		err = fmt.Errorf("REDDIT_CLIENT_ID or AWS_PARAMETER_REDDIT_CLIENT_ID if not set. %w", err)
	}
	if ConfigFile.AWS_PARAMETER_REDDIT_USERNAME == "" && ConfigFile.REDDIT_USERNAME == "" {
		err = fmt.Errorf("REDDIT_USERNAME or AWS_PARAMETER_REDDIT_USERNAME if not set. %w", err)
	}
	if ConfigFile.AWS_PARAMETER_REDDIT_CLIENT_SECRET == "" && ConfigFile.REDDIT_CLIENT_SECRET == "" {
		err = fmt.Errorf("REDDIT_CLIENT_SECRET or REDDIT_CLIENT_SECRET if not set. %w", err)
	}
	if ConfigFile.AWS_PARAMETER_REDDIT_PASSWORD == "" && ConfigFile.REDDIT_PASSWORD == "" {
		err = fmt.Errorf("AWS_PARAMETER_REDDIT_PASSWORD or REDDIT_PASSWORD if not set. %w", err)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func init() {

	// Create a config with the credentials provider.
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal("Error loading AWS config:", err)
	}

	ssmClient = ssm.NewFromConfig(cfg)
}

func (c *Config) GetDiscordToken() string {
	if c.DISCORD_TOKEN != "" {
		return c.DISCORD_TOKEN
	}
	return getAWSParameter(c.AWS_PARAMETER_DISCORD_TOKEN)
}

func (c *Config) GetRedditUsername() string {
	if c.REDDIT_USERNAME != "" {
		return c.REDDIT_USERNAME
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_USERNAME)
}

func (c *Config) GetRedditPassword() string {
	if c.REDDIT_PASSWORD != "" {
		return c.REDDIT_PASSWORD
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_PASSWORD)
}

func (c *Config) GetRedditClientID() string {
	if c.REDDIT_CLIENT_ID != "" {
		return c.REDDIT_CLIENT_ID
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_CLIENT_ID)
}

func (c *Config) GetRedditClientSecret() string {
	if c.REDDIT_CLIENT_SECRET != "" {
		return c.REDDIT_CLIENT_SECRET
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_CLIENT_SECRET)
}

func getAWSParameter(parameterName string) string {
	out, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: true,
	})
	if err != nil {
		fmt.Println(fmt.Errorf("error from fetching parameter %s. With error: %w", parameterName, err))
	}
	return *out.Parameter.Value
}
