package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
	AWS_PARAMETER_PUBLIC_DISCORD_TOKEN string
	AWS_PARAMETER_REDDIT_USERNAME      string
	AWS_PARAMETER_REDDIT_PASSWORD      string
	AWS_PARAMETER_REDDIT_CLIENT_ID     string
	AWS_PARAMETER_REDDIT_CLIENT_SECRET string
	AWS_SQS_URL                        string
	AWS_SQS_URL_OTHER                  []string

	TERMINAL_REGEX string
	STATISTICS_BOT string
	OPENAI_KEY     string
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
		AWS_SQS_URL:                        os.Getenv("AWS_SQS_URL"),
		AWS_SQS_URL_OTHER:                  strings.Split(os.Getenv("AWS_SQS_URL_OTHER"), ";"),
		AWS_PARAMETER_DISCORD_TOKEN:        os.Getenv("AWS_PARAMETER_DISCORD_TOKEN"),
		AWS_PARAMETER_PUBLIC_DISCORD_TOKEN: os.Getenv("AWS_PARAMETER_PUBLIC_DISCORD_TOKEN"),
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
		OPENAI_KEY:                         os.Getenv("OPENAI_KEY"),
	}

	if ConfigFile.TERMINAL_REGEX == "" {
		ConfigFile.TERMINAL_REGEX = `(\.|,|:|;|\?|!)$`
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

func (c *Config) GetDiscordToken() (string, error) {
	if c.DISCORD_TOKEN == "" && c.AWS_PARAMETER_DISCORD_TOKEN == "" {
		log.Fatal("DISCORD_TOKEN or AWS_PARAMETER_DISCORD_TOKEN is not set.")
	}

	if c.DISCORD_TOKEN != "" {
		return c.DISCORD_TOKEN, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_DISCORD_TOKEN)
}

func (c *Config) GetPublicDiscordToken() (string, error) {
	if c.AWS_PARAMETER_PUBLIC_DISCORD_TOKEN == "" {
		log.Fatal("AWS_PARAMETER_PUBLIC_DISCORD_TOKEN is not set.")
	}

	return getAWSParameter(c.AWS_PARAMETER_PUBLIC_DISCORD_TOKEN)
}

func (c *Config) GetRedditUsername() (string, error) {
	if c.AWS_PARAMETER_REDDIT_USERNAME == "" && c.REDDIT_USERNAME == "" {
		log.Fatal("REDDIT_USERNAME or AWS_PARAMETER_REDDIT_USERNAME is not set.")
	}

	if c.REDDIT_USERNAME != "" {
		return c.REDDIT_USERNAME, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_USERNAME)
}

func (c *Config) GetRedditPassword() (string, error) {
	if c.AWS_PARAMETER_REDDIT_PASSWORD == "" && c.REDDIT_PASSWORD == "" {
		log.Fatal("AWS_PARAMETER_REDDIT_PASSWORD or REDDIT_PASSWORD is not set.")
	}

	if c.REDDIT_PASSWORD != "" {
		return c.REDDIT_PASSWORD, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_PASSWORD)
}

func (c *Config) GetRedditClientID() (string, error) {
	if c.AWS_PARAMETER_REDDIT_CLIENT_ID == "" && c.REDDIT_CLIENT_ID == "" {
		log.Fatal("REDDIT_CLIENT_ID or AWS_PARAMETER_REDDIT_CLIENT_ID is not set.")
	}

	if c.REDDIT_CLIENT_ID != "" {
		return c.REDDIT_CLIENT_ID, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_CLIENT_ID)
}

func (c *Config) GetRedditClientSecret() (string, error) {
	if c.AWS_PARAMETER_REDDIT_CLIENT_SECRET == "" && c.REDDIT_CLIENT_SECRET == "" {
		log.Fatal("REDDIT_CLIENT_SECRET or REDDIT_CLIENT_SECRET is not set.")
	}

	if c.REDDIT_CLIENT_SECRET != "" {
		return c.REDDIT_CLIENT_SECRET, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_CLIENT_SECRET)
}

func getAWSParameter(parameterName string) (string, error) {
	out, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		fmt.Println(fmt.Errorf("error from fetching parameter %s. With error: %w", parameterName, err))
	}
	return *out.Parameter.Value, err
}

func (c *Config) GetOpenAIKey() (string, error) {
	if c.OPENAI_KEY == "" {
		log.Fatal("OPENAI_KEY is not defined")
	}
	return getAWSParameter(c.OPENAI_KEY)
}
