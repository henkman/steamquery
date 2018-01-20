package steamquery

import (
	"fmt"
	"testing"
)

const TESTSERV = "109.70.149.165:27095"

func TestQueryInfoString(t *testing.T) {
	r, ping, err := QueryInfoString(TESTSERV)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("answered in:", ping)
	fmt.Printf("%+v\n", r)
}

func TestPlayersQueryString(t *testing.T) {
	ps, ping, err := QueryPlayersString(TESTSERV)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("answered in:", ping)
	for _, p := range ps {
		fmt.Printf("%+v\n", p)
	}
}

func TestQueryRulesString(t *testing.T) {
	rs, ping, err := QueryRulesString(TESTSERV)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("answered in:", ping)
	for _, r := range rs {
		fmt.Printf("%s=%s\n", r.Name, r.Value)
	}
}
