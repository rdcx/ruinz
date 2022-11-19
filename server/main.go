package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const MAX_SPEED = 10
const TICK_RATE = 10

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Spaceship struct {
	// Name of the ship
	Name     string   `json:"name"`
	Position Position `json:"position"`
}

func (s *Spaceship) Move(velocity int, direction string) {
	// Move the spaceship
	if direction == "up" {
		s.Position.Y -= MAX_SPEED
	}

	if direction == "down" {
		s.Position.Y += MAX_SPEED
	}

	if direction == "left" {
		s.Position.X -= MAX_SPEED
	}

	if direction == "right" {
		s.Position.X += MAX_SPEED
	}
}

type Keys struct {
	Up    bool `json:"up"`
	Down  bool `json:"down"`
	Left  bool `json:"left"`
	Right bool `json:"right"`
}

type User struct {
	Socket    *websocket.Conn
	Spaceship *Spaceship
}

type Response struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (u User) Message(msg string) {
	res := &Response{
		Type: "log",
		Data: msg,
	}
	data, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	u.Socket.WriteMessage(websocket.TextMessage, data)
}

func (u User) handleCommand(command []byte) {

	// JSON decode command
	var keys Keys
	json.Unmarshal(command, &keys)

	// Validate the command
	if keys.Up {
		u.Spaceship.Move(MAX_SPEED, "up")
		u.Message(fmt.Sprintf("moved up by %v, x=%v y=%v", MAX_SPEED, u.Spaceship.Position.X, u.Spaceship.Position.Y))
	}

	if keys.Down {
		u.Spaceship.Move(MAX_SPEED, "down")
		u.Message(fmt.Sprintf("moved down by %v, x=%v y=%v", MAX_SPEED, u.Spaceship.Position.X, u.Spaceship.Position.Y))
	}

	if keys.Left {
		u.Spaceship.Move(MAX_SPEED, "left")
		u.Message(fmt.Sprintf("moved left by %v, x=%v y=%v", MAX_SPEED, u.Spaceship.Position.X, u.Spaceship.Position.Y))
	}

	if keys.Right {
		u.Spaceship.Move(MAX_SPEED, "right")
		u.Message(fmt.Sprintf("moved right by %v, x=%v y=%v", MAX_SPEED, u.Spaceship.Position.X, u.Spaceship.Position.Y))
	}
}

type UserPool struct {
	Connections map[*User]bool
}

type State struct {
	Spaceships []*Spaceship `json:"spaceships"`
}

func (up *UserPool) UpdateState() {
	// get all user spaceship locations
	state := &State{}
	for user := range up.Connections {
		state.Spaceships = append(state.Spaceships, user.Spaceship)
	}

	// send state to all users
	for user := range up.Connections {
		res := &Response{
			Type: "state",
			Data: state,
		}
		data, err := json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}

		user.Socket.WriteMessage(websocket.TextMessage, data)
	}
}

func main() {

	// server static html file
	http.Handle("/", http.FileServer(http.Dir("./static")))

	pool := &UserPool{
		Connections: make(map[*User]bool),
	}

	// handle gorrila websocket
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// upgrade http to websocket
		ws, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		// Create a new spaceship
		spaceship := &Spaceship{
			Name:     "Millennium Falcon",
			Position: Position{X: 0, Y: 0},
		}

		// Create a new user
		user := &User{
			Socket:    ws,
			Spaceship: spaceship,
		}

		pool.Connections[user] = true

		// read message from websocket
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				delete(pool.Connections, user)
				ws.Close()
				return
			}

			user.handleCommand(msg)
		}
	})

	go func() {
		for {
			time.Sleep(time.Second / TICK_RATE)
			pool.UpdateState()
		}
	}()

	// start server
	http.ListenAndServe(":8080", nil)

}
