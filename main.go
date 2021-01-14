package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"msa-ui/handler"
	"msa-ui/messageBox"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

const logPath = "logs"

func main() {
	logFile := getLogFile()
	defer logFile.Close()
	log.SetOutput(logFile)

	r := gin.Default()
	r.Static("static", "static")
	api := r.Group("/api")
	{
		api.GET("/init", handler.InitHandler)
		api.POST("/saveUser", handler.SaveUserHandler)
		api.GET("/run", handler.GinWebsocketHandler(handler.RunHandle))
	}

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		if runtime.GOOS == "windows" {
			messageBox.Show("错误！", "服务启动失败！具体错误信息请查看 logs/ui.log")
		}
		log.Panic("服务启动失败: ", err)
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", "start", "http://localhost:8080/static")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}
	// 绑定到server上
	if err = http.Serve(l, r); err != nil {
		log.Panic("服务异常终止: ", err)
	}
}

func getLogFile() *os.File {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err = os.Mkdir(logPath, os.ModePerm)
		if err != nil {
			log.Panic("日志目录创建失败", err)
		}
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFileName := filepath.Join(logPath, "ui.log")
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Panic("打开日志文件：", err)
	}
	return logFile
}
