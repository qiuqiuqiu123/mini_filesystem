package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	timestampBits  = 41 // 时间戳位数
	dataCenterBits = 5  // 数据中心ID位数
	workerBits     = 5  // 节点ID位数
	sequenceBits   = 12 // 序列号位数

	// 最大值计算（位移运算）
	maxDataCenterID = -1 ^ (-1 << dataCenterBits) // 31
	maxWorkerID     = -1 ^ (-1 << workerBits)     // 31
	maxSequence     = -1 ^ (-1 << sequenceBits)   // 4095

	// 位移量
	workerShift     = sequenceBits                               // 12
	dataCenterShift = sequenceBits + workerBits                  // 17
	timestampShift  = sequenceBits + workerBits + dataCenterBits // 22
)

type Snowflake struct {
	mu           sync.Mutex
	epoch        int64 // 起始时间（单位：毫秒）
	timestamp    int64 // 上次生成ID的时间戳
	dataCenterID int64 // 数据中心ID
	workerID     int64 // 节点ID
	sequence     int64 // 序列号
}

func NewSnowflake(epoch time.Time, dataCenterID, workerID int64) (*Snowflake, error) {
	if dataCenterID < 0 || dataCenterID > maxDataCenterID {
		return nil, fmt.Errorf("dataCenterID超出范围（0 ≤ id ≤ %d）", maxDataCenterID)
	}
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("workerID超出范围（0 ≤ id ≤ %d）", maxWorkerID)
	}
	return &Snowflake{
		epoch:        epoch.UnixNano() / 1e6, // 转为毫秒
		dataCenterID: dataCenterID,
		workerID:     workerID,
	}, nil
}

func (s *Snowflake) Generate() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano() / 1e6 // 当前毫秒时间戳
	if now < s.timestamp {
		return 0, fmt.Errorf("时钟回拨，拒绝生成ID（回拨时间：%dms）", s.timestamp-now)
	}

	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 { // 序列号溢出，等待下一毫秒
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		s.sequence = 0
	}

	s.timestamp = now
	elapsedTime := now - s.epoch

	// 组合各字段生成ID
	id := (elapsedTime << timestampShift) |
		(s.dataCenterID << dataCenterShift) |
		(s.workerID << workerShift) |
		s.sequence
	return uint64(id), nil
}

func main() {
	epoch, _ := time.Parse("2006-01-02", "2024-01-01")
	sf, _ := NewSnowflake(epoch, 1, 1) // 数据中心1，节点1

	for i := 0; i < 5; i++ {
		id, _ := sf.Generate()
		fmt.Printf("生成的ID: %d\n", id)
	}
}
