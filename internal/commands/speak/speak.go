package speak

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jonas747/dca"
	"github.com/stollenaar/aws-rotating-credentials-provider/credentials/filecreds"
	"github.com/stollenaar/copypastabotv2/internal/commands/chat"
	"github.com/stollenaar/copypastabotv2/internal/commands/markov"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/bwmarrin/discordgo"
)

type Queue struct {
	synthed   *polly.SynthesizeSpeechOutput
	sqsObject util.Object
	userID    string
}

type SpeakCommand struct {
	Name        string
	Description string
}

var (
	pollyClient *polly.Client

	queue  []*Queue
	stream *dca.StreamingSession

	guildVC map[string]*voice.Conn

	SpeakCmd = SpeakCommand{
		Name:        "speak",
		Description: "Listen to the beauty of the bot",
	}
)

func init() {
	if os.Getenv("AWS_SHARED_CREDENTIALS_FILE") != "" {
		provider := filecreds.NewFilecredentialsProvider(os.Getenv("AWS_SHARED_CREDENTIALS_FILE"))
		pollyClient = polly.New(polly.Options{
			Credentials: provider,
			Region:      os.Getenv("AWS_REGION"),
		})
	} else {
		// Create a config with the credentials provider.
		cfg, err := config.LoadDefaultConfig(context.TODO())

		if err != nil {
			panic("configuration error, " + err.Error())
		}
		pollyClient = polly.NewFromConfig(cfg)
	}

	guildVC = make(map[string]*voice.Conn)
}

func (s SpeakCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	err := event.DeferCreateMessage(util.ConfigFile.SetEphemeral() == discord.MessageFlagEphemeral)

	if err != nil {
		slog.Error("Error deferring: ", slog.Any("err", err))
		return
	}

	sub := event.SlashCommandInteractionData()

	if len(sub.Options) == 0 {
		_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
			Content: util.Pointer("You must provide at least 1 argument"),
		})
		if err != nil {
			slog.Error("Error editing the response:", slog.Any("err", err))
		}
		return
	}

	speakObject := util.Object{
		Token:         event.Token(),
		Command:       event.Data.CommandName(),
		GuildID:       event.GuildID().String(),
		ApplicationID: event.ApplicationID().String(),
	}

	if user, ok := sub.Options["user"]; ok {
		speakObject.Type = "user"
		speakObject.Data = user.Snowflake().String()
		resp, err := util.ConfigFile.SendStatsBotRequest(speakObject)
		speakObject.Data = resp.Data

		if err != nil {
			fmt.Printf("Encountered an error while processing the speak command: %v\n", err)
			return
		}

	} else if redditPost, ok := sub.Options["redditpost"]; ok {
		speakObject.Data = redditPost.String()
		speakObject.Type = "redditpost"
	} else if url, ok := sub.Options["url"]; ok {
		speakObject.Data = markov.HandleURL(url.String())
		speakObject.Type = "url"
	} else if message, ok := sub.Options["chat"]; ok {
		speakObject.Type = "chat"

		data, err := chat.GetChatGPTResponse("speak", message.String(), event.User().ID.String())
		if err != nil {
			fmt.Println(fmt.Errorf("error interacting with chatgpt char: %e", err))
			_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
				Content: util.Pointer("If you see this, and error likely happened. Whoops"),
			})
			if err != nil {
				slog.Error("Error editing the response:", slog.Any("err", err))
			}
			return
		}
		speakObject.Data = data.Response
	}

	synthData(speakObject, event.User().ID.String(), event.Client())
}

func (s SpeakCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "subreddit",
			Description: "subreddit",
			Required:    true,
		},
	}
}

