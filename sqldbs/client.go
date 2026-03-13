package sqldbs

import "encoding/json/jsontext"

type Client interface {
	CreateDB(name string, conf jsontext.Value) error
	DB(name string) (DB, bool)
	RawSQLStore() *RawSQLStore
	Close() error
}
