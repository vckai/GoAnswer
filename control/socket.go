package control

import (
	"fmt"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
)

/**
 * 监听websocket请求信息
 */
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(messageSize)
	c.ws.SetWriteDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	fmt.Println("启动一个websocket，UserID：", c.userId)

	for {
		_, message, err := c.ws.ReadMessage()
		fmt.Println("Socket读取: ", string(message))
		if err != nil {
			fmt.Println("读取socket信息错误，---ERROR：", err)
			break
		}
		param, err := simplejson.NewJson([]byte(message))

		if param == nil || err != nil {
			fmt.Println("传输数据错误, 数据: ", param, "---ERROR: ", err)
			break
		}
		h.broadcast <- param
	}
}

/**
 * 通过socket发送消息
 */
func (c *connection) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, message)
}

/**
 * 监听websocket的发送信息
 */
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				fmt.Println("关闭一个连接")
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			fmt.Println("发送消息到客户端，Message：", string(message))
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
