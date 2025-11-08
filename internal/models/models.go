package models

import "time"

// User 用户模型
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // 不在 JSON 中显示密码
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session 会话模型
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Data      string    `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Room 聊天室模型
type Room struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoomMember 房间成员模型
type RoomMember struct {
	ID       int       `json:"id"`
	RoomID   int       `json:"room_id"`
	UserID   int       `json:"user_id"`
	Role     string    `json:"role"` // "creator" 或 "member"
	JoinedAt time.Time `json:"joined_at"`
}

// Message 消息模型
type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"` // 用于显示
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InviteMemberRequest 邀请成员请求
type InviteMemberRequest struct {
	Username string `json:"username"`
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Type    string      `json:"type"` // "message", "join", "leave", "error"
	RoomID  int         `json:"room_id,omitempty"`
	Message *Message    `json:"message,omitempty"`
	Content string      `json:"content,omitempty"`
	Error   string      `json:"error,omitempty"`
	UserID  int         `json:"user_id,omitempty"`
	Username string     `json:"username,omitempty"`
}
