package client

import (
	"bazil.org/fuse"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"mini_filesystem/common"
	"os"
)

type File struct {
	fs           *MiniFsService
	fileName     string
	metaInfo     *common.InodeInfo
	dataLocation *common.FileDataLocation
}

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("file#attr exe...")
	attr.Inode = f.metaInfo.INode
	attr.Mode = os.FileMode(f.metaInfo.Mode)
	attr.Size = f.metaInfo.Size
	attr.Atime = f.metaInfo.AccessTime
	attr.Ctime = f.metaInfo.CreateTime
	attr.Mtime = f.metaInfo.ModTime
	attr.Gid = f.metaInfo.Gid
	attr.Uid = f.metaInfo.Uid
	return nil
}

func NewFile(fs *MiniFsService, fileName string, metaInfo *common.InodeInfo) *File {
	return &File{
		fs:           fs,
		fileName:     fileName,
		metaInfo:     metaInfo,
		dataLocation: &common.FileDataLocation{},
	}
}

func (f *File) writeMeta(ctx context.Context, dataLoc *common.FileDataLocation, size int) error {
	coll := f.fs.MetaCli.Collection(FileMetaInfoTable)
	filter := bson.M{"inode.inode": f.metaInfo.INode}
	update := bson.M{"$set": bson.M{"data_location": dataLoc, "inode.Size": size}}
	var m common.FileMetaInfo
	if err := coll.FindOneAndUpdate(ctx, filter, update).Decode(&m); err != nil {
		return err
	}
	return nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("file#write exe...")
	// 选择一个复制组
	repSet := f.fs.PickWriteServerGroup()
	// 将数据写入复制组
	locs, err := repSet.Write(ctx, req.Data, int64(f.metaInfo.INode))
	if err != nil {
		return err
	}
	dataLoc := &common.FileDataLocation{
		GroupId:   repSet.GroupId,
		Locations: locs,
	}
	// 将文件的元数据持久化到元数据中心
	err = f.writeMeta(ctx, dataLoc, len(req.Data))
	if err != nil {
		return err
	}
	resp.Size = len(req.Data)
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("file#readAll exe...")
	// 判断是否需要加载文件的位置信息
	if f.dataLocation == nil {
		if err := f.refreshMeta(); err != nil {
			return nil, err
		}
	}
	// 从元数据中找到复制组
	repSet, err := f.fs.PickWriteServerGroupById(f.dataLocation.GroupId)
	if err != nil {
		return nil, err
	}
	// 从复制组中读取用户数据
	content, err := repSet.Read(ctx, f.dataLocation)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (f *File) refreshMeta() error {
	return nil
}
