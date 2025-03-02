package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"go-backend-discord/commands"
	"go-backend-discord/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"os"
)

var dg *discordgo.Session
var s *http.ServeMux

var db *gorm.DB

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

	db, err = gorm.Open(sqlite.Open("./sqlite/main.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&models.User{})

}
func main() {
	dg.AddHandler(interactHandler)

	//dg.AddHandler()
	err := dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
	}
	defer dg.Close()

	s.HandleFunc("GET /user", UserHandler)
	s.HandleFunc("POST /users", CreateUserHandler)
	s.HandleFunc("GET /users/all", GetAllUsersHandler)
	s.HandleFunc("DELETE /users/delete/empty", DeleteUserEmptyHandler)
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

func UserHandler(w http.ResponseWriter, r *http.Request) {
	var id = r.URL.Query().Get("id")
	if id == "" {
		id = "179031614683217920"
	}
	if dg == nil {
		fmt.Println("dg is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user, err := dg.User(id)
	if err != nil {
		fmt.Println("error getting user,", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	data, derr := json.Marshal(user)
	if derr != nil {
		fmt.Println("error marshalling user,", derr)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var user = models.User{
		Username: r.Form.Get("username"),
		Email:    r.Form.Get("email"),
	}

	db.Create(&user)
	var created models.User

	err := db.Find(&created, user.ID).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	data, derr := json.Marshal(created)
	if derr != nil {
		fmt.Println("error marshalling user,", derr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	err := db.Find(&users).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func DeleteUserEmptyHandler(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	err := db.Find(&users).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		if user.Username == "" {
			err := db.Delete(&user, user.ID).Error
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}
