package client

import (
	"bytes"
	"context"
	"io"
	"log"
	"mini_filesystem/common"
	"sync"
)

// 复制组,采用的策略为WARO，写全部，任意读一个
type ServerGroup struct {
	GroupId int      `bson:"group_id"`
	Servers []string `bson:"servers"`
}

func (sg *ServerGroup) Write(ctx context.Context, data []byte, id int64) (retLocs []*common.Location, retErr error) {
	cli := NewClient()

	Servers := sg.Servers
	wg := sync.WaitGroup{}
	// 所有的副本都需要写入成功,并发写入
	reps := make([]Err, len(Servers))
	for idx, serverAddr := range Servers {
		wg.Add(1)
		args := &common.WriteArgs{
			ID:   uint64(id),
			Size: int64(len(data)),
			Body: bytes.NewReader(data),
		}

		go func(index int, addr string) {
			defer wg.Done()
			loc, err := cli.Write(ctx, addr, args)
			reps[index] = Err{
				idx: index,
				loc: loc,
				err: err,
			}
		}(idx, serverAddr)
	}

	wg.Wait()
	// 校验是不是所有副本都写入成功了
	for _, ret := range reps {
		if ret.err != nil {
			return nil, ret.err
		}
		retLocs = append(retLocs, ret.loc)
	}
	log.Printf("write success: %v", sg.Servers)
	return retLocs, nil
}

func (sg *ServerGroup) Read(ctx context.Context, dataLoc *common.FileDataLocation) (content []byte, retErr error) {
	// 读取的话，随意从一个副本读取出来就好
	cli := NewClient()

	pickIdx := sg.pickServerForRead()
	server := sg.Servers[pickIdx]
	args := &common.ReadArgs{Loc: *dataLoc.Locations[pickIdx]}
	reader, err := cli.Read(ctx, server, args)
	if err != nil {
		return nil, err
	}
	content, err = io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (sg *ServerGroup) pickServerForRead() int {
	// 默认读第一个
	return 1
}
