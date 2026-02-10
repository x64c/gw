package mysql

import (
	"github.com/x64c/gw/db/sqldb"
)

func Register() {
	sqldb.RegisterFactory(DBType, NewClient)
}
