package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mini_filesystem/common"
	"net/http"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

// 网络读请求
func (c *Client) Read(ctx context.Context, host string, args *common.ReadArgs) (reader io.Reader, err error) {
	urlStr := fmt.Sprintf("%v/object/read/fid/%d/off/%d/size/%d/crc/%d", host, args.Loc.FileID, args.Loc.Offset, args.Loc.Length, args.Loc.Crc)
	req, err := http.NewRequest(http.MethodPost, urlStr, nil)
	if err != nil {
		return
	}
	client := &http.Client{}
	rep, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return rep.Body, nil
}

// 网络写请求
func (c *Client) Write(ctx context.Context, host string, args *common.WriteArgs) (loc *common.Location, err error) {
	urlStr := fmt.Sprintf("%v/object/write/id/%d/size/%d", host, args.ID, args.Size)
	req, err := http.NewRequest(http.MethodPost, urlStr, args.Body)
	if err != nil {
		return
	}
	req.ContentLength = args.Size
	client := &http.Client{}
	rep, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer rep.Body.Close()
	resData, err := io.ReadAll(rep.Body)
	if err != nil {
		log.Fatal(err)
	}
	loc = &common.Location{}
	err = json.Unmarshal(resData, loc)
	if err != nil {
		log.Fatal(err)
	}
	return loc, nil

}
