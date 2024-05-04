package main

import (
	"crypto/tls"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"net/http"
	"os"
	"time"
)

var cpuPercent []float64
var signingKey []byte

var upgrader = websocket.Upgrader{}

var connections = make(map[string]*logConnection)

func loadKey() {
	err := os.Chmod("key.key", 0600)
	key, err := os.ReadFile("key.key")
	if err != nil {
		panic(err)
	}

	signingKey = key
}

func generateKey(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24 * 31).Unix()

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getTokenExpiration(tokenString string) (TokenStatus, error) {
	println(tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return TokenStatus{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return TokenStatus{
			Valid: false,
		}, fmt.Errorf("invalid token claims")
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return TokenStatus{
			Valid: false,
		}, fmt.Errorf("invalid token expiration")
	}
	// check if token is expired
	var valid = time.Now().Unix() < int64(exp)

	return TokenStatus{
		Valid: valid,
		Exp:   time.Unix(int64(exp), 0),
	}, nil
}

func getCPULoop() {
	for {
		cpuPercent, _ = cpu.Percent(time.Second, true)
	}
}

func main() {
	loadKey()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "generate":
			if len(os.Args) < 3 {
				fmt.Println("Usage: ./server generate <userID>")
				return
			}
			token, err := generateKey(os.Args[2])
			if err != nil {
				fmt.Println("Error generating key:", err)
				return
			}
			fmt.Println("Generated token (valid for 31d, app will regenerate it):", token)
			return
		case "help":
			fmt.Println("Usage: ./server [generate|help]")
			return
		}
	}

	go getCPULoop()

	r := mux.NewRouter()
	r.HandleFunc("/log", logHandler)
	r.HandleFunc("/dir", dirHandler).Methods("GET")

	r.HandleFunc("/stats", statsHandler).Methods("GET")
	r.HandleFunc("/info", infoHandler).Methods("GET")
	r.HandleFunc("/token", regenerateTokenHandler).Methods("GET")
	r.HandleFunc("/token/exp", getTokenExpirationHandler).Methods("GET")

	r.Use(authMiddleware)

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		},
	}

	fmt.Printf("Server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
