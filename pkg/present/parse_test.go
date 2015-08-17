package present

import (
	"encoding/json"
	"strings"
	"testing"
)

var imageTextImage = []string{
	`bbb
9/24/2014

Felix Sun



#

![](/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/5422741b6c08982b400000be/m/undefined "undefined")

He Orders a beer.

![](/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/542274696c08982b400000da/m/undefined "undefined")

# He Orders 0 beers.

# Orders 999999999 beers.

![](/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/542274766c08982b400000e2/m/undefined "undefined")

# Orders a lizard.

![](/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/542274896c08982b400000f6/m/undefined "undefined")

# Orders -1 beers.

# Orders a sfdeljkn`,

	`{"Title":"bbb","Subtitle":"9/24/2014","Time":"0001-01-01T00:00:00Z","Authors":[{"Elem":[{"Lines":["Felix Sun"],"Pre":false}]},{"Elem":[{"Lines":[""],"Pre":false}]}],"Sections":[{"Number":[1],"Title":"","Elem":[{"URL":"/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/5422741b6c08982b400000be/m/undefined","Width":0,"Height":0},{"Lines":["He Orders a beer."],"Pre":false},{"URL":"/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/542274696c08982b400000da/m/undefined","Width":0,"Height":0}]},{"Number":[2],"Title":"He Orders 0 beers.","Elem":null},{"Number":[3],"Title":"Orders 999999999 beers.","Elem":[{"URL":"/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/542274766c08982b400000e2/m/undefined","Width":0,"Height":0}]},{"Number":[4],"Title":"Orders a lizard.","Elem":[{"URL":"/5018d345558fbe46c4000001/537f27fc18c8bca41a000010/file/542274896c08982b400000f6/m/undefined","Width":0,"Height":0}]},{"Number":[5],"Title":"Orders -1 beers.","Elem":null},{"Number":[6],"Title":"Orders a sfdeljkn","Elem":null}]}`,
}

var chineseTitleAndList = []string{
	`presentation
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
`,

	`{"Title":"presentation","Subtitle":"","Time":"2013-11-22T14:59:00Z","Authors":[{"Elem":[{"Lines":["Van5 Hu"],"Pre":false}]}],"Sections":[{"Number":[1],"Title":"渠道薪资待遇","Elem":null},{"Number":[2],"Title":"提成：到账金额*提成比例","Elem":[{"Bullet":["当月到账预存款金额"]}]},{"Number":[3],"Title":"提成比例","Elem":[{"Bullet":["到账金额≥15万\n4.00%","10万≤到账金额＜15万\n3.00%","5万≤到账金额＜10万\n2.50%","到账金额＜5万\n2.00%"]}]},{"Number":[4],"Title":"通讯补贴 200元/月，当月出勤需在15天以上；","Elem":null},{"Number":[5],"Title":"出差补贴 80元/天。","Elem":null}]}`,
}

var spaceBetweenListItem = []string{
	`Head
16:13 28 Nov 2013
Author
# Title
* a point

* second point

* third point`,

	`{"Title":"Head","Subtitle":"","Time":"2013-11-28T16:13:00Z","Authors":[{"Elem":[{"Lines":["Author"],"Pre":false}]}],"Sections":[{"Number":[1],"Title":"Title","Elem":[{"Bullet":["a point","second point","third point"]}]}]}`,
}

var noSpaceBetweenListItem = []string{
	`Head
16:13 28 Nov 2013
Author
# Title
* a point
* second point
* third point

Everything I know.
`,

	`{"Title":"Head","Subtitle":"","Time":"2013-11-28T16:13:00Z","Authors":[{"Elem":[{"Lines":["Author"],"Pre":false}]}],"Sections":[{"Number":[1],"Title":"Title","Elem":[{"Bullet":["a point","second point","third point"]},{"Lines":["Everything I know."],"Pre":false}]}]}`,
}

var boldError = []string{
	`Head
16:13 28 Nov 2013
Author
# Title

hello**bb**hi https://qortex.cn
`,

	`{"Title":"Head","Subtitle":"","Time":"2013-11-28T16:13:00Z","Authors":[{"Elem":[{"Lines":["Author"],"Pre":false}]}],"Sections":[{"Number":[1],"Title":"Title","Elem":[{"Lines":["hello *bb* hi [[https://qortex.cn][https://qortex.cn]]"],"Pre":false}]}]}`,
}

var fixtures = [][]string{
	imageTextImage,
	chineseTitleAndList,
	spaceBetweenListItem,
	noSpaceBetweenListItem,
	boldError,
}

func TestParseSeperateListItems(t *testing.T) {
	for _, fixture := range fixtures {
		result := fixture[1]
		r := strings.NewReader(fixture[0])
		doc, err := Parse(r, "test", 0)
		if err != nil {
			panic(err)
		}
		actual, _ := json.Marshal(doc)
		if string(actual) != result {
			t.Error("Do not generate expected result actual: \n", string(actual))
		}
	}

}
