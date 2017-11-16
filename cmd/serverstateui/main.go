package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/henkman/steamquery"
	"github.com/zserge/webview"
)

type Info struct {
	Name    string
	Map     string
	Players int
}

type Server struct {
	Address string              `json:"address"`
	Info    Info                `json:"info"`
	Players []steamquery.Player `json:"players"`
}

func (s *Server) UpdatePlayers() {
	players, err := steamquery.QueryPlayersString(s.Address)
	if err != nil {
		s.Players = []steamquery.Player{}
		return
	}
	s.Players = players
}

func (s *Server) UpdateInfo() {
	rules, err := steamquery.QueryRulesString(s.Address)
	if err != nil {
		s.Info = Info{}
		return
	}
	var owningPlayerName string
	var p2 string
	var numOpenPublicConnections int
	var numPublicConnections int
	for _, r := range rules {
		if r.Name == "OwningPlayerName" {
			owningPlayerName = r.Value
		} else if r.Name == "NumOpenPublicConnections" {
			tmp, err := strconv.Atoi(r.Value)
			if err != nil {
				s.Info = Info{}
				return
			}
			numOpenPublicConnections = tmp
		} else if r.Name == "NumPublicConnections" {
			tmp, err := strconv.Atoi(r.Value)
			if err != nil {
				s.Info = Info{}
				return
			}
			numPublicConnections = tmp
		} else if r.Name == "p2" {
			p2 = r.Value
		}
	}
	s.Info = Info{
		Name:    owningPlayerName,
		Map:     p2,
		Players: numPublicConnections - numOpenPublicConnections,
	}
}

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
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(ex)
	url := "file:///" + filepath.ToSlash(dir) + "/r/index.html"
	w := webview.New(webview.Settings{
		Title:     "server state",
		URL:       url,
		Width:     650,
		Height:    800,
		Resizable: true,
	})
	defer w.Exit()
	w.Dispatch(func() {
		w.Bind("server", &Server{Address: _server})
	})
	w.Run()
}