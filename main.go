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

	for {
		buf := make([]byte, 1024)

		lengthPer, err := bufio.NewReader(c).Read(buf)
		buf = buf[:lengthPer]
		temp := strings.TrimSpace(string(buf))
		fmt.Println("hex==> ", hex.EncodeToString(buf))
		fmt.Println("string ==>>", temp)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	c.Close()
}
