package speak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jonas747/dca"
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

var (
	pollyClient *polly.Client

	queue  []*Queue
	stream *dca.StreamingSession

	guildVC map[string]*discordgo.VoiceConnection
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	pollyClient = polly.NewFromConfig(cfg)
	guildVC = make(map[string]*discordgo.VoiceConnection)
}

func Handler(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	})

	var response discordgo.WebhookEdit

	parsedArguments := util.ParseArguments([]string{"redditpost", "url", "user", "chat"}, interaction.ApplicationCommandData().Options)
	if parsedArguments["Redditpost"] == "" && parsedArguments["Url"] == "" && parsedArguments["User"] == "" && parsedArguments["Chat"] == "" {
		response.Content = aws.String("You must provide at least 1 argument")
		bot.InteractionResponseEdit(interaction.Interaction, &response)
		return
	}

	speakObject := util.Object{
		Token:         interaction.Token,
		Command:       interaction.ApplicationCommandData().Name,
		GuildID:       interaction.GuildID,
		ApplicationID: interaction.AppID,
	}

	if parsedArguments["User"] == "" {

		if parsedArguments["Redditpost"] != "" {
			speakObject.Data = parsedArguments["Redditpost"]
			speakObject.Type = "redditpost"
		} else if parsedArguments["Url"] != "" {
			speakObject.Type = "url"
			speakObject.Data = markov.HandleURL(parsedArguments["Url"])
		} else if parsedArguments["Chat"] != "" {
			speakObject.Type = "chat"

			data, err := chat.GetChatGPTResponse("speak", parsedArguments["Chat"], interaction.Member.User.ID)
			if err != nil {
				fmt.Println(fmt.Errorf("error interacting with chatgpt char: %e", err))
				response.Content = aws.String("If you see this, and error likely happened. Whoops")
				bot.InteractionResponseEdit(interaction.Interaction, &response)
				return
			}
			speakObject.Data = data.Choices[0].Message.Content
		}

	} else {
		speakObject.Type = "user"
		speakObject.Data = parsedArguments["User"]
		resp, err := util.ConfigFile.SendStatsBotRequest(speakObject)
		speakObject.Data = resp.Data

		if err != nil {
			fmt.Printf("Encountered an error while processing the speak command: %v\n", err)
			return
		}
	}
	synthData(speakObject, interaction.Member.User.ID, bot)
}

// Command create a tts experience for the generated markov
func synthData(object util.Object, userID string, bot *discordgo.Session) {

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

func doSpeech(bot *discordgo.Session) {
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
	vs, err := bot.State.VoiceState(currentSpeech.sqsObject.GuildID, currentSpeech.userID)
	if err != nil {
		fmt.Println(fmt.Errorf("error finding voice state: %v", err))
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

	vc, err := bot.ChannelVoiceJoin(currentSpeech.sqsObject.GuildID, vs.ChannelID, false, false)
	if len(queue) == 0 {
		defer vc.Disconnect()
	}

	if err != nil {
		fmt.Println(fmt.Errorf("error joining voice channel: %v", err))
	}
	guildVC[currentSpeech.sqsObject.GuildID] = vc
	err = vc.Speaking(true)
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

	done := make(chan error)
	dca.NewStream(encodeSession, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		fmt.Println(fmt.Errorf("stream error: %v", err))
	}

	// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
	encodeSession.Cleanup()
	os.Remove("/tmp/tmp.mp3")

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	if len(queue) > 0 {
		doSpeech(bot)
	} else {
		guildVC[currentSpeech.sqsObject.GuildID] = nil
	}
	return
}
