package common

import (
	"context"
	"io"
)

type Location struct {
	FileID uint64
	Offset int64
	Length int64
	Crc    uint32
}

type ReadArgs struct {
	Loc Location
}

type WriteArgs struct {
	ID   uint64
	Size int64
	Body io.Reader
}

type WriteResp struct {
	Loc Location
}

type StorageApi interface {
	Read(ctx context.Context, host string, args *ReadArgs) (reader io.Reader, err error)

	Write(ctx context.Context, host string, args *WriteArgs) (loc *Location, err error)
}
