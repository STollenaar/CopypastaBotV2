package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/disgoorg/disgo/discord"
	"github.com/joho/godotenv"
	"github.com/stollenaar/aws-rotating-credentials-provider/credentials/filecreds"
)

type Webhook struct {
	id    string
	token string
}

type Config struct {
	DEBUG bool

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
	AWS_PARAMETER_ANTHROPIC_KEY        string

	TERMINAL_REGEX    string
	STATISTICS_BOT    string
	ANTHROPIC_API_KEY string

	LITELLM_URL         string
	LITELLM_MODEL       string
	LITELLM_API_KEY     string
	AWS_LITELLM_API_KEY string

	DUCKDB_PATH string
	HEALTH_PORT string

	ADMIN_USER_ID string
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

	ConfigFile = &Config{
		DISCORD_TOKEN:                      os.Getenv("DISCORD_TOKEN"),
		DISCORD_CHANNEL_ID:                 os.Getenv("DISCORD_CHANNEL_ID"),
		DISCORD_WEBHOOK_ID:                 os.Getenv("DISCORD_WEBHOOK_ID"),
		DISCORD_WEBHOOK_TOKEN:              os.Getenv("DISCORD_WEBHOOK_TOKEN"),
		AWS_REGION:                         os.Getenv("AWS_REGION"),
		AWS_PARAMETER_DISCORD_TOKEN:        os.Getenv("AWS_PARAMETER_DISCORD_TOKEN"),
		AWS_DISCORD_CHANNEL_ID:             os.Getenv("AWS_DISCORD_CHANNEL_ID"),
		AWS_DISCORD_WEBHOOK_ID:             os.Getenv("AWS_DISCORD_WEBHOOK_ID"),
		AWS_DISCORD_WEBHOOK_TOKEN:          os.Getenv("AWS_DISCORD_WEBHOOK_TOKEN"),
		AWS_PARAMETER_PUBLIC_DISCORD_TOKEN: os.Getenv("AWS_PARAMETER_PUBLIC_DISCORD_TOKEN"),
		AWS_PARAMETER_REDDIT_USERNAME:      os.Getenv("AWS_PARAMETER_REDDIT_USERNAME"),
		AWS_PARAMETER_REDDIT_PASSWORD:      os.Getenv("AWS_PARAMETER_REDDIT_PASSWORD"),
		AWS_PARAMETER_REDDIT_CLIENT_ID:     os.Getenv("AWS_PARAMETER_REDDIT_CLIENT_ID"),
		AWS_PARAMETER_REDDIT_CLIENT_SECRET: os.Getenv("AWS_PARAMETER_REDDIT_CLIENT_SECRET"),
		AWS_PARAMETER_ANTHROPIC_KEY:        os.Getenv("AWS_PARAMETER_ANTHROPIC_KEY"),
		REDDIT_USERNAME:                    os.Getenv("REDDIT_USERNAME"),
		REDDIT_PASSWORD:                    os.Getenv("REDDIT_PASSWORD"),
		REDDIT_CLIENT_ID:                   os.Getenv("REDDIT_CLIENT_ID"),
		REDDIT_CLIENT_SECRET:               os.Getenv("REDDIT_CLIENT_SECRET"),
		TERMINAL_REGEX:                     os.Getenv("TERMINAL_REGEX"),
		STATISTICS_BOT:                     os.Getenv("STATSBOT_URL"),
		ANTHROPIC_API_KEY:                  os.Getenv("ANTHROPIC_API_KEY"),
		LITELLM_URL:                        os.Getenv("LITELLM_URL"),
		LITELLM_MODEL:                      os.Getenv("LITELLM_MODEL"),
		LITELLM_API_KEY:                    os.Getenv("LITELLM_API_KEY"),
		AWS_LITELLM_API_KEY:                os.Getenv("AWS_LITELLM_API_KEY"),
		DUCKDB_PATH:                        os.Getenv("DUCKDB_PATH"),
		HEALTH_PORT:                        os.Getenv("HEALTH_PORT"),
		ADMIN_USER_ID:                      os.Getenv("ADMIN_USER_ID"),
	}

	if ConfigFile.TERMINAL_REGEX == "" {
		ConfigFile.TERMINAL_REGEX = `(\.|,|:|;|\?|!)$`
	}
	if ConfigFile.HEALTH_PORT == "" {
		ConfigFile.HEALTH_PORT = "8080"
	}

}

