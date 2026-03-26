package sqldbs

import (
	"encoding/json/jsontext"
	"io/fs"
)

type Client interface {
	CreateDB(name string, conf jsontext.Value) error
	DB(name string) (DB, bool)
	RawSQLStore(name string) *RawSQLStore
	LoadRawSQL(name string, sqlFS fs.FS) error
	Close() error
}
