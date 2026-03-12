package storages

import (
	"context"
	"encoding/json/jsontext"
	"io"
)

type Storage interface {
	Exists(ctx context.Context, path string) (bool, error)
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	Put(ctx context.Context, path string, r io.Reader) error
	Delete(ctx context.Context, path string) error
	Size(ctx context.Context, path string) (int64, error)
	Copy(ctx context.Context, src string, dst string) error
	Move(ctx context.Context, src string, dst string) error
}

type Client interface {
	CreateStorage(name string, conf jsontext.Value) error
	Storage(name string) (Storage, bool)
}
