package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return signingKey, nil
		})

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Println("Authenticated request for:", r.RequestURI, "User:", claims["user"])
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func regenerateTokenHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	token, err := generateKey(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(token))
}

func getTokenExpirationHandler(w http.ResponseWriter, r *http.Request) {
	// get token from Authorization header
	token := strings.Split(r.Header.Get("Authorization"), " ")[1]
	exp, err := getTokenExpiration(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting token expiration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exp)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	hideCPU := r.URL.Query().Get("hideCPU") == "true"
	hideRAM := r.URL.Query().Get("hideRAM") == "true"
	hideDisk := r.URL.Query().Get("hideDisk") == "true"
	hideProcs := r.URL.Query().Get("hideProcs") == "true"

	stats, err := getSystemStats(!hideCPU, !hideRAM, !hideDisk, !hideProcs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting system stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func infoHandler(w http.ResponseWriter, _ *http.Request) {
	info, err := getSystemInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving system info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP server connection to the WebSocket protocol
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	// Retrieve the log file path from the header
	path := r.Header.Get("path")
	if path == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Missing file path"))
		conn.Close()
		return
	}

	// Tail the specified log file
	t, err := tail.TailFile(path, tail.Config{Follow: true, MustExist: true})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Could not tail file: "+err.Error()))
		conn.Close()
		return
	}

	// Store the connection and tailing process
	connectionID := fmt.Sprintf("%p", conn)
	connections[connectionID] = &logConnection{conn: conn, tail: t}

	// Start sending log lines to the client
	go func() {
		for line := range t.Lines {
			err = conn.WriteMessage(websocket.TextMessage, []byte(line.Text))
			if err != nil {
				// Log the error, close the connection, and stop tailing when there's an error
				t.Stop()
				conn.Close()
				delete(connections, connectionID)
				return
			}
		}
	}()
}

func dirHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the directory path from the header
	inputPath := r.Header.Get("path")
	if inputPath == "" {
		http.Error(w, "Missing directory path", http.StatusBadRequest)
		return
	}

	// Convert path to absolute path
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		http.Error(w, "Could not resolve absolute path: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var filesDir []Directory

	// Read the directory contents
	files, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, "Could not read directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			http.Error(w, "Could not read file info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filesDir = append(filesDir, Directory{
			Path:  info.Name(),
			IsDir: info.IsDir(),
			Size:  info.Size(),
		})
	}

	// Send the directory contents as a JSON response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(DirResponse{
		Directory: absPath,
		Dirs:      filesDir,
	})

	if err != nil {
		http.Error(w, "Could not encode directory contents: "+err.Error(), http.StatusInternalServerError)
	}
}
