package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println("Server:", scanner.Text())
		}
	}()

	input := bufio.NewReader(os.Stdin)
	for {
		text, _ := input.ReadString('\n')
		fmt.Fprint(conn, text)
	}
}
