package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

func main() {

	PORT := ":9001"
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	wbMap := make(map[string]*websocket.Conn)

	// websocket server
	go httpStart(&wbMap)

	// tcp server
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, &wbMap)
	}
}

func handleConnection(c net.Conn, wbMap *map[string]*websocket.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		buf := make([]byte, 1024)

		lengthPer, err := bufio.NewReader(c).Read(buf)
		buf = buf[:lengthPer]
		temp := strings.TrimSpace(string(buf))
		fmt.Println("tcp receive ==>> ", time.Now().Format("[2006-01-02 15:04:05]"), " ", "hex==> ", hex.EncodeToString(buf), " ", "string ==>>", temp)

		for _, v := range *wbMap {
			if v != nil {
				v.WriteMessage(1, buf)
			}
		}

		if err != nil {
			fmt.Println(err)
			break
		}

	}
	c.Close()
}

func httpStart(wbMap *map[string]*websocket.Conn) {
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Upgrade fail", http.StatusBadRequest)
			return
		}
		v4, err := uuid.NewV4()
		key := v4.String()

		// add websocket conn to list
		(*wbMap)[key] = conn

		// read data from client
		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					conn.Close()
					delete(*wbMap, key)
					fmt.Println(err)
				}
				break
			}
			fmt.Println("websocket receive ==>> ", time.Now().Format("[2006-01-02 15:04:05]"), " ", msgType, strings.TrimSpace(string(data)))
		}

	})

	http.ListenAndServe(":9002", nil)
}