func init() {

	if os.Getenv("AWS_SHARED_CREDENTIALS_FILE") != "" {
		provider := filecreds.NewFilecredentialsProvider(os.Getenv("AWS_SHARED_CREDENTIALS_FILE"))
		ssmClient = ssm.New(ssm.Options{
			Credentials: provider,
			Region:      ConfigFile.AWS_REGION,
		})
	} else {

		// Create a config with the credentials provider.
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(ConfigFile.AWS_REGION),
		)

		if err != nil {
			if _, isProfileNotExistError := err.(config.SharedConfigProfileNotExistError); isProfileNotExistError {
				cfg, err = config.LoadDefaultConfig(context.TODO(),
					config.WithRegion(ConfigFile.AWS_REGION),
				)
			}
			if err != nil {
				log.Fatal("Error loading AWS config:", err)
			}
		}

		ssmClient = ssm.NewFromConfig(cfg)
	}
}

func (c *Config) GetDiscordToken() string {
	if ConfigFile.DISCORD_TOKEN == "" && ConfigFile.AWS_PARAMETER_DISCORD_TOKEN == "" {
		log.Fatal("DISCORD_TOKEN or AWS_PARAMETER_NAME is not set")
	}

	if ConfigFile.DISCORD_TOKEN != "" {
		return ConfigFile.DISCORD_TOKEN
	}
	out, err := getAWSParameter(ConfigFile.AWS_PARAMETER_DISCORD_TOKEN)
	if err != nil {
		log.Fatal(err)
	}
	return out
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
		slog.Error("Error fetching AWS parameter", slog.String("parameter", parameterName), slog.Any("err", err))
		return "", err
	}
	return *out.Parameter.Value, err
}

func (c *Config) GetAnthropicKey() (string, error) {
	if c.ANTHROPIC_API_KEY == "" && c.AWS_PARAMETER_ANTHROPIC_KEY == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not defined")
	}
	if c.ANTHROPIC_API_KEY != "" {
		return c.ANTHROPIC_API_KEY, nil
	}
	return getAWSParameter(c.AWS_PARAMETER_ANTHROPIC_KEY)
}

func (c *Config) SendStatsBotRequest(sqsObject Object) (UserMessagesResponse, error) {
	jsonData, err := json.Marshal(sqsObject)
	if err != nil {
		return UserMessagesResponse{}, err
	}
	bodyReader := bytes.NewReader(jsonData)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/userMessages", c.STATISTICS_BOT), bodyReader)

	if err != nil {
		return UserMessagesResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	slog.Debug("Doing request", slog.Any("request", *req))

	res, err := client.Do(req)
	if err != nil {
		return UserMessagesResponse{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return UserMessagesResponse{}, fmt.Errorf("error reading statistics bot response: %w", err)
	}

	slog.Debug("Response body", slog.Int("status", res.StatusCode), slog.String("body", string(body)))

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		snippet := string(body)
		if len(snippet) > 512 {
			snippet = snippet[:512] + "..."
		}
		return UserMessagesResponse{}, fmt.Errorf("statistics bot returned status %s: %s", res.Status, snippet)
	}

	var response UserMessagesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		slog.Error("Error unmarshaling", slog.Any("err", err))
		return UserMessagesResponse{}, fmt.Errorf("error unmarshaling statistics bot response: %w", err)
	}

	return response, nil
}

func GetLiteLLMAPIKey() (string, error) {
	if ConfigFile.LITELLM_API_KEY != "" {
		return ConfigFile.LITELLM_API_KEY, nil
	}
	if ConfigFile.AWS_LITELLM_API_KEY == "" {
		return "", errors.New("LITELLM_API_KEY or AWS_LITELLM_API_KEY is not set")
	}
	return getAWSParameter(ConfigFile.AWS_LITELLM_API_KEY)
}

func (c *Config) SetEphemeral() discord.MessageFlags {
	if c.DEBUG {
		return discord.MessageFlagEphemeral
	} else {
		return discord.MessageFlagsNone
	}
}

func (c *Config) SetComponentV2Flags() *discord.MessageFlags {
	eph := c.SetEphemeral()
	eph = eph.Add(discord.MessageFlagIsComponentsV2)
	return &eph
}
