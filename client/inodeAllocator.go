package client

import (
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
}

func (inoLoc *InodeAllocator) AllocInode() (uint64, error) {
	inoLoc.lock.Lock()
	defer inoLoc.lock.Unlock()
	inode := inoLoc.superBlock.InodeNumBase + 1
	inoLoc.superBlock.InodeNumBase = inode
	return inode, nil
}
