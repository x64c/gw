package sqldbs

import "github.com/x64c/gw/model"

type Deletable[T any, ID comparable] interface {
	Tabular[T]
	model.Identifiable[ID]
}
