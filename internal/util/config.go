package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/joho/godotenv"
)

type Webhook struct {
	id    string
	token string
}

type Config struct {
	DISCORD_TOKEN         string
	DISCORD_WEBHOOK_ID    string
	DISCORD_WEBHOOK_TOKEN string
	DISCORD_CHANNEL_ID    string

	REDDIT_USERNAME      string
	REDDIT_PASSWORD      string
	REDDIT_CLIENT_ID     string
	REDDIT_CLIENT_SECRET string

	AWS_REGION string

	AWS_DISCORD_CHANNEL_ID      string
	AWS_PARAMETER_DISCORD_TOKEN string
	AWS_DISCORD_WEBHOOK_ID      string
	AWS_DISCORD_WEBHOOK_TOKEN   string

	AWS_PARAMETER_PUBLIC_DISCORD_TOKEN string
	AWS_PARAMETER_REDDIT_USERNAME      string
	AWS_PARAMETER_REDDIT_PASSWORD      string
	AWS_PARAMETER_REDDIT_CLIENT_ID     string
	AWS_PARAMETER_REDDIT_CLIENT_SECRET string
	AWS_SNS_TOPIC_ARN                  string
	AWS_SQS_URL                        string
	AWS_SQS_URL_OTHER                  []string

	TERMINAL_REGEX string
	STATISTICS_BOT string
	OPENAI_KEY     string

	DATE_STRING string
}

var (
	ssmClient  *ssm.Client
	ConfigFile *Config

	dnsResolverIP        = "100.100.100.100:53" // Tailscale DNS resolver.
	dnsResolverProto     = "tcp"                // Protocol to use for the DNS resolver
	dnsResolverTimeoutMs = 5000

	dialer = &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}
	dialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}
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
		DISCORD_CHANNEL_ID:                 os.Getenv("DISCORD_CHANNEL_ID"),
		DISCORD_WEBHOOK_ID:                 os.Getenv("DISCORD_WEBHOOK_ID"),
		DISCORD_WEBHOOK_TOKEN:              os.Getenv("DISCORD_WEBHOOK_TOKEN"),
		AWS_REGION:                         os.Getenv("AWS_REGION"),
		AWS_SNS_TOPIC_ARN:                  os.Getenv("AWS_SNS_TOPIC_ARN"),
		AWS_SQS_URL:                        os.Getenv("AWS_SQS_URL"),
		AWS_SQS_URL_OTHER:                  strings.Split(os.Getenv("AWS_SQS_URL_OTHER"), ";"),
		AWS_PARAMETER_DISCORD_TOKEN:        os.Getenv("AWS_PARAMETER_DISCORD_TOKEN"),
		AWS_DISCORD_CHANNEL_ID:             os.Getenv("AWS_DISCORD_CHANNEL_ID"),
		AWS_DISCORD_WEBHOOK_ID:             os.Getenv("AWS_DISCORD_WEBHOOK_ID"),
		AWS_DISCORD_WEBHOOK_TOKEN:          os.Getenv("AWS_DISCORD_WEBHOOK_TOKEN"),
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
		DATE_STRING:                        os.Getenv("DATE_STRING"),
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
		return "", fmt.Errorf("DISCORD_TOKEN or AWS_PARAMETER_DISCORD_TOKEN is not set: %s", string(debug.Stack()))
	}

	if c.DISCORD_TOKEN != "" {
		return c.DISCORD_TOKEN, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_DISCORD_TOKEN)
}

func (c *Config) GetDiscordWebhook() (Webhook, error) {
	if c.DISCORD_WEBHOOK_ID == "" && c.AWS_DISCORD_WEBHOOK_ID == "" {
		return Webhook{}, fmt.Errorf("DISCORD_WEBHOOK_ID or AWS_DISCORD_WEBHOOK_ID is not set: %s", string(debug.Stack()))
	}
	if c.DISCORD_WEBHOOK_TOKEN == "" && c.AWS_DISCORD_WEBHOOK_TOKEN == "" {
		return Webhook{}, fmt.Errorf("DISCORD_WEBHOOK_TOKEN or AWS_DISCORD_WEBHOOK_TOKEN is not set: %s", string(debug.Stack()))
	}

	webhook := Webhook{}
	if c.DISCORD_WEBHOOK_ID != "" {
		webhook.id = c.DISCORD_WEBHOOK_ID
	} else {
		par, err := getAWSParameter(c.AWS_DISCORD_WEBHOOK_ID)
		if err != nil {
			return Webhook{}, err
		}
		webhook.id = par
	}
	if c.DISCORD_WEBHOOK_TOKEN != "" {
		webhook.token = c.DISCORD_WEBHOOK_TOKEN
	} else {
		par, err := getAWSParameter(c.AWS_DISCORD_WEBHOOK_TOKEN)
		if err != nil {
			return Webhook{}, err
		}
		webhook.token = par
	}
	return webhook, nil
}

