package speak

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
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
	"github.com/stollenaar/copypastabotv2/internal/database"
	"github.com/stollenaar/copypastabotv2/internal/util"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
)

type Queue struct {
	chunkID     int64 // DB row in speak_queue_chunks holding the MP3 bytes
	sqsObject   util.Object
	userID      string
	dbID        int64
	isLastChunk bool
}

type SpeakCommand struct {
	Name        string
	Description string
}

var (
	pollyClient *polly.Client

	mu          sync.Mutex
	queues      map[string][]*Queue      // per-guild queue
	guildActive map[string]bool          // per-guild "playback in progress" flag
	guildVC     map[string]*voice.Conn   // per-guild active voice connection
	skipSignals map[string]chan struct{} // per-guild skip signal for current playback

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
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			slog.Error("Error loading AWS config", slog.Any("err", err))
		}
		pollyClient = polly.NewFromConfig(cfg)
	}

	queues = make(map[string][]*Queue)
	guildActive = make(map[string]bool)
	guildVC = make(map[string]*voice.Conn)
	skipSignals = make(map[string]chan struct{})
}

// QueueItem is a read-only snapshot of a pending queue entry for display purposes.
type QueueItem struct {
	Position int
	UserID   string
	Type     string
	Snippet  string // first 100 chars of Data
}

// GetQueueItems returns a snapshot of the pending queue for a guild.
func GetQueueItems(guildID string) []QueueItem {
	mu.Lock()
	defer mu.Unlock()
	items := queues[guildID]
	result := make([]QueueItem, len(items))
	for i, item := range items {
		snippet := item.sqsObject.Data
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		result[i] = QueueItem{
			Position: i + 1,
			UserID:   item.userID,
			Type:     item.sqsObject.Type,
			Snippet:  snippet,
		}
	}
	return result
}

// FlushQueue clears the pending queue for a guild without interrupting current playback.
func FlushQueue(guildID string) {
	mu.Lock()
	defer mu.Unlock()
	queues[guildID] = nil
}

// SkipCurrent signals the active playback goroutine for a guild to stop the current item.
func SkipCurrent(guildID string) {
	mu.Lock()
	ch := skipSignals[guildID]
	mu.Unlock()
	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
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

	switch *sub.SubCommandName {
	case "user":
		speakObject.Type = "user"
		user := sub.Options["user"].Snowflake()

		speakObject.Data = user.String()
		resp, err := util.ConfigFile.SendStatsBotRequest(speakObject)
		speakObject.Data = resp.Data
		if err != nil {
			slog.Error("Encountered an error while processing the speak command", slog.Any("err", err))
			return
		}
	case "redditpost":
		redditPost := sub.Options["postid"].String()
		speakObject.Data = redditPost
		speakObject.Type = "redditpost"
	case "url":
		url := sub.Options["url"].String()
		speakObject.Data = markov.HandleURL(url)
		speakObject.Type = "url"
	case "chat":
		speakObject.Type = "chat"
		message := sub.Options["message"].String()
		data, err := chat.GetLLMResponse("speak", message, event.User().ID.String())
		if err != nil {
			slog.Error("Error interacting with LLM", slog.Any("err", err))
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

	_, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.MessageUpdate{
		Content: util.Pointer("Added to speech queue"),
	})
	if err != nil {
		slog.Error("Error editing the response:", slog.Any("err", err))
	}

	go synthData(speakObject, event.User().ID.String(), event.Client())
}

func (s SpeakCommand) CreateCommandArguments() []discord.ApplicationCommandOption {
	return []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "url",
			Description: "Run the speech from the text of a site",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "url",
					Description: "URL of the page to make a markov chain from",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "user",
			Description: "Run the speech from one of the user messages",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "User to create a markov chain of",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "redditpost",
			Description: "Speech from a reddit post",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "postid",
					Description: "Reddit post ID",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "chat",
			Description: "Have copypastabot read the text",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "message",
					Description: "Message for the copypastabot",
					Required:    true,
				},
			},
		},
	}
}

