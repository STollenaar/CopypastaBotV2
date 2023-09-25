package util

import "github.com/bwmarrin/discordgo"

// MessageObject general messageobject for functions
type MessageObject struct {
	GuildID   string               `bson:"GuildID" json:"GuildID"`
	ChannelID string               `bson:"ChannelID" json:"ChannelID"`
	MessageID string               `bson:"_id" json:"MessageID"`
	Author    string               `bson:"Author" json:"Author"`
	Content   []string             `bson:"Content" json:"Content"`
	Date      discordgo.TimeStamps `bson:"Date" json:"Date"`
}

type ResponseObject struct {
	Data discordgo.InteractionResponseData `json:"data"`
	Type discordgo.InteractionResponseType `json:"type"`
}

type EditResponseObject struct {
	Data discordgo.WebhookEdit             `json:"data"`
	Type discordgo.InteractionResponseType `json:"type"`
}
