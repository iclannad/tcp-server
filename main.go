package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	client "github.com/influxdata/influxdb1-client"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"tcp-server/influxd"
	"time"
)

func main() {
	influxd.Init()

	done := make(chan bool)
	go func() {
		startPort := 10001
		for i := 0; i < 30; i++ {
			tcpPort := startPort + i
			go StartWsTcp(tcpPort)
		}
		done <- true
	}()

	<-done
	http.ListenAndServe(":9002", nil)
}

func StartWsTcp(tcpPort int)  {
	tcpPortStr := fmt.Sprint(":",tcpPort)
	fmt.Println("listen ==>",tcpPortStr)

	l, err := net.Listen("tcp4", tcpPortStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	wbMap := make(map[string]*websocket.Conn)
	tcpMap := make(map[string]net.Conn)

	// websocket server
	go httpStart(tcpPort,&wbMap,&tcpMap)

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(tcpPortStr,c, &wbMap,&tcpMap)
	}
}


func handleConnection(port string,c net.Conn, wbMap *map[string]*websocket.Conn,tcpMap *map[string]net.Conn) {
	// add tcp socket conn to list
	v4, _ := uuid.NewV4()
	key := v4.String()
	(*tcpMap)[key] = c

	con := influxd.GetInfluxCli()

	for {
		buf := make([]byte, 1024)

		lengthPer, err := bufio.NewReader(c).Read(buf)
		buf = buf[:lengthPer]
		temp := strings.TrimSpace(string(buf))
		fmt.Println("data from tcp socket", port, "  ",time.Now().Format("[2006-01-02 15:04:05]"), " ", "hex==> ", hex.EncodeToString(buf), " ", "string ==>>", temp)

		dst := make([]byte, hex.EncodedLen(len(buf)))
		hex.Encode(dst, buf)

		for _, v := range *wbMap {
			if v != nil {
				v.WriteMessage(1, dst)
			}
		}

		if err != nil {
			fmt.Println(err)
			break
		} else {
			field := make(map[string]interface{})
			field["origin_data"] = temp
			tags := make(map[string]string)

			pts := make([]client.Point, 0)
			var point client.Point
			point = client.Point{
				Measurement: "test",
				Tags: tags,
				Fields:      field,
				Time:        time.Now(),
				Precision:   "s",
			}
			pts = append(pts, point)

			bps := client.BatchPoints{
				Points:    pts,
				Database:  "tcp_server",
				Precision: "s",
			}
			_, err = con.Write(bps)
			if err != nil {
				fmt.Println(err)
			}

		}

	}
	c.Close()
	// remove tcp socket conn from list
	delete(*tcpMap, key)
}

func httpStart(port int,wbMap *map[string]*websocket.Conn,tcpMap *map[string]net.Conn) {
	http.HandleFunc(fmt.Sprint("/websocket","/",port), func(w http.ResponseWriter, r *http.Request) {
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
			fmt.Println("data from websocket ",port," >> ", time.Now().Format("[2006-01-02 15:04:05]"), " ", msgType, strings.TrimSpace(string(data)))

			dst := make([]byte, hex.DecodedLen(len(data)))
			hex.Decode(dst, data)

			// send msg to device
			for _, v := range *tcpMap {
				if v != nil {
					v.Write(dst)
				}
			}
		}

	})
}
