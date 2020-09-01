package builder

import (
	"github.com/zzztttkkk/sqlr"
)

func NewSelect(cols ...string) *sqlr.SelectBuilder {
	if isPostgres {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Select(cols...)
	}
	return sqlr.Select(cols...)
}

func NewUpdate(table string) *sqlr.UpdateBuilder {
	if isPostgres {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Update(table)
	}
	return sqlr.Update(table)
}

func NewDelete(what ...string) *sqlr.DeleteBuilder {
	if isPostgres {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Delete(what...)
	}
	return sqlr.Delete(what...)
}

func NewInsert(table string) *sqlr.InsertBuilder {
	if isPostgres {
		return sqlr.StatementBuilder.PlaceholderFormat(sqlr.Dollar).Insert(table)
	}
	return sqlr.Insert(table)
}
