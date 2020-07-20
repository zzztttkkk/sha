package sqls

import (
	"github.com/jmoiron/sqlx"

	"github.com/zzztttkkk/suna/ini"
)

var config *ini.Config
var leader *sqlx.DB
var followers []*sqlx.DB

func Init(conf *ini.Config, leaderV *sqlx.DB, followersV []*sqlx.DB) {
	config = conf
	leader = leaderV
	followers = followersV
}
