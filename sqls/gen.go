package sqls

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"net/url"
	"strconv"
	"strings"
)

type info struct {
	Field   string         `db:"Field"`
	Type    string         `db:"Type"`
	Null    string         `db:"Null"`
	Key     sql.NullString `db:"Key"`
	Default sql.NullString `db:"Default"`
	Extra   sql.NullString `db:"Extra"`
}

func Gen(source string) {
	db, err := sqlx.Open("mysql", source)
	if err != nil {
		panic(err)
	}

	var tables []string
	err = db.Select(&tables, `show tables`)
	if err != nil {
		panic(err)
	}

	for _, tablename := range tables {
		var lst []*info
		err = db.Select(&lst, `desc `+tablename)
		if err != nil {
			panic(err)
		}
		fmt.Println(gen(tablename, lst))
	}
}

func named(n string) string {
	bytesV := []byte(n)
	if bytesV[0] >= 'a' && bytesV[0] <= 'z' {
		bytesV[0] = bytesV[0] - uint8(int('a')-int('A'))
	}
	return string(bytesV)
}

func gen(name string, lst []*info) string {
	buf := strings.Builder{}

	buf.WriteString(fmt.Sprintf("type %s struct{\n", named(name)))

	for _, item := range lst {
		buf.WriteString("\t")
		buf.WriteString(named(item.Field))
		buf.WriteString(" ")

		stype := ""
		unsigned := false
		notnull := false
		primary := false
		unique := false
		incr := false
		defaultV := sql.NullString{}
		length := -2
		isStringType := false

		if ind := strings.IndexByte(item.Type, ' '); ind > 1 {
			stype = item.Type[:ind]
			unsigned = true
		} else if ind := strings.IndexByte(item.Type, '('); ind > 1 {
			stype = item.Type[:ind]
			v, err := strconv.ParseUint(item.Type[ind+1:len(item.Type)-1], 10, 32)
			if err != nil {
				panic(err)
			}
			length = int(v)
		} else {
			stype = item.Type
		}

		if item.Null == "YES" {
			notnull = true
		}

		if item.Key.Valid {
			switch item.Key.String {
			case "PRI":
				primary = true
			case "UNI":
				unique = true
			}
		}

		if item.Default.Valid {
			defaultV.Valid = true
			defaultV.String = item.Default.String
		}

		if item.Extra.Valid {
			if strings.Index(item.Extra.String, "auto_increment") > -1 {
				incr = true
			}
		}

		switch stype {
		case "bool":
			buf.WriteString(" bool")
		case "tinyint":
			buf.WriteString(" int8")
		case "smallint":
			buf.WriteString(" int16")
		case "int":
			if unsigned {
				buf.WriteString(" uint32")
			} else {
				buf.WriteString(" int32")
			}
		case "bigint":
			if unsigned {
				buf.WriteString(" uint64")
			} else {
				buf.WriteString(" int64")
			}
		case "char", "varchar":
			isStringType = true
			buf.WriteString(" string")
		case "blob":
			buf.WriteString(" []byte")
			length = -1
		case "text":
			buf.WriteString(" string")
			length = 0
		}

		buf.WriteString(" `ddl:\"")
		if notnull {
			buf.WriteString("notnull;")
		}
		if primary {
			buf.WriteString("primary;")
		}
		if unique {
			buf.WriteString("unique;")
		}
		if incr {
			buf.WriteString("incr;")
		}

		if length > -2 {
			buf.WriteString(fmt.Sprintf("L<%d>;", length))
		}

		if defaultV.Valid {
			if !isStringType {
				buf.WriteString(fmt.Sprintf("D<%s>;", url.QueryEscape(defaultV.String)))
			} else {
				buf.WriteString(fmt.Sprintf("D<'%s'>;", url.QueryEscape(defaultV.String)))
			}
		}

		buf.WriteString("\"`\n")
	}
	buf.WriteString("}")

	return buf.String()
}
