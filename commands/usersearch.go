package commands

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

var UserSearch Command = Command{
	Id:      "usersearch",
	Command: UserSearchCommand,
	Handler: UserSearchHandler,
}

var UserSearchCommand = &discordgo.ApplicationCommand{
	Name:        "usersearch",
	Description: "Search for users",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "user",
			Description: "User to search for",
			Required:    false,
		},
	},
	DefaultMemberPermissions: &DefaultMemberPermissions,
}

var DefaultMemberPermissions int64 = discordgo.PermissionManageServer

func UserSearchHandler(s *discordgo.Session, i *discordgo.InteractionCreate, opts OptionMap) {
	var id = "@me"
	if opts["user"] != nil {
		id = opts["user"].StringValue()
	}

	user, err := s.User(id)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error getting user",
			},
		})
		return
	}

	member, merr := s.GuildMember(i.GuildID, user.ID)
	if merr != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error getting member",
			},
		})
		return
	}

	ts, terr := discordgo.SnowflakeTimestamp(user.ID)
	if terr != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error getting timestamp",
			},
		})
		return
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.Username,
			IconURL: user.AvatarURL("64x64"),
		},
		Color: user.AccentColor,
		Title: "User Search",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "ID",
				Value:  user.ID,
				Inline: true,
			},
			{
				Name:   "Name",
				Value:  user.Username,
				Inline: true,
			},
			{
				Name:  "Joined At",
				Value: member.JoinedAt.Format(time.RFC3339),
			},
			{
				Name:  "Created At",
				Value: ts.Format(time.RFC3339),
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: user.BannerURL("4096"),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL("64x64"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Searched by " + user.Username,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	resp := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: resp,
	})
}
