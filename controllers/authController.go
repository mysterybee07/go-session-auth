package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type session struct {
	username string
	role     string
	expiry   time.Time
}

var (
	user = map[string]struct {
		Password string
		Role     string
	}{
		"user1": {"password123", "user"},
		"user2": {"password123", "admin"},
	}

	sessions = make(map[string]session)
)

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "invalid request payload",
		})
		return
	}

	userData, exist := user[creds.Username]
	if !exist || userData.Password != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unauthorized User",
		})
		return
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(5 * time.Minute)

	sessions[sessionToken] = session{
		username: creds.Username,
		role:     userData.Role,
		expiry:   expiresAt,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_cookie",
		Expires:  expiresAt,
		Value:    sessionToken,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "user login successfull",
		"token":   sessionToken,
		"role":    userData.Role,
	})

}

func Home(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_cookie")
	if err != nil {
		if err == http.ErrNoCookie {
			json.NewEncoder(w).Encode(map[string]string{
				"message": "No cookies found",
			})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionToken := cookie.Value

	userSession, exists := sessions[sessionToken]

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Write([]byte(fmt.Sprintf("Welcome %s ! Your Role is %s", userSession.username, userSession.role)))
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_cookie")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, `{"message": "No session cookie provided"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	sessionToken := cookie.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		http.Error(w, `{"message": "Invalid session token"}`, http.StatusUnauthorized)
		return
	}

	if userSession.isExpired() {
		delete(sessions, sessionToken)
		http.Error(w, `{"message": "Session expired"}`, http.StatusUnauthorized)
		return
	}

	// Generate a new session token
	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(30 * time.Minute) // Extend session duration

	// Store new session and delete the old one
	sessions[newSessionToken] = session{
		username: userSession.username,
		role:     userSession.role,
		expiry:   expiresAt,
	}
	delete(sessions, sessionToken)

	// Set a new cookie with updated session token
	http.SetCookie(w, &http.Cookie{
		Name:     "session_cookie",
		Value:    newSessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Respond with session details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Session refreshed successfully",
		"username": userSession.username,
		"role":     userSession.role,
		"expiry":   expiresAt,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session cookie
	cookie, err := r.Cookie("session_cookie")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, `{"message": "No session cookie provided"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Get the session token from the cookie
	sessionToken := cookie.Value

	// Remove the session from the in-memory store
	delete(sessions, sessionToken)

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_cookie",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Second),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
