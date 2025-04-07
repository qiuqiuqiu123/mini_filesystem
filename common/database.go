package common

import (
	"time"
)

type SuperBlock struct {
	Name         string `bson:"name"`
	InodeNumBase uint64 `bson:"inode_num_base"`
	InodeStep    uint64 `bson:"inode_step"`
}

type FileMetaInfo struct {
	Name     string            `bson:"name"`
	Inode    InodeInfo         `bson:"inode"`
	DataLocs *FileDataLocation `bson:"data_location"`
}

// 文件系统核心
type InodeInfo struct {
	INode      uint64    `bson:"inode"`
	Mode       uint32    `bson:"mode"`
	Size       uint64    `bson:"size"`
	Uid        uint32    `bson:"uid"`
	Gid        uint32    `bson:"gid"`
	Generation uint64    `bson:"generation"`
	ModTime    time.Time `bson:"mtime"`
	CreateTime time.Time `bson:"ctime"`
	AccessTime time.Time `bson:"atime"`
}

type FileDataLocation struct {
	GroupId   int         `bson:"group_id"`
	Locations []*Location `bson:"locations"`
}
