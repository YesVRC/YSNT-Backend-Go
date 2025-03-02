package commands

import "github.com/bwmarrin/discordgo"

type CommandRegistry []Command

type Handler func(s *discordgo.Session, i *discordgo.InteractionCreate, opts OptionMap)

type Command struct {
	Id      string
	Command *discordgo.ApplicationCommand
	Handler Handler
}

var Registry = CommandRegistry{
	Ping,
	UserSearch,
}
var HandlerCache map[string]Handler

type OptionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func (registry CommandRegistry) GetCommands() []*discordgo.ApplicationCommand {
	var Commands []*discordgo.ApplicationCommand
	for _, command := range registry {
		Commands = append(Commands, command.Command)
	}
	return Commands
}

func (registry CommandRegistry) GetHandlers() map[string]Handler {
	if HandlerCache == nil {
		HandlerCache = make(map[string]Handler)
		for _, command := range registry {
			HandlerCache[command.Id] = command.Handler
		}
	}
	return HandlerCache
}

func ParseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om OptionMap) {
	om = make(OptionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}