// synthData synthesizes all content chunks via Polly, appends them to the
// per-guild queue, and starts playback if the guild is not already active.
func synthData(object util.Object, userID string, bot *bot.Client) {
	if object.Type == "redditpost" {
		post := util.GetRedditPost(object.Data)
		object.Data = post.Post.Body
	}

	dbID, err := database.EnqueueSpeakItem(database.QueueRecord{
		GuildID: object.GuildID,
		UserID:  userID,
		Content: object.Data,
		CmdType: object.Type,
		CmdName: object.Command,
		AppID:   object.ApplicationID,
		Token:   object.Token,
	})
	if err != nil {
		slog.Error("Error persisting speak item to DB", slog.Any("err", err))
	}

	contents := util.BreakContent(object.Data, 2950)
	for i, content := range contents {
		resp, err := util.WrapIntoSSML(content, "system")
		textType := types.TextTypeText
		engine := types.EngineNeural
		if err != nil {
			slog.Error("Error wrapping into SSML", slog.Any("err", err))
		} else {
			content = resp.Choices[0].Message.Content
			// if !strings.Contains(content, "<speak>") {
			// 	content = "<speak>" + content + "</speak>"
			// }
			textType = types.TextTypeSsml
			engine = types.EngineStandard
		}
		slog.Debug("Content to be synthesized", slog.String("content", content))
		synthed, err := pollyClient.SynthesizeSpeech(context.TODO(), &polly.SynthesizeSpeechInput{
			Text:         aws.String(content),
			TextType:     textType,
			OutputFormat: types.OutputFormatMp3,
			Engine:       engine,
			VoiceId:      types.VoiceIdMatthew,
			LanguageCode: types.LanguageCodeEnUs,
		})
		if err != nil {
			slog.Error("Error synthesizing content", slog.Any("err", err))
			continue
		}

		isLast := i == len(contents)-1
		audioBytes, err := io.ReadAll(synthed.AudioStream)
		synthed.AudioStream.Close()
		if err != nil {
			slog.Error("Error reading Polly audio stream", slog.Any("err", err))
			continue
		}

		chunkID, err := database.StoreSpeakChunk(dbID, i, audioBytes, isLast)
		if err != nil {
			slog.Error("Error storing audio chunk in DB", slog.Any("err", err))
			continue
		}

		mu.Lock()
		queues[object.GuildID] = append(queues[object.GuildID], &Queue{
			chunkID:     chunkID,
			userID:      userID,
			sqsObject:   object,
			dbID:        dbID,
			isLastChunk: isLast,
		})
		mu.Unlock()
	}

	// Atomically check-and-set the active flag so only one goroutine starts doSpeech per guild.
	mu.Lock()
	shouldStart := !guildActive[object.GuildID]
	if shouldStart {
		guildActive[object.GuildID] = true
	}
	mu.Unlock()

	if shouldStart {
		doSpeech(object.GuildID, bot)
	}
}

