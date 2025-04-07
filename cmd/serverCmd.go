package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"mini_filesystem/logger"
	"mini_filesystem/server"
	"net/http"
)

func main() {
	// 检查命令行参数
	// 默认/home/sti/my_filesystem/node/1
	logger.Init(logger.LevelInfo)
	startServer("/app/data", 9095)
}

func startServer(dataPath string, port int64) {

	s := server.NewStorageService(dataPath)
	s.Init()

	router := mux.NewRouter()

	router.HandleFunc("/object/write/id/{id}/size/{size}", s.ObjectWrite)
	router.HandleFunc("/object/read/fid/{fid}/off/{off}/size/{size}/crc/{crc}", s.ObjectRead)

	address := fmt.Sprintf("0.0.0.0:%d", port)
	srv := http.Server{
		Addr:    address,
		Handler: router,
	}

	// 开启服务器
	logger.GetLogger().Info("server begin start")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.GetLogger().Error("server side error: ", err)
	}

}
