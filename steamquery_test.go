package steamquery

import (
	"fmt"
	"testing"
)

const TESTSERV = "109.70.149.165:27095"

func TestQueryString(t *testing.T) {
	r, err := QueryInfoString(TESTSERV)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", r)
}

func TestPlayersQueryString(t *testing.T) {
	ps, err := QueryPlayersString(TESTSERV)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		fmt.Printf("%+v\n", p)
	}
}
