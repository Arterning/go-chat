package handlers

import (
	"go-chat/internal/database"
	"go-chat/internal/middleware"
	"go-chat/internal/models"
	"go-chat/internal/services/hub"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境中应该验证 origin
	},
}

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取房间 ID
		vars := mux.Vars(r)
		roomID, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}

		// 获取用户信息
		userID, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, _ := middleware.GetUsername(r)

		// 检查用户是否是房间成员
		var exists bool
		err = database.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)",
			roomID, userID,
		).Scan(&exists)

		if err != nil || !exists {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// 升级 HTTP 连接到 WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		// 创建客户端
		client := &hub.Client{
			Hub:      h,
			Conn:     hub.NewConnection(conn),
			RoomID:   roomID,
			UserID:   userID,
			Username: username,
			Send:     make(chan []byte, 256),
		}

		// 注册客户端
		client.Hub.Register(client)

		// 启动读写协程
		go client.WritePump()
		go client.ReadPump(saveMessageToDB)
	}
}

// saveMessageToDB 保存消息到数据库
func saveMessageToDB(msg *models.Message) error {
	return database.DB.QueryRow(
		"INSERT INTO messages (room_id, user_id, content, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		msg.RoomID, msg.UserID, msg.Content, msg.CreatedAt,
	).Scan(&msg.ID)
}
