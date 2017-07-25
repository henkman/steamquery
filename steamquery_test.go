package steamquery

import (
	"fmt"
	"testing"
)

func TestQueryString(t *testing.T) {
	r, err := QueryString("37.114.96.46:27019")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", r)
}
