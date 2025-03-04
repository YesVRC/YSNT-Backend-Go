package main

import (
	_ "embed"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"go-backend-discord/commands"
	"go-backend-discord/modules/database"
	"go-backend-discord/modules/routes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"os"
)

var dg *discordgo.Session
var s *http.ServeMux

func init() {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	s = http.NewServeMux()

	dg, err = discordgo.New(os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}
	_, err = dg.ApplicationCommandBulkOverwrite(os.Getenv("DISCORD_BOT_ID"), os.Getenv("DISCORD_DEBUG_GUILD"), commands.Registry.GetCommands())
	if err != nil {
		panic(err)
	}

	database.Db, err = gorm.Open(sqlite.Open("./sqlite/main.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	database.Db.AutoMigrate(&database.User{}, &database.Session{}, &database.PlatformConnection{})

}
func main() {
	dg.AddHandler(interactHandler)

	err := dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
	}
	defer dg.Close()

	s.HandleFunc("GET /user", UserHandler)

	RegisterRoutes(s, routes.AuthRoutes)
	RegisterRoutes(s, routes.LiveRoutes)
	err = http.ListenAndServe(":8080", s)
	if err != nil {
		return
	}
}

func interactHandler(session *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	if commands.Registry.GetHandlers()[data.Name] == nil {
		return
	}

	fmt.Printf("Handling interaction %s\n", data.Name)
	commands.Registry.GetHandlers()[data.Name](session, i, commands.ParseOptions(data.Options))
}

func RegisterRoutes(s *http.ServeMux, routes []routes.Route) {
	for _, route := range routes {
		final := route.Handler
		for _, middleware := range route.Middleware {
			final = middleware(final)
		}
		s.Handle(route.Path, final)
	}
}
