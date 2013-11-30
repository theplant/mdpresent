package present

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var fixture = `presentation
14:59 22 Nov 2013

Van5 Hu
# 渠道薪资待遇

# 提成：到账金额*提成比例

- 当月到账预存款金额

# 提成比例

- 到账金额≥15万
4.00%

- 10万≤到账金额＜15万
3.00%

- 5万≤到账金额＜10万
2.50%

- 到账金额＜5万
2.00%

# 通讯补贴 200元/月，当月出勤需在15天以上；
# 出差补贴 80元/天。
`

var result = `{"Title":"presentation","Subtitle":"","Time":"2013-11-22T14:59:00Z","Authors":[{"Elem":[{"Lines":["Van5 Hu"],"Pre":false}]}],"Sections":[{"Number":[1],"Title":"渠道薪资待遇","Elem":null},{"Number":[2],"Title":"提成：到账金额*提成比例","Elem":[{"Bullet":["当月到账预存款金额"]}]},{"Number":[3],"Title":"提成比例","Elem":[{"Bullet":["到账金额≥15万\n4.00%","10万≤到账金额＜15万\n3.00%","5万≤到账金额＜10万\n2.50%","到账金额＜5万\n2.00%"]}]},{"Number":[4],"Title":"通讯补贴 200元/月，当月出勤需在15天以上；","Elem":null},{"Number":[5],"Title":"出差补贴 80元/天。","Elem":null}]}`

func TestParseSeperateListItems(t *testing.T) {
	r := strings.NewReader(fixture)
	doc, err := Parse(r, "test", 0)
	if err != nil {
		panic(err)
	}
	expected, _ := json.Marshal(doc)
	if string(expected) != result {
		t.Error("Do not generate expected result")
	}

	// if len(doc.Sections) != 4 {
	// 	t.Error("Should only parse out four sections, but have ", len(doc.Sections))
	// }
	// if len(doc.Sections[1].Elem) != 1 && len(doc.Sections[1].Elem) == 4 {
	// 	t.Errorf("Should only parse out a list with 4 bullets, but have %+v\n", len(doc.Sections[1].Elem))
	// }
}

// var fixture2 = `Head
// 16:13 28 Nov 2013
// Author
// # Title
// * a point
// * second point
// * third point

// Everything I know.
// `

// func TestParseCompactListItems(t *testing.T) {
// 	r := strings.NewReader(fixture2)
// 	doc, err := Parse(r, "test", 0)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%+v\n", doc)
// }
