package hub

import (
	"go-chat/internal/models"
	"log"
	"sync"
)

// Client 代表一个 WebSocket 客户端
type Client struct {
	Hub      *Hub
	Conn     *Connection
	RoomID   int
	UserID   int
	Username string
	Send     chan []byte
}

// Hub 管理所有活跃的客户端和房间
type Hub struct {
	// rooms 存储每个房间的所有客户端
	// key: roomID, value: map of clients
	rooms map[int]map[*Client]bool

	// broadcast 广播消息到特定房间
	broadcast chan *BroadcastMessage

	// register 注册新客户端
	register chan *Client

	// unregister 注销客户端
	unregister chan *Client

	// mutex 用于并发安全
	mu sync.RWMutex
}

// BroadcastMessage 广播消息结构
type BroadcastMessage struct {
	RoomID  int
	Message []byte
	Sender  *Client // 可选，用于排除发送者
}

// NewHub 创建新的 Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int]map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.RoomID] == nil {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()

			log.Printf("Client registered: User %d (%s) joined room %d",
				client.UserID, client.Username, client.RoomID)

			// 通知房间其他成员有新用户加入
			joinMsg := models.WebSocketMessage{
				Type:     "join",
				RoomID:   client.RoomID,
				UserID:   client.UserID,
				Username: client.Username,
			}
			h.BroadcastToRoom(client.RoomID, joinMsg, nil)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.RoomID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)

					// 如果房间没有客户端了，删除房间
					if len(clients) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

			log.Printf("Client unregistered: User %d (%s) left room %d",
				client.UserID, client.Username, client.RoomID)

			// 通知房间其他成员有用户离开
			leaveMsg := models.WebSocketMessage{
				Type:     "leave",
				RoomID:   client.RoomID,
				UserID:   client.UserID,
				Username: client.Username,
			}
			h.BroadcastToRoom(client.RoomID, leaveMsg, nil)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.rooms[msg.RoomID]; ok {
				for client := range clients {
					// 如果指定了发送者，则不发送给发送者自己
					if msg.Sender != nil && client == msg.Sender {
						continue
					}

					select {
					case client.Send <- msg.Message:
					default:
						// 发送失败，关闭连接
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register 注册客户端到 Hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 从 Hub 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast 向房间广播消息
func (h *Hub) Broadcast(roomID int, message []byte, sender *Client) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Sender:  sender,
	}
}

// BroadcastToRoom 向指定房间广播消息
func (h *Hub) BroadcastToRoom(roomID int, message models.WebSocketMessage, excludeClient *Client) {
	// 此函数在外部调用时使用
	// 实际的广播逻辑在 Run() 的 broadcast channel 中处理
}

// GetRoomClients 获取房间的所有客户端
func (h *Hub) GetRoomClients(roomID int) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if roomClients, ok := h.rooms[roomID]; ok {
		for client := range roomClients {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetRoomClientCount 获取房间的客户端数量
func (h *Hub) GetRoomClientCount(roomID int) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		return len(clients)
	}
	return 0
}
