package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

var addr = flag.String("addr", ":8080", "http service address")

func authorize(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte("secret"), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["userName"].(string), nil
	}

	return "", err
}

func main() {
	flag.Parse()
	log.Println("Chat server starting")

	s := getServer()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		username, err := authorize(q.Get("jwt"))

		if err != nil {
			log.Printf("Authorization error: %v\n", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			log.Printf("Connection: %+v, username: %s\n", r.RemoteAddr, username)
			serveWs(s, w, r, username)
		}
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
