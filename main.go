package main

import
(
	"fmt"
	"github.com/godbus/dbus"
	"os"
  "strings"
)

const dest = "org.mpris.MediaPlayer2.spotify"
const path = "/org/mpris/MediaPlayer2"
const memb = "org.mrpis.MediaPlayer2.Player"

func PerformAction(command string) {
  conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	call := obj.Call("org.mpris.MediaPlayer2.Player."+command, 0)
	if call.Err != nil {
		fmt.Fprintln(os.Stderr, "Failed to add match:", call.Err)
		os.Exit(1)
	}
}


type Metadata struct {
  Artist string
  Title string
  Rating int
  Status string
  Url string
  ArtUrl string
  ArtFile string
}

func (c *Metadata) Current() {
  song := SongInfo()
  pstatus := Status()
  songData := song.Value().(map[string]dbus.Variant)
  c.Artist = songData["xesam:artist"].Value().([]string)[0]
  c.Title = songData["xesam:title"].Value().(string)
  c.Rating = int(songData["xesam:autoRating"].Value().(float64) * 100)
  c.Status = pstatus.Value().(string)
  c.Url = songData["xesam:url"].Value().(string)
  c.ArtUrl = songData["mpris:artUrl"].Value().(string)

  idx := strings.LastIndex(c.ArtUrl, "/")
  c.ArtFile = c.ArtUrl[idx+1:]
}

func Status() *dbus.Variant {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
  pstatus, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
  if err != nil {
    panic(err)
  }
  return &pstatus
}

func SongInfo() *dbus.Variant {
	conn, _ := dbus.SessionBus()
	obj := conn.Object(dest, path)
	song, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
  if err != nil {
    panic(err)
  }

  return &song
}

//TODO: fix listener
func (c *Metadata) Listener() {
	conn, _ := dbus.SessionBus()
  for _, v := range []string{"signal"} {
		call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
			"eavesdrop='true',type='"+v+"',path='/org/mpris/MediaPlayer2'")
		if call.Err != nil {
			fmt.Fprintln(os.Stderr, "Failed to add match:", call.Err)
			os.Exit(1)
		}
  }
  ch := make(chan *dbus.Message, 10)
  conn.Eavesdrop(ch)
  fmt.Println("Listening for everything")
  for v := range ch {
    if v != nil {
      fmt.Println(v)
      // add printing of current song here
    } else {
      fmt.Println("Something went very wrong.")
    }
  }
}

func (c *Metadata) Print() {
  c.Current()
  fmt.Printf("%s - %s", c.Artist, c.Title)
}

func main() {
  S := new(Metadata)

  if os.Args[0] == nil {
    S.Print()
    os.Exit(0)
  }
	flag := os.Args[1]
	opt := map[string]string{
		"next":    "Next",
		"prev":    "Previous",
		"play":    "PlayPause",
		"current": "current",
		"listen": "listen",
	}


	if opt[flag] == "current" {
    S.Print()
		os.Exit(0)
	}

  if opt[flag] == "listen" {
    S.Listener()
    os.Exit(0)
  }

	if opt[flag] != "" {
		PerformAction(opt[flag])
		os.Exit(0)
	}

	/** TODO: write a listener function **/
	/** TODO: write the metadata function **/

}
