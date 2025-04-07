package server

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"mini_filesystem/common"
	"mini_filesystem/logger"
	"os"
	"sync"
)

type FileStorage struct {
	lock sync.Mutex
	off  int64
	fd   *os.File
	id   int64
}

const headerSize = 20
const footerSize = 4

// 自制文件系统魔数
const magicNum uint32 = 6719

// 写入文件时除了写入原本的数据，还会添加一个固定长度的藏头和藏尾
// 藏头记录了用户数据的元数据、包括魔数、文件ID和数据长度等
// 藏尾记录了CRC校验码，用于数据紧急恢复的场景
func (fs *FileStorage) Write(id int64, reader io.Reader, size int64) (loc *common.Location, err error) {
	fs.lock.Lock()
	startPos, pos := fs.off, fs.off
	fs.off += size + int64(headerSize) + int64(footerSize)
	fs.lock.Unlock()
	crc := crc32.NewIEEE()
	reader = io.LimitReader(reader, size)
	reader = io.TeeReader(reader, crc)
	header := make([]byte, headerSize)
	footer := make([]byte, footerSize)
	magic := make([]byte, 4)
	binary.LittleEndian.PutUint32(magic, magicNum)
	copy(header[:4], magic[:])
	binary.BigEndian.PutUint64(header[4:], uint64(id))
	binary.BigEndian.PutUint64(header[4+8:], uint64(size))
	logger.GetLogger().Debugf("build header: %v", header)

	// 写入藏头
	_, err = fs.fd.WriteAt(header, pos)
	if err != nil {
		return nil, err
	}
	logger.GetLogger().Info("write header success")

	pos += int64(headerSize)
	// 写入数据本身
	writer := &common.Writer{WriterAt: fs.fd, Offset: pos}
	n, err := io.Copy(writer, reader)
	if err != nil {
		log.Fatal(err)
	}
	logger.GetLogger().Info("write body success")

	pos += n
	crc32Sum := crc.Sum32()
	binary.BigEndian.PutUint32(footer, crc32Sum)
	// 写入藏尾
	logger.GetLogger().Debugf("build footer: %v", footer)
	fs.fd.WriteAt(footer, pos)
	loc = &common.Location{
		FileID: uint64(fs.id),
		Offset: startPos,
		Length: size,
		Crc:    crc32Sum,
	}
	logger.GetLogger().Info("write footer success")
	return loc, nil
}

func (fs *FileStorage) Read(loc *common.Location) ([]byte, error) {
	header := make([]byte, headerSize)
	footer := make([]byte, footerSize)
	secReader := io.NewSectionReader(fs.fd, loc.Offset, loc.Length+int64(headerSize)+int64(footerSize))
	// 读取头部
	_, err := secReader.Read(header)

	if err != nil {
		logger.GetLogger().Errorf("read header failed: %v", err)
	}
	// 校验头部
	_maigc := binary.LittleEndian.Uint32(header[:4])
	if _maigc != magicNum {
		return nil, fmt.Errorf("invalid file magic number: %d", _maigc)
	}
	id := binary.BigEndian.Uint64(header[4:])
	if id != loc.FileID {
		return nil, fmt.Errorf("invalid file id: %d", id)
	}
	size := binary.BigEndian.Uint64(header[4+8:])
	if size != uint64(loc.Length) {
		return nil, fmt.Errorf("invalid file size: %d", size)
	}
	// 读取数据
	crc := crc32.NewIEEE()
	reader := io.LimitReader(secReader, int64(loc.Length))
	reader = io.TeeReader(reader, crc)
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	// 读取尾部
	_, err = secReader.Seek(headerSize+loc.Length, io.SeekStart)
	if err != nil {
		logger.GetLogger().Errorf("seek footer failed: %v", err)
	}
	_, err = secReader.Read(footer)
	if err != nil {
		logger.GetLogger().Errorf("read footer failed: %v", err)
	}
	_crcSum := binary.BigEndian.Uint32(footer)
	// 校验CRC
	crcSum := crc.Sum32()
	if _crcSum != crcSum {
		logger.GetLogger().Error("crc not match")
	}
	return data, nil
}