// Command create a tts experience for the generated markov
func synthData(object util.Object, userID string, bot *bot.Client) {

	if object.Type == "redditpost" {
		post := util.GetRedditPost(object.Data)
		object.Data = post.Post.Body
	}

	contents := util.BreakContent(object.Data, 2950)
	for _, content := range contents {
		resp, err := util.WrapIntoSSML(content, "system")
		textType := types.TextTypeText
		engine := types.EngineNeural
		if err != nil {
			fmt.Println(err)
		} else {
			content = resp.Choices[0].Message.Content
			// if !strings.Contains(content, "<speak>") {
			// 	content = "<speak>" + content + "</speak>"
			// }
			textType = types.TextTypeSsml
			engine = types.EngineStandard
		}
		fmt.Printf("Content to be synthesized: %s\n", content)
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     textType,
			OutputFormat: types.OutputFormatMp3,
			Engine:       engine,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})

		if err != nil {
			fmt.Printf("error synthesizing content: %v", err)
			continue
		}
		queue = append(queue, &Queue{
			synthed:   synthed,
			userID:    userID,
			sqsObject: object,
		})
	}
	if stream != nil {
		if finished, _ := stream.Finished(); !finished {
			fmt.Println("stream is not finished and not null")
			return
		}
	}

	doSpeech(bot)
}

func doSpeech(bot *bot.Client) {
	if len(queue) == 0 {
		fmt.Println("Queue is empty")
		return
	}
	currentSpeech := queue[0]
	queue = queue[1:]

	d := "Playing now"
	response := discordgo.WebhookEdit{
		Content: &d,
	}
	data, _ := json.Marshal(response)
	if _, snerr := strconv.Atoi(currentSpeech.sqsObject.Token); snerr != nil {
		util.SendRequest("PATCH", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
		defer util.SendRequest("DELETE", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
	}

	vs, found := bot.Caches.VoiceState(snowflake.MustParse(currentSpeech.sqsObject.GuildID), snowflake.MustParse(currentSpeech.userID))
	if !found {
		if len(queue) > 0 {
			doSpeech(bot)
		} else {
			guildVC[currentSpeech.sqsObject.GuildID] = nil
		}
		return
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.

	// Send the buffer data.
	stream := currentSpeech.synthed.AudioStream
	conn := bot.VoiceManager.CreateConn(snowflake.MustParse(currentSpeech.sqsObject.GuildID))

	err := conn.Open(context.TODO(), *vs.ChannelID, false, false)
	if len(queue) == 0 {
		defer conn.Close(context.TODO())
	}

	if err != nil {
		fmt.Println(fmt.Errorf("error joining voice channel: %v", err))
	}
	guildVC[currentSpeech.sqsObject.GuildID] = &conn
	err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone)
	if err != nil {
		fmt.Println(fmt.Errorf("error setting speaking to true: %v", err))
	}
	// write the whole body at once
	outFile, err := os.Create("/tmp/tmp.mp3")
	if err != nil {
		fmt.Println(fmt.Errorf("error creating tmp mp3 file: %v", err))
	}
	defer outFile.Close()
	// handle err
	_, err = io.Copy(outFile, stream)
	if err != nil {
		fmt.Println(fmt.Errorf("error copying stream to tmp file: %v", err))
	}
	// Encoding a file and saving it to disk
	encodeSession, err := dca.EncodeFile("/tmp/tmp.mp3", dca.StdEncodeOptions)
	if err != nil {
		fmt.Println(fmt.Errorf("error dca encoding file: %v", err))
	}

	writeOpus(encodeSession, conn.UDP())
	// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
	os.Remove("/tmp/tmp.mp3")

	// Stop speaking
	err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagNone)
	if err != nil {
		fmt.Println(fmt.Errorf("error setting speaking to false: %v", err))
	}

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	if len(queue) > 0 {
		doSpeech(bot)
	} else {
		guildVC[currentSpeech.sqsObject.GuildID] = nil
	}
}


func writeOpus(encodeSession *dca.EncodeSession, w io.Writer) {
	defer 	encodeSession.Cleanup()

	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	var frameLen int16
	// Don't wait for the first tick, run immediately.
	for ; true; <-ticker.C {
		err := binary.Read(encodeSession, binary.LittleEndian, &frameLen)
		if err != nil {
			panic("error reading file: " + err.Error())
		}

		// Copy the frame.
		_, err = io.CopyN(w, encodeSession, int64(frameLen))
		if err != nil && err != io.EOF {
			return
		}
	}
}