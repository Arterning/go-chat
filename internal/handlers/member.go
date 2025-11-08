package handlers

import (
	"database/sql"
	"encoding/json"
	"go-chat/internal/database"
	"go-chat/internal/middleware"
	"go-chat/internal/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// InviteMember 邀请成员加入房间
func InviteMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 检查当前用户是否是房间的创建者
	var role string
	err = database.DB.QueryRow(
		"SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2",
		roomID, currentUserID,
	).Scan(&role)

	if err == sql.ErrNoRows {
		http.Error(w, "You are not a member of this room", http.StatusForbidden)
		return
	} else if err != nil {
		log.Printf("Error checking user role: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if role != "creator" {
		http.Error(w, "Only the room creator can invite members", http.StatusForbidden)
		return
	}

	var req models.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// 查找要邀请的用户
	var invitedUserID int
	err = database.DB.QueryRow(
		"SELECT id FROM users WHERE username = $1",
		req.Username,
	).Scan(&invitedUserID)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error finding user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 检查用户是否已经是房间成员
	var exists bool
	err = database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)",
		roomID, invitedUserID,
	).Scan(&exists)

	if err != nil {
		log.Printf("Error checking membership: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "User is already a member of this room", http.StatusConflict)
		return
	}

	// 添加用户到房间
	_, err = database.DB.Exec(
		"INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, $3)",
		roomID, invitedUserID, "member",
	)

	if err != nil {
		log.Printf("Error adding member: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Member invited successfully",
	})
}

// RemoveMember 从房间移除成员
func RemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	memberID, err := strconv.Atoi(vars["memberId"])
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 检查当前用户是否是房间的创建者
	var role string
	err = database.DB.QueryRow(
		"SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2",
		roomID, currentUserID,
	).Scan(&role)

	if err == sql.ErrNoRows {
		http.Error(w, "You are not a member of this room", http.StatusForbidden)
		return
	} else if err != nil {
		log.Printf("Error checking user role: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if role != "creator" {
		http.Error(w, "Only the room creator can remove members", http.StatusForbidden)
		return
	}

	// 不能移除创建者
	var memberRole string
	err = database.DB.QueryRow(
		"SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2",
		roomID, memberID,
	).Scan(&memberRole)

	if err == sql.ErrNoRows {
		http.Error(w, "Member not found in this room", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking member role: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if memberRole == "creator" {
		http.Error(w, "Cannot remove the room creator", http.StatusForbidden)
		return
	}

	// 移除成员
	_, err = database.DB.Exec(
		"DELETE FROM room_members WHERE room_id = $1 AND user_id = $2",
		roomID, memberID,
	)

	if err != nil {
		log.Printf("Error removing member: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Member removed successfully",
	})
}

// LeaveRoom 离开房间
func LeaveRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 检查用户的角色
	var role string
	err = database.DB.QueryRow(
		"SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2",
		roomID, currentUserID,
	).Scan(&role)

	if err == sql.ErrNoRows {
		http.Error(w, "You are not a member of this room", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking user role: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if role == "creator" {
		http.Error(w, "Room creator cannot leave the room. Please delete the room instead.", http.StatusForbidden)
		return
	}

	// 离开房间
	_, err = database.DB.Exec(
		"DELETE FROM room_members WHERE room_id = $1 AND user_id = $2",
		roomID, currentUserID,
	)

	if err != nil {
		log.Printf("Error leaving room: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Left room successfully",
	})
}
