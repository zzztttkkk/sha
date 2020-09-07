package sqlx

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// auto create table monthly.
type MonthlyOperator struct {
	Operator
	end  time.Time
	lock sync.Mutex
}

func (op *MonthlyOperator) TableName() string {
	op.lock.Lock()
	defer op.lock.Unlock()

	now := time.Now()
	y, m, _ := now.Date()
	name := fmt.Sprintf("%s_%04d_%02d", op.Operator.TableName(), y, m)
	if now.After(op.end) {
		CreateTable(op._GetDbLeader(), op.ele, name)

		var end = now
		for {
			end = end.AddDate(0, 0, 1)
			if end.Month() != m {
				break
			}
		}

		end = end.AddDate(0, 0, -1)
		op.end = end
	}
	return name
}

func (op *MonthlyOperator) Init(ele reflect.Value) {
	op.ele = ele
	CreateTable(op._GetDbLeader(), ele, op.TableName())
}
