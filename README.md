# sps

A CLI tool to control Spotify and Google Play Music via DBus on Linux

## Installation

Install with

`go get github.com/enjuus/sps`


## Usage

```
Usage: sps [option]

arguments:
	play 		Play or pause the current song
	next		Skip to the next song
	prev 		Skip to the previous song
	current 	Print the current song
	listen 		Event listener that outputs the current song
	url 		The Spotify/GPM album cover URL
	album 		The title of the current album
	status 		The status if a song is playing or paused
```