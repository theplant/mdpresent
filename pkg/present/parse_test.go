package present

import (
	"fmt"
	"strings"
	"testing"
)

var fixture = `Head
16:13 28 Nov 2013
Author
# Title
* a point

* second point

* third point
`

func TestParseSeperateListItems(t *testing.T) {
	r := strings.NewReader(fixture)
	doc, err := Parse(r, "test", 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", doc)
}

var fixture2 = `Head
16:13 28 Nov 2013
Author
# Title
* a point
* second point
* third point

Everything I know.
`

func TestParseCompactListItems(t *testing.T) {
	r := strings.NewReader(fixture2)
	doc, err := Parse(r, "test", 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", doc)
}
