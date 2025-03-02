package database

import "gorm.io/gorm"

type Platform string

const (
	Twitch    Platform = "Twitch"
	Youtube   Platform = "Youtube"
	X         Platform = "X"
	TikTok    Platform = "TikTok"
	Instagram Platform = "Instagram"
)

type User struct {
	gorm.Model
	Username  string `gorm:"unique;not null"`
	Email     string `gorm:"unique;not null"`
	DiscordID string `gorm:"unique;not null"`
}

type Session struct {
	gorm.Model
	SessionID string `gorm:"unique;not null"`
	User      User
	UserID    uint64
}

type PlatformConnection struct {
	gorm.Model
	Platform   Platform
	PlatformID string `gorm:"unique;not null"`
	User       User
	UserID     uint64
}
