package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// PARAMS
// user presense is 2 second ping timeout (inactive)
// 2 min disconnect for tcp/ websockets
// 2 hrs token expiration

const (
	presenseTO   = 2           // seconds
	disconnectTO = 2 * 60      // 2 min
	tokenTO      = 2 * 60 * 60 // 2 hrs
)

var (

	// log-in/out, tokens (TO)
	auth = make(chan map[string]string)

	// make, del, addMsg, get
	// manage: rooms and messages in the rooms
	// indicate presense: ... 2sec
	roomMgr = make(chan map[string]string)

	// one go routine and chan per authenticated user (with token)
	// 2sec (presense), 2min disconnect (reconnect if you need)

)

func main() {
	// TCP only and Websocket

	// TODO: TCP manager

	// TODO: websocket

	// TODO: Authentication and token verification is one Chan

	// TODO: Rooms manager
	// Make, Del, AddMsg, Get([from:to], default last 10)

	// TODO: connecting via one proto disconnects previous

	// http Manager, client pull only no push
	http.HandleFunc("/", rootHander)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// PWD мап пользователей и хаш паролей
var PWD = map[string][32]byte{

	// понятное дело это делается один раз во время регистрации нового пользователя
	// и далее сохраняется в файле или БД
	"ivan1": sha256.Sum256([]byte("xyz")),
	"vova2": sha256.Sum256([]byte("abc")),
	"dima3": sha256.Sum256([]byte("ijk")),
}

// TKN мап токенов
var TKN = map[string]string{}

// login проверяет пароль и выдает токен
func login(params map[string]string) (map[string]string, error) {
	user, uok := params["user"]
	pwd, pok := params["pwd"]
	if !(uok && pok) {
		return nil, fmt.Errorf("missing 'user' and/or 'pwd'")
	}

	if PWD[user] != sha256.Sum256([]byte(pwd)) {
		return nil, fmt.Errorf("user-password missmatch")
	}

	token := fmt.Sprintf("%x", uuid.New())
	TKN[user] = token

	repl := make(map[string]string)
	repl["user"] = user
	repl["token"] = token

	return repl, nil
}

// logout удаляет токен из
func logout(params map[string]string) (map[string]string, error) {
	fmt.Println("login")
	return nil, nil
}

// expecting parameter "msg"
func post(params map[string]string) (map[string]string, error) {
	fmt.Println("login")
	return nil, nil
}

// change chat room, expecting param "room"
func room(params map[string]string) (map[string]string, error) {
	fmt.Println("login")
	return nil, nil
}

// OPS список операций
var OPS = map[string]func(map[string]string) (map[string]string, error){
	"login":  login,
	"logout": logout,
	"post":   post,
	"room":   room,
}

func rootHander(w http.ResponseWriter, r *http.Request) {
	d, err := getMsg(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// If it is, allow CORS.
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	fmt.Printf("%v\n", d)
	op, ok := d["op"]
	if !ok {
		fmt.Fprintf(os.Stderr, "missing 'op' in the json data\n%v\n", d)
		return
	}

	fmt.Fprintf(w, `{"Status":"OK", "op":"%s"}`, op)

}

func getMsg(r io.ReadCloser) (map[string]string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func runServer() {
	fmt.Println("Launching server...")
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listening: %v\n", err)
		return
	}
	defer ln.Close()

	for {
		fmt.Println("Waiting for connection...")
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error accepting: %v\n", err)
			return
		}
		fmt.Println("New connection. Waiting for messages...")

		go func(conn net.Conn) { // обрабатываем в го рутине чтобы мы могли общаться с кучей клиентов
			defer conn.Close()

			fmt.Printf("Serving new conn %v\n", conn)

			connReader := bufio.NewReader(conn) // ридер создается один раз

			for {
				message, err := connReader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						fmt.Printf("Connection %v closed.\n", conn)
						break
					}
					fmt.Fprintf(os.Stderr, "error reading from conn: %v\n", err)
					break
				}
				message = message[:len(message)-1] // удаляем \n

				fmt.Printf("From: %v Received: %s\n", conn, string(message))

				newmessage := strings.ToUpper(message)
				_, err = conn.Write([]byte(newmessage + "\n"))
				if err != nil {
					fmt.Fprintf(os.Stderr, "error writing to conn: %v\n", err)
					break
				}
			}

			fmt.Printf("Done serving client %v\n", conn)
		}(conn)
	}
}

func runClient() {
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error dialing tcp: %v\nServer needs to be runing before client.\n", err)
		return
	}
	defer conn.Close()

	// ридеры следует создавать один раз а не для каждого сообщения
	console := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)

	for {
		fmt.Print("Ваше сообщение: ")
		text, err := console.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading string: %v", err)
			return
		}
		text = text[:len(text)-1] // удаляем \n

		fmt.Fprintf(conn, text+"\n")

		fmt.Printf("%#x\n", text)

		if text == "exit" {
			fmt.Println("Закрываем соединение")
			return
		}

		message, err := connReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading from conn: %v", err)
			return
		}
		message = message[:len(message)-1] // удаляем \n
		fmt.Printf("От сервера: %s\n", message)
	}

}
