package commands

import (
	"context"
	"copypastabot/util"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/nint8835/parsley"
)

var pollyClient *polly.Client

var queue []*Queue
var stream *dca.StreamingSession

type Queue struct {
	synthed *polly.SynthesizeSpeechOutput
	message *discordgo.MessageCreate
}

func SpeakInit(parser *parsley.Parser) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	pollyClient = polly.NewFromConfig(cfg)

	parser.NewCommand("speak", "Speak the generated markov", SpeakCommand)
}

// SpeakCommand create a tts experience for the generated markov
func SpeakCommand(message *discordgo.MessageCreate, args commandArgs) {
	_, err := Bot.findUserVoiceState(message.Author.ID)

	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("You must be in a VoiceChannel in order to do this!"))
	}

	var markov string
	if _, err := url.ParseRequestURI(args.Word); err == nil {
		markov, err = MarkovURLCommand(message, args, false)
	} else if strings.Contains(args.Word, "@") {
		markov, err = MarkovUserCommand(message, args, false)
	} else if args.Word != "" {
		postCommnents := util.GetRedditPost(args.Word)
		markov = postCommnents.Post.Body

		if postCommnents.Post.Body == "" {
			markov = postCommnents.Post.URL
		}

	} else {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Not a argument was provided"))
		return
	}
	if err != nil {
		Bot.ChannelMessageSend(message.ChannelID, fmt.Sprintln("Something went wrong"))
		return
	}
	contents := util.BreakContent(markov, 2950)
	for _, content := range contents {
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     types.TextTypeText,
			OutputFormat: types.OutputFormatOggVorbis,
			Engine:       types.EngineNeural,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})

		if err == nil {
			queue = append(queue, &Queue{
				synthed: synthed,
				message: message,
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
	currentSpeech := queue[0]
	queue = queue[1:]

	vs, err := Bot.findUserVoiceState(currentSpeech.message.Author.ID)
	if err != nil {
		if len(queue) > 0 {
			doSpeech()
		}
		return
	}

	vc, err := Bot.ChannelVoiceJoin(currentSpeech.message.GuildID, vs.ChannelID, false, true)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	fmt.Println(currentSpeech.synthed.ResultMetadata)
	stream := currentSpeech.synthed.AudioStream
	InBuf := make([]byte, 4096)
	binary.Read(stream, binary.LittleEndian, &InBuf)
	// Stop speaking
	vc.Speaking(false)
	if len(queue) > 0 {
		doSpeech()
	}
}

func (s DiscordBot) findUserVoiceState(userid string) (*discordgo.VoiceState, error) {
	for _, guild := range s.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errors.New("Could not find user's voice state")
}

// Reads an opus packet to send over the vc.OpusSend channel
func readOpus(source io.Reader) ([]byte, error) {
	var opuslen int16
	err := binary.Read(source, binary.LittleEndian, &opuslen)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, err
		}
		return nil, errors.New("ERR reading opus header")
	}

	var opusframe = make([]byte, opuslen)
	err = binary.Read(source, binary.LittleEndian, &opusframe)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, err
		}
		return nil, errors.New("ERR reading opus frame")
	}

	return opusframe, nil
}
