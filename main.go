package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"msa-ui/conf"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

type ConfData struct {
	Active    string
	Username  string
	Password  string
	Points    string
	ThreadNum int
	Info      string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFile, err := os.OpenFile("msa.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Panic("打开日志文件：", err)
	}
	log.SetOutput(logFile)

	r := gin.Default()
	r.Static("static", "static")
	api := r.Group("/api")
	{
		api.GET("/init", initHandler)
		api.POST("/saveUser", saveUserHandler)
		api.GET("/run", ginWebsocketHandler(wsConnHandle))
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
func initHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 200,
		"data": gin.H{
			"users":     conf.Users,
			"threadNum": runtime.NumCPU(),
		},
	})
}
func saveUserHandler(c *gin.Context) {
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

func ginWebsocketHandler(wsConnHandle websocket.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("终端接入", "ip:", c.Request.RemoteAddr)
		if c.IsWebsocket() {
			wsConnHandle.ServeHTTP(c.Writer, c.Request)
		} else {
			_, _ = c.Writer.WriteString("===not websocket request===")
		}
	}
}

func wsConnHandle(conn *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			sendMsg(conn, fmt.Sprintf("连接故障: %s", err))
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "msa")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sendMsg(conn, fmt.Sprintf("获取抢号信息失败: %s", err))
		return
	}
	reader := bufio.NewReader(stdout)
	go func() {
		for {
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				log.Println(err2)
				conn.Close()
				break
			}
			sendMsg(conn, line)
		}
	}()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		sendMsg(conn, fmt.Sprintf("获取抢号信息失败: %s", err))
		return
	}
	reader2 := bufio.NewReader(stderr)
	go func() {
		for {
			line, err2 := reader2.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				log.Println(err2)
				conn.Close()
				break
			}
			log.Print("[error] ", line)
			sendMsg(conn, line)
		}
	}()
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
		cd := new(ConfData)
		if err := json.Unmarshal(msg, cd); err != nil {
			sendMsg(conn, fmt.Sprintf("解析ws消息失败: %s", err))
			return
		}
		if err := writDataToFile(cd); err != nil {
			sendMsg(conn, fmt.Sprintf("初始化配置失败: %s", err))
			return
		}
		if err := cmd.Start(); err != nil {
			sendMsg(conn, fmt.Sprintf("启动程序失败: %s", err))
			return
		}
	}
}

func writDataToFile(cd *ConfData) error {
	confFile, err := os.OpenFile("conf.yml", os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.WithMessage(err, "打开conf.yml")
	}
	defer confFile.Close()
	confStr := fmt.Sprintf(`# 系统状态
# prod - 正式抢点
# dev - 试访问网站检查参数是否正确
active: %s

# 账号密码
username: %s
password: %s

#可用时间点，时间点空格分隔
points: %s

# 同一个点用几个线程去抢
threadNum: %d`, cd.Active, cd.Username, cd.Password, cd.Points, cd.ThreadNum)
	if _, err = confFile.Write([]byte(confStr)); err != nil {
		return errors.WithMessage(err, "修改conf.yml")
	}

	infoFile, err := os.OpenFile("info.yml", os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.WithMessage(err, "打开info.yml")
	}
	defer infoFile.Close()
	if _, err = infoFile.Write([]byte(cd.Info)); err != nil {
		return errors.WithMessage(err, "修改info.yml")
	}
	return nil
}

func sendMsg(conn *websocket.Conn, msg string) {
	if _, err := conn.Write([]byte(msg)); err != nil {
		log.Println("[error]", "发送ws消息:", err)
		return
	}
}
