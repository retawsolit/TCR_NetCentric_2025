package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"tcr/data"
	"tcr/utils"
	"time"
)

var playerConns [2]net.Conn
var players [2]data.Player
var gameOver atomic.Bool

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("🟢 Server is listening on port 8080...")

	var wg sync.WaitGroup
	wg.Add(2)

	for i := 0; i < 2; i++ {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		playerConns[i] = conn
		go handleLogin(conn, i, &wg)
	}

	wg.Wait()
	fmt.Println("✅ Both players connected. Starting the game...\n")
	go handleGame(0)
	go handleGame(1)

	select {}
}

func handleLogin(conn net.Conn, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	reader := bufio.NewReader(conn)
	conn.Write([]byte(fmt.Sprintf("Welcome Player %d! Please enter your username:\n", id+1)))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	conn.Write([]byte("Please enter your password:\n"))
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	dataBytes, err := os.ReadFile("data/players.json")
	if err != nil {
		conn.Write([]byte("Server error loading data\n"))
		return
	}

	var all []data.Player
	if err := json.Unmarshal(dataBytes, &all); err != nil {
		conn.Write([]byte("Error parsing player data\n"))
		return
	}

	for _, p := range all {
		if p.Username == username && p.Password == password {
			// Tăng lại HP tương ứng level
			for i := range p.Towers {
				baseHP := 1000 // hoặc giá trị gốc tùy tower type
				if p.Towers[i].Type == "King Tower" {
					baseHP = 2000
				}

				finalHP := float64(baseHP)
				for l := 1; l < p.Level; l++ {
					finalHP *= 1.1
				}

				// Nếu tower HP ≤ 0 → reset lại tương ứng level
				if p.Towers[i].HP <= 0 {
					p.Towers[i].HP = int(math.Round(finalHP))
				}
			}

			players[id] = p
			startManaRegen(&players[id])
			conn.Write([]byte("✅ Login success!\n"))
			fmt.Printf("✅ Player %d (%s) logged in.\n", id+1, username)
			return
		}
	}

	conn.Write([]byte("❌ Login failed.\n"))
	fmt.Printf("❌ Login failed for %s\n", username)
}

