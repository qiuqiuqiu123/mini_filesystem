package client

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"mini_filesystem/common"
	"os"
	"syscall"
	"time"
)

const FileMetaInfoTable = "file_meta_info"

type Dir struct {
	fs    *MiniFsService
	inode uint64
	files map[string]*File
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("dir#create exe...")
	newIno, err := d.fs.inoAllocator.AllocInode()
	if err != nil {
		return nil, nil, fmt.Errorf("error allocating new inode: %v", err)
	}

	meta := &common.FileMetaInfo{
		Name: req.Name,
		Inode: common.InodeInfo{
			INode:      newIno,
			Mode:       uint32(req.Mode),
			Size:       0,
			Uid:        req.Uid,
			Gid:        req.Gid,
			ModTime:    time.Time{},
			CreateTime: time.Time{},
			AccessTime: time.Time{},
		},
	}

	coll := d.fs.MetaCli.Collection(FileMetaInfoTable)
	_, err = coll.InsertOne(context.TODO(), meta)

	if err != nil {
		return nil, nil, fmt.Errorf("creat dir error: %v", err)
	}

	f := NewFile(d.fs, req.Name, meta)
	d.files[req.Name] = f
	return f, f, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("dir#readDirAll exe...")
	var dirDirs = []fuse.Dirent{}
	// 从元数据中心获取文件列表
	coll := d.fs.MetaCli.Collection(FileMetaInfoTable)
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	var results []common.FileMetaInfo
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		//err := cursor.Decode(&result)
		//if err != nil {
		//	return nil, err
		//}
		dirent := fuse.Dirent{Inode: result.Inode.INode, Name: result.Name, Type: fuse.DT_File}
		dirDirs = append(dirDirs, dirent)
	}
	return dirDirs, nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("dir#Lookup exe...")
	coll := d.fs.MetaCli.Collection(FileMetaInfoTable)
	var meta common.FileMetaInfo
	filter := bson.M{"name": name}
	err := coll.FindOne(context.TODO(), filter).Decode(&meta)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, syscall.ENOENT
		}
	}

	f := NewFile(d.fs, name, &meta)
	return f, nil
}

func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("dir#attr exe...")
	attr.Inode = d.inode
	attr.Mode = os.ModeDir | 0444
	return nil
}
