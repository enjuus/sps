package main

import
(
	"fmt"
	"github.com/godbus/dbus"
	"os"
	"encoding/json"
)

const dest = "org.mpris.MediaPlayer2.spotify"
const path = "/org/mpris/MediaPlayer2"
const memb = "org.mrpis.MediaPlayer2.Player"

func PerformAction(command string, conn *dbus.Conn) {
	obj := conn.Object(dest, path)
	call := obj.Call("org.mpris.MediaPlayer2.Player."+command, 0)
	if call.Err != nil {
		fmt.Fprintln(os.Stderr, "Failed to add match:", call.Err)
		os.Exit(1)
	}
}

type Metadata struct {
	TrackID string `json:"mpris:trackid"`
	Length uint64 `json:"mpris:length"`
	ArtUrl string `json:"mpris:artUrl"`
	Album string `json:"xesam:album"`
	AlbumArtist map[string]string `json:"xesam:albumArtist"`
	Artist map[string]string `json:"xesam:artist"`
	AutoRating float64 `json:"xesam:autoRating"`
	DiscNumber int32 `json:"xesam:discNumber"`
	Title string `json:"xesam:title"`
	TrackNumber uint32 `json:"xesam:trackNumber"`
	URL string `json:"xesam:url"`
}

func CurrentlyPlaying(conn *dbus.Conn) {
	obj := conn.Object(dest, path)
	call := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.Player", "Metadata")
	dat := fmt.Sprint(call.Body[0])
	fmt.Println(dat)
	// TODO: pls no more i  cant take it ffs
	in := `{"mpris:artUrl": "https://open.spotify.com/image/ea5eaa59a06f0406e6e8619caad316d44fb781d7", "mpris:length": 280693000, "mpris:trackid": "spotify:track:36IIXP08n7pZGIeMa9ziqN", "xesam:album": "Kontakt", "xesam:albumArtist": ["Fjørt"], "xesam:artist": ["Fjørt"], "xesam:autoRating": 0.26, "xesam:discNumber": 1, "xesam:title": "Lebewohl", "xesam:trackNumber": 11, "xesam:url": "https://open.spotify.com/track/36IIXP08n7pZGIeMa9ziqN"}`
	bytes, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}

	var p Metadata
	err = json.Unmarshal(bytes, &p)
	if err != nil {
		panic(err)
	}

	fmt.Printf("&+v", p)
}

func main() {

	conn, _ := dbus.SessionBus()

	flag := os.Args[1]

	opt := map[string]string{
		"next":    "Next",
		"prev":    "Previous",
		"play":    "PlayPause",
		"current": "current",
	}

	if opt[flag] == "current" {
		CurrentlyPlaying(conn)
		os.Exit(0)
	}

	if opt[flag] != "" {
		PerformAction(opt[flag], conn)
		os.Exit(0)
	}

	/** TODO: write a listener function **/
	/** TODO: write the metadata function **/

}
