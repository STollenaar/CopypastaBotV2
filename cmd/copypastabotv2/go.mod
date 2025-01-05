module github.com/stollenaar/copypastabotv2

go 1.23.4

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/stollenaar/copypastabotv2/internal/commands/browse v0.0.0-20241231222915-4a1ec13e8501
	github.com/stollenaar/copypastabotv2/internal/commands/chat v0.0.0-20241231222915-4a1ec13e8501
	github.com/stollenaar/copypastabotv2/internal/commands/help v0.0.0-20241231222915-4a1ec13e8501
	github.com/stollenaar/copypastabotv2/internal/commands/markov v0.0.0-20241231222915-4a1ec13e8501
	github.com/stollenaar/copypastabotv2/internal/commands/pasta v0.0.0-20241231222915-4a1ec13e8501
	github.com/stollenaar/copypastabotv2/internal/commands/speak v0.0.0-20241231222915-4a1ec13e8501
	github.com/stollenaar/copypastabotv2/internal/util v0.0.0-20241231222915-4a1ec13e8501
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.17.48 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/polly v1.45.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.33.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.56.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.3 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/jonas747/dca v0.0.0-20210930103944-155f5e5f0cc7 // indirect
	github.com/jonas747/ogg v0.0.0-20161220051205-b4f6f4cf3757 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/stollenaar/aws-rotating-credentials-provider/credentials v0.0.0-20240112205114-26346908241a // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20250103183323-7d7fa50e5329 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/PuerkitoBio/goquery v1.10.1 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2 v1.32.7 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.28.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.26 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/sashabaranov/go-openai v1.36.1 // indirect
	github.com/stollenaar/copypastabotv2/pkg/markov v0.0.0-20241231222915-4a1ec13e8501 // indirect
	github.com/vartanbeno/go-reddit/v2 v2.0.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect; indirect`
	golang.org/x/oauth2 v0.25.0 // indirect
)

replace (
	github.com/stollenaar/copypastabotv2/internal/commands/browse => ../../internal/commands/browse
	github.com/stollenaar/copypastabotv2/internal/commands/chat => ../../internal/commands/chat
	github.com/stollenaar/copypastabotv2/internal/commands/help => ../../internal/commands/help
	github.com/stollenaar/copypastabotv2/internal/commands/markov => ../../internal/commands/markov
	github.com/stollenaar/copypastabotv2/internal/commands/pasta => ../../internal/commands/pasta
	github.com/stollenaar/copypastabotv2/internal/commands/speak => ../../internal/commands/speak
	github.com/stollenaar/copypastabotv2/internal/util => ../../internal/util
)
