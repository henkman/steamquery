package steamquery

import (
	"fmt"
	"testing"
)

func TestQueryString(t *testing.T) {
	const SERVER = "37.114.96.46:27019"
	r, err := QueryString(SERVER)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", r)
}
