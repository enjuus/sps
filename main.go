package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus"
)

const dest = "org.mpris.MediaPlayer2.google-play-music-desktop-player"
const path = "/org/mpris/MediaPlayer2"
const memb = "org.mrpis.MediaPlayer2.Player"

func PerformAction(command string) {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	call := obj.Call("org.mpris.MediaPlayer2.Player."+command, 0)
	if call.Err != nil {
		fmt.Println("GPMDP is not running.")
		os.Exit(1)
	}
}

type Metadata struct {
	Artist  string
	Title   string
	Rating  int
	Status  string
	Url     string
	ArtUrl  string
	ArtFile string
	Album   string
}

func (c *Metadata) Current() {
	song := SongInfo()
	pstatus := Status()
	songData := song.Value().(map[string]dbus.Variant)
	c.Artist = songData["xesam:artist"].Value().([]string)[0]
	c.Title = songData["xesam:title"].Value().(string)
	c.Album = songData["xesam:album"].Value().(string)
	//c.Rating = int(songData["xesam:autoRating"].Value().(float64) * 100)
	c.Status = pstatus.Value().(string)
	//c.Url = songData["xesam:url"].Value().(string)
	c.ArtUrl = songData["mpris:artUrl"].Value().(string)

	idx := strings.LastIndex(c.ArtUrl, "/")
	c.ArtFile = c.ArtUrl[idx+1:]
}

func Status() *dbus.Variant {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	pstatus, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		fmt.Println("GPMDP is not running.")
		os.Exit(1)
	}
	return &pstatus
}

func SongInfo() *dbus.Variant {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	song, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		fmt.Println("GPMDP is not running.")
		os.Exit(1)
	}

	return &song
}

func DownloadFile(filename string, url string) error {

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
func (c *Metadata) Listener() {
	conn, _ := dbus.SessionBus()
	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"sender='"+dest+"', path='/org/mpris/MediaPlayer2', type='signal', member='PropertiesChanged'")
	if call.Err != nil {
		fmt.Println(os.Stderr, "failed to add match: ", call.Err)
		os.Exit(1)
	}
	ch := make(chan *dbus.Signal, 5)
	conn.Signal(ch)
	c.Print()
	current := fmt.Sprintf("%s - %s", c.Artist, c.Title)
	for v := range ch {
		if v != nil {
			// Not the nicest solution to do it this way, but dbus
			// keeps giving out multiple signals
			c.Current()
			if current != fmt.Sprintf("%s - %s", c.Artist, c.Title) {
				c.Print()
				c.GetAlbumArt()
				current = fmt.Sprintf("%s - %s", c.Artist, c.Title)
			}
		} else {
			fmt.Println("Something went very wrong.")
		}
	}
}

func (c *Metadata) Print() {
	c.Current()
	fmt.Println(c.Artist, "-", c.Title)
}

func (c *Metadata) PrintArtUrl() {
	c.Current()
	fmt.Println(c.ArtUrl)
}

func (c *Metadata) PrintArtFile() {
	c.Current()
	fmt.Println(c.ArtFile)
}

func (c *Metadata) PrintAlbum() {
	c.Current()
	fmt.Println(c.Album)
}

func (c *Metadata) GetAlbumArt() {
	c.Current()
	err := DownloadFile("np.png", c.ArtUrl)
	if err != nil {
		panic(err)
	}
}

func (c *Metadata) PrintStatus() {
	c.Current()
	fmt.Println(c.Status)
}

func main() {
	S := new(Metadata)

	if len(os.Args) == 1 {
		S.Print()
		os.Exit(0)
	}

	flag := os.Args[1]
	opt := map[string]string{
		"next":    "Next",
		"prev":    "Previous",
		"play":    "PlayPause",
		"current": "current",
		"listen":  "listen",
		"url":     "url",
		"file":    "file",
		"album":   "album",
		"status":  "status",
	}

	if opt[flag] == "current" {
		S.Print()
		os.Exit(0)
	}

	if opt[flag] == "url" {
		S.PrintArtUrl()
		os.Exit(0)
	}

	if opt[flag] == "file" {
		S.GetAlbumArt()
		os.Exit(0)
	}

	if opt[flag] == "album" {
		S.PrintAlbum()
		os.Exit(0)
	}

	if opt[flag] == "listen" {
		S.Listener()
		os.Exit(0)
	}

	if opt[flag] == "status" {
		S.PrintStatus()
		os.Exit(0)
	}

	if opt[flag] != "" {
		PerformAction(opt[flag])
		os.Exit(0)
	}
}
