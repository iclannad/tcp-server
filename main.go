package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
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

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}

}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	data := make([]byte, 0)
	for {
		buf := make([]byte, 1024)
		lengthPer, err := bufio.NewReader(c).Read(buf)
		if err != nil {
			data = append(data, buf[0:lengthPer]...)
			fmt.Println(err)
			break
		}
		data = append(data, buf[0:lengthPer]...)
	}
	temp := strings.TrimSpace(string(data))
	fmt.Println(time.Now().Format("[2006-01-02 15:04:05]"))
	fmt.Println("hex==> ", hex.EncodeToString(data))
	fmt.Println("string ==>>", temp)

	c.Close()
}
