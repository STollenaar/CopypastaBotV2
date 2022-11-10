package speakCommand

import (
	"context"
	"copypastabot/lib/markovCommand"
	"copypastabot/util"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

var (
	pollyClient *polly.Client

	queue  []*Queue
	stream *dca.StreamingSession
	Bot    *discordgo.Session
)

type Queue struct {
	synthed *polly.SynthesizeSpeechOutput
	guildID string
	userID  string
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("personal"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	pollyClient = polly.NewFromConfig(cfg)

}

// Command create a tts experience for the generated markov
func Command(bot *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if Bot == nil {
		Bot = bot
	}

	userID := interaction.Member.User.ID
	_, err := findUserVoiceState(userID)

	if err != nil {
		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You must be in a VoiceChannel in order to do this!",
			},
		})
		return
	} else {
		resp := "I am preparing my beautifull voice"

		Bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: resp,
			},
		})
	}

	var markov string

	parsedArguments := util.ParseArguments([]string{"url", "user", "redditpost"}, interaction)

	if url, ok := parsedArguments["Url"]; ok {
		markov, err = markovCommand.GetMarkovURL(url)
	} else if user, ok := parsedArguments["User"]; ok {
		markov, err = markovCommand.GetUserMarkov(interaction.GuildID, user)
	} else if post, ok := parsedArguments["Redditpost"]; ok {
		postCommnents := util.GetRedditPost(post)
		markov = postCommnents.Post.Body

		if postCommnents.Post.Body == "" {
			markov = postCommnents.Post.URL
		}

	} else {
		NO_VALID_ARGUMENT := "No valid argument was provided"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &NO_VALID_ARGUMENT,
		})
		return
	}
	if err != nil {
		ERROR_CALLED := "Something went wrong trying to create the Markov!!"
		Bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &ERROR_CALLED,
		})
		return
	}

	contents := util.BreakContent(markov, 2950)
	for _, content := range contents {
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     types.TextTypeText,
			OutputFormat: types.OutputFormatMp3,
			Engine:       types.EngineNeural,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})

		if err == nil {
			queue = append(queue, &Queue{
				synthed: synthed,
				userID:  userID,
				guildID: interaction.GuildID,
			})
		}
	}
	if stream != nil {
		if finished, _ := stream.Finished(); !finished {
			return
		}
	}

	doSpeech()
}

func doSpeech() {
	if len(queue) == 0 {
		return
	}
	currentSpeech := queue[0]
	queue = queue[1:]

	vs, err := findUserVoiceState(currentSpeech.userID)
	if err != nil {
		if len(queue) > 0 {
			doSpeech()
		}
		return
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.

	// Send the buffer data.
	stream := currentSpeech.synthed.AudioStream

	vc, _ := Bot.ChannelVoiceJoin(currentSpeech.guildID, vs.ChannelID, false, false)
	vc.Speaking(true)

	// write the whole body at once
	outFile, err := os.Create("tmp.mp3")
	if err != nil {
		fmt.Println(err)
	}
	// handle err
	_, err = io.Copy(outFile, stream)
	if err != nil {
		fmt.Println(err)
	}
	outFile.Close()
	// Encoding a file and saving it to disk
	encodeSession, err := dca.EncodeFile("tmp.mp3", dca.StdEncodeOptions)
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		// Handle the error
		fmt.Println(err)
	}

	done := make(chan error)
	dca.NewStream(encodeSession, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		fmt.Println(err)
	}

	// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
	encodeSession.Cleanup()
	os.Remove("tmp.mp3")

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	if len(queue) > 0 {
		doSpeech()
	}
}

func findUserVoiceState(userid string) (*discordgo.VoiceState, error) {
	for _, guild := range Bot.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("could not find user's voice state")
}