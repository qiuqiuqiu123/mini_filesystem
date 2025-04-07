package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"mini_filesystem/common"
	"mini_filesystem/logger"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var (
	infoLogger  = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)
	warnLogger  = log.New(os.Stdout, "WARN: ", log.LstdFlags|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile)
)

type StorageService struct {
	lock     sync.Mutex
	current  *FileStorage
	files    map[int64]*FileStorage
	rootPath string
	seqIdx   int64
}

func NewStorageService(rootPath string) *StorageService {

	filePath := rootPath + "/storage_path"
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal("create file storage err:", err)
	}

	fileStorage := &FileStorage{
		lock: sync.Mutex{},
		off:  0,
		fd:   file,
		id:   0,
	}

	files := make(map[int64]*FileStorage, 1)
	files[0] = fileStorage

	return &StorageService{
		rootPath: rootPath,
		files:    files,
		seqIdx:   0,
		lock:     sync.Mutex{},
		current:  fileStorage,
	}
	return &StorageService{}
}

func (ss *StorageService) Init() {

}

func (ss *StorageService) ObjectWrite(w http.ResponseWriter, r *http.Request) {
	vals := mux.Vars(r)
	idStr := vals["id"]
	SizeStr := vals["size"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	size, err := strconv.ParseInt(SizeStr, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	logger.GetLogger().Infof("write data id: %v size: %v", id, size)
	// 获取文件的写句柄
	ss.lock.Lock()
	current := ss.current
	ss.lock.Unlock()

	// 写数据
	loc, err := current.Write(id, r.Body, size)
	if err != nil {
		logger.GetLogger().Errorf("Error writing file: %v", err)
		w.WriteHeader(500)
		return
	}
	logger.GetLogger().Infof("write success, loc=%v", loc)

	// 返回写入数据位置信息
	data, err := json.Marshal(loc)
	if err != nil {
		logger.GetLogger().Errorf("json marshal err: %v", err)
	}
	_, err = w.Write(data)
	if err != nil {
		logger.GetLogger().Errorf("return loc err: %v", err)
	}
	logger.GetLogger().Info("return file loc success")

}

func (ss *StorageService) ObjectRead(w http.ResponseWriter, r *http.Request) {
	vals := mux.Vars(r)

	fidStr := vals["fid"]
	offStr := vals["off"]
	lenStr := vals["size"]
	crcStr := vals["crc"]
	logger.GetLogger().Infof("read data fid: %v,off: %v,size: %v,crc: %v", fidStr, offStr, lenStr, crcStr)
	fileID, err := strconv.ParseInt(fidStr, 10, 64)
	if err != nil {
		logger.GetLogger().Errorf("parse file id err: %v", err)
	}
	offset, err := strconv.ParseInt(offStr, 10, 64)
	if err != nil {
		logger.GetLogger().Errorf("parse offset err: %v", err)
	}
	length, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		logger.GetLogger().Errorf("parse length err: %v", err)
	}
	crcSum, err := strconv.ParseUint(crcStr, 10, 32)
	if err != nil {
		logger.GetLogger().Errorf("parse crc err: %v", err)
	}

	loc := &common.Location{
		FileID: uint64(fileID),
		Offset: offset,
		Length: length,
		Crc:    uint32(crcSum),
	}

	// 获取文件句柄
	ss.lock.Lock()
	stor, exist := ss.files[fileID]
	ss.lock.Unlock()

	if !exist {
		logger.GetLogger().Errorf("file %v not exist", fileID)
		w.WriteHeader(400)
		return
	}
	// 读取数据，并返回
	data, err := stor.Read(loc)
	if err != nil {
		logger.GetLogger().Errorf("read data err: %v", err)
	}
	w.Write(data)
	logger.GetLogger().Infof("read success, len(data)=%d", len(data))
}
