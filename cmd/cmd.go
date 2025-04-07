package main

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"flag"
	"log"
	"mini_filesystem/client"
)

func main() {
	var mountpoint string

	flag.StringVar(&mountpoint, "mountpoint", "", "mount point(dir)?")
	flag.Parse()

	if mountpoint == "" {
		log.Fatal("mountpoint is required")
	}

	// 创建一个解析和封装Fuse请求监听通道对象
	c, err := fuse.Mount(mountpoint, fuse.FSName("mini_filesystem"), fuse.Subtype("minifs"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// 注册FS到server,以便可以处理回调请求
	minifs := &client.MiniFsService{}
	minifs.Init()
	log.Println("register fs")
	err = fs.Serve(c, minifs)
	if err != nil {
		log.Fatal(err)
	}
}
