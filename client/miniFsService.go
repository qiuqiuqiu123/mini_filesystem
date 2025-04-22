package client

import (
	"bazil.org/fuse/fs"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"mini_filesystem/common"
	"sync"
	"time"
)

const ServerGroupTable = "server_group"
const SuperBlockTable = "super_block"

type MiniFsService struct {
	MetaCli      *mongo.Database
	replication  []*ServerGroup
	inoAllocator *InodeAllocator
}

func (fs *MiniFsService) AllocInode() (uint64, error) {
	// 分配完成之后，需要落库,避免下次启动的重复分配了
	inode, err := fs.inoAllocator.AllocInode()
	if err != nil {
		return 0, err
	}
	// 修改落库
	coll := fs.MetaCli.Collection(SuperBlockTable)
	filter := bson.M{"name": "minifs"}
	update := bson.M{"$set": bson.M{"inode_num_base": inode}}
	var block common.SuperBlock
	err = coll.FindOneAndUpdate(context.TODO(), filter, update, options.FindOneAndUpdate()).Decode(&block)
	if err != nil {
		return 0, err
	}
	return inode, nil
}

func (fs *MiniFsService) PickWriteServerGroup() *ServerGroup {
	return fs.replication[1]
}

func (fs *MiniFsService) PickWriteServerGroupById(groupId int) (*ServerGroup, error) {
	return fs.replication[groupId], nil
}

func (fs *MiniFsService) Init() {
	log.Println("miniFs init start.....")
	fs.initMongo()
	fs.initServerGroup()
	fs.initInoAllocator()
}

func (fs *MiniFsService) initMongo() {
	log.Println("miniFs init mongo start.....")
	// 1. 设置 MongoDB 连接 URI

	uri := "mongodb://admin:admin@localhost:27017"
	// 2. 配置客户端选项
	clientOptions := options.Client().ApplyURI(uri)
	// 3. 创建客户端
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	// 5. 打印连接成功信息
	fmt.Println("Connected to MongoDB!")
	fs.MetaCli = client.Database("mydatabase")
}

func (fs *MiniFsService) initServerGroup() {
	log.Println("miniFs server group start.....")
	coll := fs.MetaCli.Collection(ServerGroupTable)

	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var results []ServerGroup
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	fs.replication = make([]*ServerGroup, 3)
	for idx, result := range results {
		fs.replication[result.GroupId] = &results[idx]
	}
}

func (fs *MiniFsService) initInoAllocator() {
	log.Println("miniFs inoAllocator start.....")
	coll := fs.MetaCli.Collection(SuperBlockTable)
	cursor, err := coll.Find(context.TODO(), bson.M{
		"name": "minifs",
	})
	if err != nil {
		log.Fatal(err)
	}

	var result []common.SuperBlock
	if err = cursor.All(context.TODO(), &result); err != nil {
		log.Fatal(err)
	}
	fs.inoAllocator = &InodeAllocator{
		lock:       sync.Mutex{},
		superBlock: &result[0],
		metaCli:    fs.MetaCli,
	}
}

func (fs *MiniFsService) Root() (fs.Node, error) {
	return &Dir{
		fs:    fs,
		inode: 1,
		files: make(map[string]*File),
	}, nil
}
