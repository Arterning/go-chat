package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("your-secret-key-change-this-in-production"))

// RequireAuth 要求用户必须登录
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		userID, ok := session.Values["user_id"]

		if !ok || userID == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserID 从 session 中获取用户 ID
func GetUserID(r *http.Request) (int, bool) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	return userID, ok
}

// GetUsername 从 session 中获取用户名
func GetUsername(r *http.Request) (string, bool) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"].(string)
	return username, ok
}
