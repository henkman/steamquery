package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/henkman/steamquery"
)

var (
	_server string
)

func init() {
	flag.StringVar(&_server, "s", "", "server")
	flag.Parse()
}

func main() {
	if _server == "" {
		flag.Usage()
		return
	}
	r, err := steamquery.QueryString(_server)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Protocol:", r.Protocol)
	fmt.Println("Name:", r.Name)
	fmt.Println("Map:", r.Map)
	fmt.Println("Folder:", r.Folder)
	fmt.Println("Game:", r.Game)
	fmt.Println("ID:", r.ID)
	fmt.Println("Players:", r.Players)
	fmt.Println("MaxPlayers:", r.MaxPlayers)
	fmt.Println("Bots:", r.Bots)
	fmt.Printf("ServerType: %c\n", r.ServerType)
	fmt.Printf("Environment: %c\n", r.Environment)
	fmt.Println("Private:", r.Private)
	fmt.Println("VAC:", r.VAC)
	fmt.Println("Version:", r.Version)
	fmt.Println("Port:", r.Port)
	fmt.Println("SteamID:", r.SteamID)
	if r.SourceTV.Name != "" {
		fmt.Println("SourceTV.Name:", r.SourceTV.Name)
		fmt.Println("SourceTV.Port:", r.SourceTV.Port)
	}
	fmt.Println("Keywords:", r.Keywords)
	fmt.Println("GameID:", r.GameID)
}
