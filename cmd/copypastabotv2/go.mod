module github.com/stollenaar/copypastabotv2

go 1.23.4

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/stollenaar/copypastabotv2/internal/commands/browse v0.0.0-00010101000000-000000000000
	github.com/stollenaar/copypastabotv2/internal/commands/chat v0.0.0-00010101000000-000000000000
	github.com/stollenaar/copypastabotv2/internal/commands/help v0.0.0-00010101000000-000000000000
	github.com/stollenaar/copypastabotv2/internal/commands/markov v0.0.0-00010101000000-000000000000
	github.com/stollenaar/copypastabotv2/internal/commands/pasta v0.0.0-00010101000000-000000000000
	github.com/stollenaar/copypastabotv2/internal/commands/speak v0.0.0-00010101000000-000000000000
	github.com/stollenaar/copypastabotv2/internal/util v0.0.0-20241229024158-7c156679fb17
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.17.28 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.12 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/polly v1.45.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.31.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.52.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.4 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/jonas747/dca v0.0.0-20210930103944-155f5e5f0cc7 // indirect
	github.com/jonas747/ogg v0.0.0-20161220051205-b4f6f4cf3757 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
)

require (
	github.com/PuerkitoBio/goquery v1.9.2 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.32.7 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.26 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/sashabaranov/go-openai v1.31.0 // indirect
	github.com/stollenaar/copypastabotv2/pkg/markov v0.0.0-20241003212403-edcf95f330b2 // indirect
	github.com/vartanbeno/go-reddit/v2 v2.0.1 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.28.0 // indirect; indirect`
	golang.org/x/oauth2 v0.22.0 // indirect
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
