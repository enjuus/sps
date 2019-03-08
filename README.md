# sps

A CLI tool to control Spotify and [GPMDP](https://www.googleplaymusicdesktopplayer.com/) via DBus on Linux

## Installation

Install with

`go get github.com/enjuus/sps`

## Usage

```man
NAME:
   sps - Commandline interface to Spotify/GPMDP

USAGE:
   sps [command]

COMMANDS:
     album        Print the album of the currently playing song
     current      Returns currently playing song
     file         Downloads the album art
     listen       Starts in listening mode
     next, n      Skips to next song
     previous, p  Skips to the previous song
     status       Print the player status
     url          Print URL to album art
     volume, vol  Show the current player volume
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```
