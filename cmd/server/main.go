package main

import (
	"go-chat/internal/database"
	"go-chat/internal/handlers"
	"go-chat/internal/middleware"
	"go-chat/internal/services/hub"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件（如果存在）
	// 尝试从当前目录和项目根目录加载
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, using environment variables or defaults")
		} else {
			log.Println("Loaded configuration from .env file")
		}
	} else {
		log.Println("Loaded configuration from .env file")
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 运行数据库迁移
	if err := database.RunMigrations("../../migrations/001_init.sql"); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
		log.Println("Continuing anyway... (migrations may have already been applied)")
	}

	// 创建并启动 WebSocket Hub
	wsHub := hub.NewHub()
	go wsHub.Run()

	// 创建路由
	r := mux.NewRouter()

	// 静态文件
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../../web/static"))))

	// 公开路由
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	r.HandleFunc("/login", handlers.ShowLoginPage).Methods("GET")
	r.HandleFunc("/register", handlers.ShowRegisterPage).Methods("GET")
	r.HandleFunc("/api/login", handlers.Login).Methods("POST")
	r.HandleFunc("/api/register", handlers.Register).Methods("POST")
	r.HandleFunc("/logout", handlers.Logout).Methods("GET")

	// 需要认证的路由
	authRouter := r.PathPrefix("/").Subrouter()
	authRouter.Use(middleware.RequireAuth)

	// 房间相关路由
	authRouter.HandleFunc("/rooms", handlers.ShowRoomsList).Methods("GET")
	authRouter.HandleFunc("/rooms/{id:[0-9]+}", handlers.ShowRoom).Methods("GET")
	authRouter.HandleFunc("/api/rooms", handlers.CreateRoom).Methods("POST")
	authRouter.HandleFunc("/api/rooms/{id:[0-9]+}/members", handlers.GetRoomMembers).Methods("GET")
	authRouter.HandleFunc("/api/rooms/{id:[0-9]+}/invite", handlers.InviteMember).Methods("POST")
	authRouter.HandleFunc("/api/rooms/{id:[0-9]+}/members/{memberId:[0-9]+}", handlers.RemoveMember).Methods("DELETE")
	authRouter.HandleFunc("/api/rooms/{id:[0-9]+}/leave", handlers.LeaveRoom).Methods("POST")

	// WebSocket 路由
	authRouter.HandleFunc("/ws/rooms/{id:[0-9]+}", handlers.HandleWebSocket(wsHub)).Methods("GET")

	// 启动服务器
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Printf("Please ensure PostgreSQL is running and the database 'gochat' exists")
	log.Printf("You can create it with: CREATE DATABASE gochat;")

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
