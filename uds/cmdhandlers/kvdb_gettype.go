package cmdhandlers

import (
	"fmt"
	"io"

	"github.com/x64c/gw/framework"
	"github.com/x64c/gw/kvdbs"
)

type KvdbGetType struct {
	AppProvider framework.AppProviderFunc
	KVDB        kvdbs.DB
}

func (h *KvdbGetType) GroupName() string {
	return "kvdb"
}

func (h *KvdbGetType) Command() string {
	return "kvdb-get-type"
}

func (h *KvdbGetType) Desc() string {
	return "Print the type of the given key in KV database"
}

func (h *KvdbGetType) Usage() string {
	return h.Command() + " key"
}

func (h *KvdbGetType) HandleCommand(args []string, w io.Writer) error {
	argLen := len(args)
	if argLen != 1 {
		return fmt.Errorf("usage: %s", h.Usage())
	}
	key := args[0]
	ctx := h.AppProvider().AppCore().RootCtx
	found, err := h.KVDB.Exists(ctx, key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("key not found")
	}
	typeName, err := h.KVDB.Type(ctx, key)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w, typeName)
	return nil
}
