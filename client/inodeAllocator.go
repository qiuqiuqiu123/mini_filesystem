package client

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"mini_filesystem/common"
	"mini_filesystem/utils"
	"sync"
)

type SnowFlakeInodeAllocator struct {
	idGen *utils.Snowflake
}

func (inoLoc *SnowFlakeInodeAllocator) AllocInode() (uint64, error) {
	// 雪花算法分配一个
	return inoLoc.idGen.Generate()
}

type InodeAllocator struct {
	lock       sync.Mutex
	superBlock *common.SuperBlock
	metaCli    *mongo.Database
}

func (inoLoc *InodeAllocator) AllocInode() (uint64, error) {
	inoLoc.lock.Lock()
	defer inoLoc.lock.Unlock()
	inode := inoLoc.superBlock.InodeNumBase + 1
	inoLoc.superBlock.InodeNumBase = inode
	err := inoLoc.save()
	if err != nil {
		return 0, err
	}
	return inode, nil
}

func (inoLoc *InodeAllocator) save() error {
	coll := inoLoc.metaCli.Collection(SuperBlockTable)
	filer := bson.M{"name": inoLoc.superBlock.Name}
	update := bson.M{"$set": bson.M{"inode_num_base": inoLoc.superBlock.InodeNumBase}}
	_, err := coll.UpdateOne(context.Background(), filer, update)
	if err != nil {
		return err
	}
	return nil
}