func doSpeech(guildID string, bot *bot.Client) {
	mu.Lock()
	if len(queues[guildID]) == 0 {
		guildActive[guildID] = false
		mu.Unlock()
		slog.Debug("Queue is empty", slog.String("guild_id", guildID))
		return
	}
	currentSpeech := queues[guildID][0]
	queues[guildID] = queues[guildID][1:]
	mu.Unlock()

	d := "Playing now"
	response := discord.MessageUpdate{
		Content: &d,
	}
	data, _ := json.Marshal(response)
	if _, snerr := strconv.Atoi(currentSpeech.sqsObject.Token); snerr != nil {
		_, err := util.SendRequest("PATCH", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
		if err != nil {
			slog.Error("Error sending patch", slog.Any("err", err))
		}
		defer func() {
			_, err := util.SendRequest("DELETE", currentSpeech.sqsObject.ApplicationID, currentSpeech.sqsObject.Token, util.WEBHOOK, data)
			if err != nil {
				slog.Error("Error sending delete", slog.Any("err", err))
			}
		}()
	}

	vs, found := bot.Caches.VoiceState(snowflake.MustParse(guildID), snowflake.MustParse(currentSpeech.userID))
	if !found {
		// User left the VC; skip this item and try the next one.
		doSpeech(guildID, bot)
		return
	}

	// Sleep for a specified amount of time before playing the sound.
	time.Sleep(250 * time.Millisecond)

	audioBytes, err := database.GetSpeakChunkAudio(currentSpeech.chunkID)
	if err != nil {
		slog.Error("Error fetching chunk audio from DB", slog.Any("err", err))
		doSpeech(guildID, bot)
		return
	}
	if err := database.DeleteSpeakChunk(currentSpeech.chunkID); err != nil {
		slog.Error("Error deleting chunk audio from DB", slog.Any("err", err))
	}

	conn := bot.VoiceManager.CreateConn(snowflake.MustParse(guildID))

	err = conn.Open(context.TODO(), *vs.ChannelID, false, false)

	// Stay connected between items; disconnect only when the queue is drained.
	mu.Lock()
	hasMore := len(queues[guildID]) > 0
	mu.Unlock()
	if !hasMore {
		defer conn.Close(context.TODO())
	}

	if err != nil {
		slog.Error("Error joining voice channel", slog.Any("err", err))
	}

	mu.Lock()
	guildVC[guildID] = &conn
	mu.Unlock()

	err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagMicrophone)
	if err != nil {
		slog.Error("Error setting speaking to true", slog.Any("err", err))
	}

	encodeSession, err := dca.EncodeMem(bytes.NewReader(audioBytes), dca.StdEncodeOptions)
	if err != nil {
		slog.Error("Error DCA encoding audio", slog.Any("err", err))
	}

	skipCh := make(chan struct{}, 1)
	mu.Lock()
	skipSignals[guildID] = skipCh
	mu.Unlock()
	defer func() {
		mu.Lock()
		if skipSignals[guildID] == skipCh {
			skipSignals[guildID] = nil
		}
		mu.Unlock()
	}()

	writeOpus(encodeSession, conn.UDP(), skipCh)

	if currentSpeech.isLastChunk {
		if err := database.SetSpeakItemStatus(currentSpeech.dbID, "done"); err != nil {
			slog.Error("Error marking speak item as done", slog.Any("err", err))
		}
	}

	err = conn.SetSpeaking(context.TODO(), voice.SpeakingFlagNone)
	if err != nil {
		slog.Error("Error setting speaking to false", slog.Any("err", err))
	}

	mu.Lock()
	guildVC[guildID] = nil
	mu.Unlock()

	// Sleep for a specified amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	doSpeech(guildID, bot)
}

// Reenqueue restores pending speak items into the in-memory queue on startup.
// Chunks already synthesized and stored in the DB are reused; items whose chunks
// were never stored (crash mid-synthesis) are re-synthesized as a fallback.
func Reenqueue(items []database.QueueRecord, bot *bot.Client) {
	for _, item := range items {
		chunks, err := database.GetSpeakChunksForQueue(item.ID)
		if err != nil {
			slog.Error("Error fetching chunks for speak item", slog.Any("err", err))
			continue
		}

		obj := util.Object{
			Type:          item.CmdType,
			Command:       item.CmdName,
			Data:          item.Content,
			GuildID:       item.GuildID,
			ApplicationID: item.AppID,
			Token:         item.Token,
		}

		if len(chunks) == 0 {
			// Chunks were never stored (crashed during synthesis); re-synthesize.
			if err := database.SetSpeakItemStatus(item.ID, "done"); err != nil {
				slog.Error("Error marking speak item as done", slog.Any("err", err))
			}
			go synthData(obj, item.UserID, bot)
			continue
		}

		// Reload stored chunks into the in-memory queue.
		mu.Lock()
		for _, chunk := range chunks {
			queues[item.GuildID] = append(queues[item.GuildID], &Queue{
				chunkID:     chunk.ID,
				userID:      item.UserID,
				sqsObject:   obj,
				dbID:        item.ID,
				isLastChunk: chunk.IsLastChunk,
			})
		}
		shouldStart := !guildActive[item.GuildID]
		if shouldStart {
			guildActive[item.GuildID] = true
		}
		mu.Unlock()

		if shouldStart {
			go doSpeech(item.GuildID, bot)
		}
	}
}

func writeOpus(encodeSession *dca.EncodeSession, w io.Writer, done <-chan struct{}) {
	defer encodeSession.Cleanup()

	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	var frameLen int16
	// Don't wait for the first tick, run immediately.
	for ; true; <-ticker.C {
		select {
		case <-done:
			return
		default:
		}

		err := binary.Read(encodeSession, binary.LittleEndian, &frameLen)
		if err != nil {
			slog.Error("Error reading opus file", slog.Any("err", err))
			return
		}

		// Copy the frame.
		_, err = io.CopyN(w, encodeSession, int64(frameLen))
		if err != nil && err != io.EOF {
			return
		}
	}
}
