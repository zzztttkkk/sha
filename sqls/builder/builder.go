package builder

import (
	"github.com/zzztttkkk/sqrl"
)

func NewSelect(cols ...string) *sqrl.SelectBuilder {
	if isPostgres {
		return sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar).Select(cols...)
	}
	return sqrl.Select(cols...)
}

func NewUpdate(table string) *sqrl.UpdateBuilder {
	if isPostgres {
		return sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar).Update(table)
	}
	return sqrl.Update(table)
}

func NewDelete(what ...string) *sqrl.DeleteBuilder {
	if isPostgres {
		return sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar).Delete(what...)
	}
	return sqrl.Delete(what...)
}

func NewInsert(table string) *sqrl.InsertBuilder {
	if isPostgres {
		return sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar).Insert(table)
	}
	return sqrl.Insert(table)
}
