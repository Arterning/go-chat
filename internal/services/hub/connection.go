package hub

import (
	"encoding/json"
	"go-chat/internal/models"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入超时时间
	writeWait = 10 * time.Second

	// Pong 超时时间
	pongWait = 60 * time.Second

	// Ping 间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Connection WebSocket 连接包装
type Connection struct {
	ws *websocket.Conn
}

// NewConnection 创建新的连接
func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{ws: ws}
}

// ReadPump 从 WebSocket 读取消息
func (c *Client) ReadPump(saveMessage func(*models.Message) error) {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.ws.Close()
	}()

	c.Conn.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.ws.SetPongHandler(func(string) error {
		c.Conn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var wsMsg models.WebSocketMessage
		err := c.Conn.ws.ReadJSON(&wsMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 处理不同类型的消息
		switch wsMsg.Type {
		case "message":
			// 创建消息对象
			msg := &models.Message{
				RoomID:    c.RoomID,
				UserID:    c.UserID,
				Username:  c.Username,
				Content:   wsMsg.Content,
				CreatedAt: time.Now(),
			}

			// 保存消息到数据库
			if err := saveMessage(msg); err != nil {
				log.Printf("Error saving message: %v", err)
				errorMsg := models.WebSocketMessage{
					Type:  "error",
					Error: "Failed to save message",
				}
				c.SendMessage(errorMsg)
				continue
			}

			// 广播消息到房间的所有客户端
			broadcastMsg := models.WebSocketMessage{
				Type:    "message",
				RoomID:  c.RoomID,
				Message: msg,
			}

			messageBytes, err := json.Marshal(broadcastMsg)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			c.Hub.Broadcast(c.RoomID, messageBytes, nil)
		}
	}
}

// WritePump 向 WebSocket 写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				c.Conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息一起发送
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 发送消息给客户端
func (c *Client) SendMessage(msg models.WebSocketMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
	default:
		// 通道已满，关闭连接
		close(c.Send)
		return err
	}

	return nil
}
