package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/hako/durafmt"
	"github.com/henkman/clitable"
	"github.com/henkman/steamquery"
)

var (
	_dontprintplayer  bool
	_sortplayerbyname bool
	_server           string
)

type SortPlayerByName []steamquery.Player

func (a SortPlayerByName) Len() int           { return len(a) }
func (a SortPlayerByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortPlayerByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type SortPlayerByScore []steamquery.Player

func (a SortPlayerByScore) Len() int           { return len(a) }
func (a SortPlayerByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortPlayerByScore) Less(i, j int) bool { return a[i].Score > a[j].Score }

type PlayerRow steamquery.Player

func (p PlayerRow) GetValue(i int) string {
	switch i {
	case 0:
		return p.Name
	case 1:
		return fmt.Sprint(p.Score)
	case 2:
		return durafmt.Parse(p.Duration).String()
	case 3:
		return fmt.Sprint(float32(p.Score) / float32(p.Duration.Seconds()))
	}
	panic("invalid index")
}

func init() {
	flag.BoolVar(&_dontprintplayer, "np", false, "don't print player names")
	flag.BoolVar(&_sortplayerbyname, "sn", false, "sort player by name instead of score")
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
	if !_dontprintplayer {
		fmt.Println("Players: ")
		ps, err := steamquery.QueryPlayersString(_server)
		if err != nil {
			log.Fatal(err)
		}
		if _sortplayerbyname {
			sort.Sort(SortPlayerByName(ps))
		} else {
			sort.Sort(SortPlayerByScore(ps))
		}
		rows := make([]clitable.Row, len(ps))
		for i, p := range ps {
			rows[i] = PlayerRow(p)
		}
		clitable.Write(os.Stdout,
			[]string{"Name", "Score", "Online", "Score/Second"},
			rows)
	}
}
