package storages

import "github.com/x64c/gw/storages/keystores"

type Conf struct {
	KeyStoreConf keystores.Conf `json:"key_store"`
}