func handleGame(id int) {
	conn := playerConns[id]
	reader := bufio.NewReader(conn)

	player := &players[id]
	enemy := &players[1-id]
	targetConn := playerConns[1-id]

	for {
		if gameOver.Load() {
			return
		}

		// 1. Hiển thị lựa chọn troop
		conn.Write([]byte("🧠 Your turn! Type a troop name to deploy:\n"))
		for _, t := range player.Troops {
			conn.Write([]byte(fmt.Sprintf("- %s (ATK: %d, MANA: %d)\n", t.Name, t.ATK, t.MANA)))
		}
		conn.Write([]byte(fmt.Sprintf("Your current MANA: %d\n", player.Mana)))

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if gameOver.Load() {
			conn.Write([]byte("❌ Game already ended.\n"))
			return
		}

		var chosen *data.Troop
		for i := range player.Troops {
			if strings.EqualFold(strings.TrimSpace(player.Troops[i].Name), text) {
				chosen = &player.Troops[i]
				break
			}
		}
		if chosen == nil {
			conn.Write([]byte("❌ Invalid troop name.\n"))
			continue
		}

		if player.Mana < chosen.MANA {
			conn.Write([]byte("❌ Not enough mana to deploy this troop.\n"))
			continue
		}

		// 2. Hiển thị danh sách tower còn sống hoặc đã bị phá
		conn.Write([]byte("Choose a tower to attack:\n"))
		for i, t := range enemy.Towers {
			status := fmt.Sprintf("HP: %d", t.HP)
			if t.HP <= 0 {
				status = "DESTROYED ❌"
			}
			conn.Write([]byte(fmt.Sprintf("[%d] %s (%s)\n", i, t.Type, status)))
		}
		conn.Write([]byte("Enter tower index (e.g., 0 for King, 1 for Guard1...):\n"))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		towerIdx, err := strconv.Atoi(input)
		if err != nil || towerIdx < 0 || towerIdx >= len(enemy.Towers) {
			conn.Write([]byte("❌ Invalid tower index.\n"))
			continue
		}

		// Chặn chọn tower đã bị phá
		if enemy.Towers[towerIdx].HP <= 0 {
			conn.Write([]byte("❌ This tower has already been destroyed.\n"))
			continue
		}

		// Chặn đánh tower 2 nếu tower 1 chưa bị phá
		if towerIdx == 2 && enemy.Towers[1].HP > 0 {
			conn.Write([]byte("❌ Must destroy Guard Tower 1 before Guard Tower 2.\n"))
			continue
		}

		if towerIdx == 0 && (enemy.Towers[1].HP > 0 || enemy.Towers[2].HP > 0) {
			conn.Write([]byte("❌ Must destroy both Guard Towers before attacking King Tower.\n"))
			continue
		}

		// Giao tranh
		player.Mana -= chosen.MANA
		if GainEXP(player, chosen.EXP) {
			conn.Write([]byte(fmt.Sprintf("🎉 Level UP! You are now Level %d\n", player.Level)))
		}

		tower := &enemy.Towers[towerIdx]
		damage := utils.AttackTower(chosen, tower, id)

		msg := fmt.Sprintf("🔥 Player %d used %s. Dealt %d damage to Player %d's %s. HP left: %d\n",
			id+1, chosen.Name, damage, (1-id)+1, tower.Type, tower.HP)

		conn.Write([]byte(msg))
		targetConn.Write([]byte(msg))

		// Kiểm tra nếu tower bị phá
		if tower.HP <= 0 {
			notify := fmt.Sprintf("🎯 Player %d (%s) destroyed %s!\n", id+1, player.Username, tower.Type)
			playerConns[0].Write([]byte(notify))
			playerConns[1].Write([]byte(notify))
		}

		// Kiểm tra end game khi King Tower bị phá
		if enemy.Towers[0].HP <= 0 {
			winMsg := fmt.Sprintf("🎉 Player %d (%s) wins the game!\n", id+1, player.Username)
			conn.Write([]byte(winMsg))
			targetConn.Write([]byte(winMsg))

			if GainEXP(player, tower.EXP) {
				conn.Write([]byte(fmt.Sprintf("🎉 Level UP! You are now Level %d\n", player.Level)))
			}

			// Ghi lại trạng thái
			allPlayers := []data.Player{players[0], players[1]}
			dataBytes, err := json.MarshalIndent(allPlayers, "", "  ")
			if err == nil {
				_ = os.WriteFile("data/players.json", dataBytes, 0644)
				fmt.Println("✅ Saved players.json with updated EXP and Level.")
			}

			gameOver.Store(true)
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func GainEXP(p *data.Player, gained int) bool {
	p.EXP += gained
	required := int(100 + float64(p.Level)*10)
	if p.EXP >= required {
		p.Level++
		p.EXP = 0
		BuffPlayerStats(p)
		return true
	}
	return false
}

func BuffPlayerStats(p *data.Player) {
	for i := range p.Troops {
		p.Troops[i].HP = int(math.Round(float64(p.Troops[i].HP) * 1.1))
		p.Troops[i].ATK = int(math.Round(float64(p.Troops[i].ATK) * 1.1))
		p.Troops[i].DEF = int(math.Round(float64(p.Troops[i].DEF) * 1.1))
	}

	for i := range p.Towers {
		p.Towers[i].HP = int(math.Round(float64(p.Towers[i].HP) * 1.1))
		p.Towers[i].ATK = int(math.Round(float64(p.Towers[i].ATK) * 1.1))
		p.Towers[i].DEF = int(math.Round(float64(p.Towers[i].DEF) * 1.1))
	}
}

func startManaRegen(player *data.Player) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			if gameOver.Load() {
				return
			}
			if player.Mana < 10 {
				player.Mana++
			}
		}
	}()
}
