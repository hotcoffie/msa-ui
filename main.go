package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"msa-ui/handler"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

const logPath = "logs"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFileName := filepath.Join(logPath, "ui.log")
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Panic("打开日志文件：", err)
	}
	log.SetOutput(logFile)

	r := gin.Default()
	r.Static("static", "static")
	api := r.Group("/api")
	{
		api.GET("/init", handler.InitHandler)
		api.POST("/saveUser", handler.SaveUserHandler)
		api.GET("/run", handler.GinWebsocketHandler(handler.RunHandle))
	}
	go r.Run()
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "start", "http://localhost:8080/static")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}

	c := make(chan os.Signal)
	// 监听退出信号量
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}
