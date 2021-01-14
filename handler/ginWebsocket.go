package handler

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
	"log"
)

func GinWebsocketHandler(wsConnHandle websocket.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("终端接入", "ip:", c.Request.RemoteAddr)
		if c.IsWebsocket() {
			wsConnHandle.ServeHTTP(c.Writer, c.Request)
		} else {
			_, _ = c.Writer.WriteString("===not websocket request===")
		}
	}
}

func sendMsg(conn *websocket.Conn, msg string) {
	if _, err := conn.Write([]byte(msg)); err != nil {
		log.Println("[error]", "发送ws消息:", err)
		return
	}
}
