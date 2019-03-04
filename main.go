package main

import "C"
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/godbus/dbus"
	"github.com/urfave/cli"
)

const dest = "org.mpris.MediaPlayer2.spotify"
const path = "/org/mpris/MediaPlayer2"
const memb = "org.mpris.MediaPlayer2.Player"

type metadata struct {
	Artist  string
	Title   string
	Rating  int
	Status  string
	URL     string
	ArtURL  string
	ArtFile string
	Album   string
	Volume  int
}

// Perform a command/dbus method against the MediaPlayer2 interface
func performAction(command string) {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	call := obj.Call(memb+"."+command, 0)
	if call.Err != nil {
		switch call.Err.(type) {
		case dbus.Error:
			obj := conn.Object("org.mpris.MediaPlayer2.google-play-music-desktop-player", path)
			call := obj.Call(memb+"."+command, 0)
			fmt.Println(call.Done)
			if call.Err != nil {
				fmt.Println("No media player is currently running")
				os.Exit(1)
			}
		default:
			fmt.Println("What the h* just happened?")
			os.Exit(1)
		}
	}
}

// Retrieve info dbus.Property
func retrieveInfo(info string) *dbus.Variant {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	playerinfo, err := obj.GetProperty("org.mpris.MediaPlayer2.Player."+info)
	if err != nil {
		switch err.(type) {
		case dbus.Error:
			obj := conn.Object("org.mpris.MediaPlayer2.google-play-music-desktop-player", path)
			playerinfo, err := obj.GetProperty("org.mpris.MediaPlayer2.Player."+info)
			if err != nil {
				fmt.Println("No media player is currently running")
				os.Exit(1)
			}
			return &playerinfo
		default:
			fmt.Println("What the h* just happened?")
			os.Exit(1)
		}
	}
	return &playerinfo
}

// Update metadata for the currently playing song
func (c *metadata) current() {
	song := retrieveInfo("Metadata")
	pstatus := retrieveInfo("PlaybackStatus")
	volume := retrieveInfo("Volume")

	songData, _ := song.Value().(map[string]dbus.Variant)
	if songData["xesam:artist"].Value() != nil {
		c.Artist = songData["xesam:artist"].Value().([]string)[0]
		c.Title = songData["xesam:title"].Value().(string)
		c.Album = songData["xesam:album"].Value().(string)
		c.Volume = int(volume.Value().(float64) * 100)
		if songData["xesam:autoRating"].Value() != nil {
			c.Rating = int(songData["xesam:autoRating"].Value().(float64) * 100)
		}
		c.Status = pstatus.Value().(string)
		if songData["xesam:url"].Value() != nil {
			c.URL = songData["xesam:url"].Value().(string)
		}
		c.ArtURL = songData["mpris:artUrl"].Value().(string)

		idx := strings.LastIndex(c.ArtURL, "/")
		c.ArtFile = c.ArtURL[idx+1:]
	} else { log.Println("Start playing a song..")}
}

func downloadFile(filename string, url string) error {

	usr, _ := user.Current()
	dir := usr.HomeDir
	path := filepath.Join(dir+"/", filename)

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

//TODO: fix listener
func (c *metadata) listener() {
	conn, _ := dbus.SessionBus()
	sptfy := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"sender='"+dest+"', path='/org/mpris/MediaPlayer2', type='signal', member='PropertiesChanged'")
	gpmdp := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"sender='org.mpris.MediaPlayer2.google-play-music-desktop-player', path='/org/mpris/MediaPlayer2', type='signal', member='PropertiesChanged'")
	//apparently we never go here?
	if sptfy.Err != nil { //apparently we never go here?
		fmt.Println(os.Stderr, "failed to add match: ", sptfy.Err)
		os.Exit(1)
	}
	//or here
	if gpmdp.Err != nil {
		fmt.Println(os.Stderr, "failed to add match: ", gpmdp.Err)
		os.Exit(1)
	}
	ch := make(chan *dbus.Signal, 5)
	conn.Signal(ch)
	c.print()
	current := fmt.Sprintf("%s - %s", c.Artist, c.Title)
	for v := range ch {
		if v != nil {
			// Not the nicest solution to do it this way, but dbus
			// keeps giving out multiple signals
			c.current()
			if current != fmt.Sprintf("%s - %s", c.Artist, c.Title) {
				c.print()
				c.getAlbumArt()
				current = fmt.Sprintf("%s - %s", c.Artist, c.Title)
			}
		} else {
			fmt.Println("Something went very wrong.")
		}
	}
}

func (c *metadata) print() {
	c.current()
	fmt.Println(c.Artist+" - "+c.Title)
}

func (c *metadata) getAlbumArt() {
	c.current()
	err := downloadFile("np.png", c.ArtURL)
	if err != nil {
		panic(err)
	}
}

func main() {
	S := new(metadata)

	if len(os.Args) == 1 {
		S.current()
		S.print()
		os.Exit(0)
	}

	app := cli.NewApp()
	app.Name = "sps"
	app.Usage = "Commandline interface to Spotify/GPMDP"
	app.UsageText = "sps [command]"
	app.HideVersion = true
	app.Commands = []cli.Command {
		{
			Name: "current",
			Usage: "Returns currently playing song",
			Action: func(c *cli.Context) error {
				S.print()
				return nil
			},
		},
		{
			Name: "listen",
			Usage: "Starts in listening mode",
			Action: func(c *cli.Context) error {
				S.listener()
				return nil
			},
		},
		{
			Name: "url",
			Usage: "Print URL to album art",
			Action: func(c *cli.Context) error {
				S.current()
				fmt.Println(S.ArtURL)
				return nil
			},
		},
		{
			Name: "file",
			Usage: "Downloads the album art to $HOME/np.png",
			Action: func(c *cli.Context) error {
				S.getAlbumArt()
				return nil
			},
		},
		{
			Name: "album",
			Usage: "Print the album of the currently playing song",
			Action: func(c *cli.Context) error {
				S.current()
				fmt.Println(S.Album)
				return nil
			},
		},
		{
			Name: "status",
			Usage: "Print the player status",
			Action: func(c *cli.Context) error {
				S.current()
				fmt.Println(S.Status)
				return nil
			},
		},
		{
			Name: "volume",
			Aliases: []string{"vol"},
			Usage: "Show the current player volume",
			Action: func(c *cli.Context) error {
				S.current()
				fmt.Println(strconv.Itoa(S.Volume)+"%")
				return nil
			},
		},
		{
			Name: "next",
			Aliases: []string{"n"},
			Usage: "Skips to next song",
			Action: func(c *cli.Context) error {
				performAction("Next")
				return nil
			},
		},
		{
			Name: "previous",
			Aliases: []string{"p"},
			Usage: "Skips to the previous song",
			Action: func(c *cli.Context) error {
				performAction("Previous")
				return nil
			},
		},
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
