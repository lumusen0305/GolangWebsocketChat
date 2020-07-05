package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type subscription struct {
	Sender string
	conn *connection
}

type Hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	sub map[*connection]bool
	// Inbound messages from the connections.
	broadcast chan Message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}
var hub = Hub{
	broadcast:  make(chan Message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}
func (h *Hub) run() {
	for {
		select {
		case s := <-h.register:
			//===========SearchData==============
			pool, err := sqlx.Open("mysql", dsn)
			if err != nil {
				panic(err)
			}
			var roomid []string
			err = pool.Select(&roomid, "SELECT roomid FROM singlechatroom where userid = ?",s.Sender)
			//=====================================
			for _,chatroom := range roomid {
				connections := h.rooms[chatroom]
				//＝＝＝＝＝＝＝＝房間添加bool空間＝＝＝＝＝＝＝
				if connections == nil {
					connections = make(map[*connection]bool)
					h.rooms[chatroom] = connections
				}
				//＝＝＝＝＝＝＝＝＝＝＝＝把房間對應map設為ture＝＝＝＝＝＝＝＝＝＝＝＝＝
				h.rooms[chatroom][s.conn] = true
				fmt.Print(chatroom+"號房在線人數：")
				fmt.Println(len(connections))
			}
		case s := <-h.unregister:
			pool, err := sqlx.Open("mysql", dsn)
			if err != nil {
				panic(err)
			}
			var roomid []string
			err = pool.Select(&roomid, "SELECT roomid FROM singlechatroom where userid = ?",s.Sender)
			//=====================================
			for count,chatroom := range roomid {
				connections := h.rooms[chatroom]
				if connections != nil {
					if _, ok := connections[s.conn]; ok {
						delete(connections, s.conn)
						if count==0{
							close(s.conn.send)
						}
						fmt.Print(chatroom+"號房在線人數：")
						fmt.Println(len(connections))
						if len(connections) == 0 {
							delete(h.rooms, chatroom)
							fmt.Println(chatroom+"號房以刪除")
						}
					}
				}
			}
		case m := <-h.broadcast:
			connections := h.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, m.room)
					}
				}
			}
		}
	}
}