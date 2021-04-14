package jsonx

import (
	"fmt"
	"testing"
)

func TestPeek(t *testing.T) {
	obj, _ := NewObject(
		[]byte(`{
  "status": "success",
  "data": {
    "id": "5b8b9572e116fb3714e72b05",
    "content": "斜风细雨作春寒，对尊前，忆前欢。",
    "popularity": 49500,
    "origin": {
      "title": "江城子·赏春",
      "dynasty": "宋代",
      "author": "朱淑真",
      "content": [
        "斜风细雨作春寒，对尊前，忆前欢。曾把梨花，寂寞泪阑干。芳草断烟南浦路，和别泪，看青山。",
        "昨宵结得梦夤缘，水云间，悄无言。争奈醒来，愁恨又依然。展转衾裯空懊恼，天易见，见伊难。"
      ],
      "translate": null
    },
    "matchTags": [
      "小雨",
      "春",
      "雨",
      "寒冷",
      "风"
    ],
    "recommendedReason": "",
    "cacheAt": "2021-03-17T11:03:16.943834"
  },
  "token": "H2+o0ggox1x+KTl4hhGkeLNunPZSJ4SS",
  "ipAddress": "101.71.197.89",
  "warning": null
}`),
	)

	fmt.Println(obj)
	fmt.Println(obj.PeekIntDefault(-1, "data", "popularity"))
	fmt.Println(obj.PeekStringDefault("", "data", "content"))
	fmt.Println(obj.PeekStringDefault("", "data", "origin", "content", "1"))
	fmt.Println(obj.IsNull("warning"))
	fmt.Println(obj.IsNull("data", "origin", "translate"))
	fmt.Println(obj.PeekTimeFromString("2006-01-02T15:04:05.000000", "data", "cacheAt"))
}

func TestUnmarshal(t *testing.T) {
	type A struct {
		S string
	}

	var a A
	err := Unmarshal([]byte(`{"S":"<div></div>"}`), &a)
	fmt.Println(a.S, err)
}