func (c *Config) GetPublicDiscordToken() (string, error) {
	if c.AWS_PARAMETER_PUBLIC_DISCORD_TOKEN == "" {
		return "", fmt.Errorf("AWS_PARAMETER_PUBLIC_DISCORD_TOKEN is not set: %s", string(debug.Stack()))
	}

	return getAWSParameter(c.AWS_PARAMETER_PUBLIC_DISCORD_TOKEN)
}

func (c *Config) GetRedditUsername() (string, error) {
	if c.AWS_PARAMETER_REDDIT_USERNAME == "" && c.REDDIT_USERNAME == "" {
		return "", fmt.Errorf("REDDIT_USERNAME or AWS_PARAMETER_REDDIT_USERNAME is not set: %s", string(debug.Stack()))
	}

	if c.REDDIT_USERNAME != "" {
		return c.REDDIT_USERNAME, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_USERNAME)
}

func (c *Config) GetRedditPassword() (string, error) {
	if c.AWS_PARAMETER_REDDIT_PASSWORD == "" && c.REDDIT_PASSWORD == "" {
		return "", fmt.Errorf("AWS_PARAMETER_REDDIT_PASSWORD or REDDIT_PASSWORD is not set: %s", string(debug.Stack()))
	}

	if c.REDDIT_PASSWORD != "" {
		return c.REDDIT_PASSWORD, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_PASSWORD)
}

func (c *Config) GetRedditClientID() (string, error) {
	if c.AWS_PARAMETER_REDDIT_CLIENT_ID == "" && c.REDDIT_CLIENT_ID == "" {
		return "", fmt.Errorf("REDDIT_CLIENT_ID or AWS_PARAMETER_REDDIT_CLIENT_ID is not set: %s", string(debug.Stack()))
	}

	if c.REDDIT_CLIENT_ID != "" {
		return c.REDDIT_CLIENT_ID, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_CLIENT_ID)
}

func (c *Config) GetRedditClientSecret() (string, error) {
	if c.AWS_PARAMETER_REDDIT_CLIENT_SECRET == "" && c.REDDIT_CLIENT_SECRET == "" {
		return "", fmt.Errorf("REDDIT_CLIENT_SECRET or REDDIT_CLIENT_SECRET is not set: %s", string(debug.Stack()))
	}

	if c.REDDIT_CLIENT_SECRET != "" {
		return c.REDDIT_CLIENT_SECRET, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_REDDIT_CLIENT_SECRET)
}

func (c *Config) GetDiscordChannelID() (string, error) {
	if c.AWS_DISCORD_CHANNEL_ID == "" && c.DISCORD_CHANNEL_ID == "" {
		return "", fmt.Errorf("AWS_DISCORD_CHANNEL_ID or DISCORD_CHANNEL_ID is not set: %s", string(debug.Stack()))
	}

	if c.DISCORD_CHANNEL_ID != "" {
		return c.DISCORD_CHANNEL_ID, nil
	}

	return getAWSParameter(c.AWS_DISCORD_CHANNEL_ID)
}

func getAWSParameter(parameterName string) (string, error) {
	out, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		fmt.Println(fmt.Errorf("error from fetching parameter %s. With error: %w", parameterName, err))
		return "", err
	}
	return *out.Parameter.Value, err
}

func (c *Config) GetOpenAIKey() (string, error) {
	if c.OPENAI_KEY == "" {
		return "", fmt.Errorf("OPENAI_KEY is not defined")
	}
	return getAWSParameter(c.OPENAI_KEY)
}

func (c *Config) SendStatsBotRequest(sqsObject Object) (Object, error) {
	jsonData, err := json.Marshal(sqsObject)
	if err != nil {
		return Object{}, err
	}
	bodyReader := bytes.NewReader(jsonData)

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/userMessages", c.STATISTICS_BOT), bodyReader)

	if err != nil {
		return Object{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	fmt.Printf("Doing Request: %v", *req)

	res, err := client.Do(req)
	if err != nil {
		return Object{}, err
	}
	body, _ := io.ReadAll(res.Body)
	var object Object
	json.Unmarshal(body, &object)
	fmt.Printf("Response body: %s\n", string(body))
	return object, err
}
