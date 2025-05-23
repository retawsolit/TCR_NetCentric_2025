package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Print("📂 Enter path to log file: ")

	reader := bufio.NewReader(os.Stdin)
	path, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("❌ Error reading input:", err)
		return
	}
	path = strings.TrimSpace(path) // giữ nguyên path có khoảng trắng

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("❌ Cannot open log file:", err)
		return
	}
	defer file.Close()

	fmt.Println("🎬 Starting replay...\n")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		time.Sleep(700 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("❌ Error reading log file:", err)
		return
	}

	fmt.Println("\n🏁 Replay ended.")
}
