package captcha

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestNewRandTokenGenerator(t *testing.T) {
	cg := NewRandTokenGenerator(
		func(ctx context.Context) string {
			v := make([]byte, 10)
			utils.RandBytes(v, utils.Base58BytesPool)
			return string(v)
		},
		"*",
		&Options{
			OffsetX: 5,
			OffsetY: 5,
			Points:  100,
		},
	)

	of, _ := os.OpenFile("rtg.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = cg.GenerateTo(context.Background(), of)
}

func TestNewShuffleStringGenerator(t *testing.T) {
	cg := NewShuffleStringGenerator(
		func(ctx context.Context) []string {
			res, _ := http.Get("https://v2.jinrishici.com/one.json")
			defer res.Body.Close()
			buf, _ := ioutil.ReadAll(res.Body)
			data, _ := jsonx.NewObject(buf)
			verse := data.PeekStringDefault("", "data", "origin", "content", jsonx.SliceRand) // choice a verse randomly

			// keep the structure
			var splitVerse = func(v string) []string {
				fmt.Println(verse)

				for _, sep := range strings.Split("。？！——；", "") {
					v = strings.Replace(v, sep, ",", -1)
				}
				v = v[:len(v)-1]
				return strings.Split(v, "，")
			}

			return splitVerse(verse)
		},
		"*",
		&ShuffleOptions{
			Options: Options{
				OffsetX: 5,
				OffsetY: 5,
				Points:  200,
			},
			PCT: 30,
			Sep: []rune("，"),
		},
	)

	of, _ := os.OpenFile("ssg.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = cg.GenerateTo(context.Background(), of)
}
