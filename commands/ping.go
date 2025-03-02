package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

var Ping Command = Command{
	Id:      PingCommandId,
	Command: PingCommand,
	Handler: PingHandler,
}
var PingCommandId = "ping"
var PingCommand = &discordgo.ApplicationCommand{
	Name:        PingCommandId,
	Description: "Ping",
}

func PingHandler(session *discordgo.Session, i *discordgo.InteractionCreate, opts OptionMap) {
	err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong! Again!",
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}
