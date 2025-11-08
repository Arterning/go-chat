package handlers

import (
	"encoding/json"
	"go-chat/internal/database"
	"go-chat/internal/middleware"
	"go-chat/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ShowRoomsList 显示房间列表页面
func ShowRoomsList(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	// 获取用户加入的所有房间
	rows, err := database.DB.Query(`
		SELECT r.id, r.name, r.description, r.creator_id, r.created_at
		FROM rooms r
		INNER JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY r.created_at DESC
	`, userID)

	if err != nil {
		log.Printf("Error querying rooms: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID, &room.CreatedAt); err != nil {
			log.Printf("Error scanning room: %v", err)
			continue
		}
		rooms = append(rooms, room)
	}

	username, _ := middleware.GetUsername(r)
	data := struct {
		Rooms    []models.Room
		Username string
	}{
		Rooms:    rooms,
		Username: username,
	}

	tmpl := template.Must(template.ParseFiles("web/templates/rooms.html"))
	tmpl.Execute(w, data)
}

// CreateRoom 创建新房间
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Room name is required", http.StatusBadRequest)
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 创建房间
	var roomID int
	err = tx.QueryRow(
		"INSERT INTO rooms (name, description, creator_id) VALUES ($1, $2, $3) RETURNING id",
		req.Name, req.Description, userID,
	).Scan(&roomID)

	if err != nil {
		log.Printf("Error creating room: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 将创建者添加为房间成员
	_, err = tx.Exec(
		"INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, $3)",
		roomID, userID, "creator",
	)

	if err != nil {
		log.Printf("Error adding creator to room: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"room_id": roomID,
		"message": "Room created successfully",
	})
}

// ShowRoom 显示聊天室页面
func ShowRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	userID, _ := middleware.GetUserID(r)

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

	// 获取房间信息
	var room models.Room
	err = database.DB.QueryRow(
		"SELECT id, name, description, creator_id FROM rooms WHERE id = $1",
		roomID,
	).Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID)

	if err != nil {
		log.Printf("Error querying room: %v", err)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// 获取历史消息
	rows, err := database.DB.Query(`
		SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m
		INNER JOIN users u ON m.user_id = u.id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT 50
	`, roomID)

	if err != nil {
		log.Printf("Error querying messages: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt); err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// 反转消息顺序（最旧的在前）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	username, _ := middleware.GetUsername(r)
	data := struct {
		Room     models.Room
		Messages []models.Message
		UserID   int
		Username string
	}{
		Room:     room,
		Messages: messages,
		UserID:   userID,
		Username: username,
	}

	tmpl := template.Must(template.ParseFiles("web/templates/room.html"))
	tmpl.Execute(w, data)
}

// GetRoomMembers 获取房间成员列表
func GetRoomMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(`
		SELECT u.id, u.username, rm.role
		FROM users u
		INNER JOIN room_members rm ON u.id = rm.user_id
		WHERE rm.room_id = $1
	`, roomID)

	if err != nil {
		log.Printf("Error querying room members: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Member struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	var members []Member
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.ID, &member.Username, &member.Role); err != nil {
			log.Printf("Error scanning member: %v", err)
			continue
		}
		members = append(members, member)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
