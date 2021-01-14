package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"msa-ui/conf"
	"os/exec"
	"runtime"
	"syscall"
)

const configPath = "setting"

func InitHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 200,
		"data": gin.H{
			"users":     conf.Users,
			"threadNum": runtime.NumCPU(),
		},
	})
}
func SaveUserHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if err := conf.UpdateUser(username, password); err != nil {
		log.Print("[error] ", err)
		c.JSON(200, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
	})
}

func RunHandle(conn *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			sendMsg(conn, fmt.Sprintf("连接故障: %s", err))
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "msa")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	readCmdOut(cmd, conn)

	for {
		msg := make([]byte, 1000)
		if err := websocket.Message.Receive(conn, &msg); err != nil {
			log.Println("[error]", "读取ws消息:", err)
			return
		}
		if string(msg) == "stop" {
			cancel()
			return
		}
		cd := new(conf.ConfData)
		if err := json.Unmarshal(msg, cd); err != nil {
			sendMsg(conn, fmt.Sprintf("解析ws消息失败: %s", err))
			return
		}
		if err := conf.WritConfForRun(cd); err != nil {
			sendMsg(conn, fmt.Sprintf("初始化配置失败: %s", err))
			return
		}
		if err := cmd.Start(); err != nil {
			sendMsg(conn, fmt.Sprintf("启动程序失败: %s", err))
			return
		}
	}
}
func readCmdOut(cmd *exec.Cmd, conn *websocket.Conn) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sendMsg(conn, fmt.Sprintf("获取抢号信息失败: %s", err))
		return
	}
	go sendMsgLoop(stdout, conn, func(l string) {
		sendMsg(conn, l)
	})

	stderr, err := cmd.StderrPipe()
	if err != nil {
		sendMsg(conn, fmt.Sprintf("获取抢号信息失败: %s", err))
		return
	}
	go sendMsgLoop(stderr, conn, func(l string) {
		log.Print("error", l)
	})
}

func sendMsgLoop(rc io.ReadCloser, conn *websocket.Conn, cb func(l string)) {
	reader := bufio.NewReader(rc)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			log.Println(err2)
			conn.Close()
			break
		}
		cb(line)
	}
}
