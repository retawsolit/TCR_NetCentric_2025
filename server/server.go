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
var logs [2][]string

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

	// ✅ Load danh sách tổng troop
	allTroops, err := data.LoadTroops("data/troops.json")
	if err != nil {
		conn.Write([]byte("Error loading troop data\n"))
		return
	}

	for _, p := range all {
		if p.Username == username && p.Password == password {
			// ✅ Random chọn 3 troop từ danh sách tổng
			p.Troops = data.PickRandomTroops(allTroops, 3)

			// Tăng lại HP tương ứng level
			for i := range p.Towers {
				baseHP := 1000
				if p.Towers[i].Type == "King Tower" {
					baseHP = 2000
				}

				finalHP := float64(baseHP)
				for l := 1; l < p.Level; l++ {
					finalHP *= 1.1
				}

				p.Towers[i].HP = int(math.Round(finalHP))
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

	startTime := time.Now()
	troopUsage := map[string]int{}
	totalDamage := 0

	for {
		// ⏱️ Time up
		if time.Since(startTime) >= 3*time.Minute {
			logs[id] = append(logs[id], "⏰ Time's up!")

			playerAlive, enemyAlive := 0, 0
			for _, t := range player.Towers {
				if t.HP > 0 {
					playerAlive++
				}
			}
			for _, t := range enemy.Towers {
				if t.HP > 0 {
					enemyAlive++
				}
			}

			if playerAlive > enemyAlive {
				conn.Write([]byte("🏆 You win by having more towers!\n"))
				targetConn.Write([]byte("❌ You lose. Opponent has more towers.\n"))
				GainEXP(player, 30)
			} else if enemyAlive > playerAlive {
				conn.Write([]byte("❌ You lose. Opponent has more towers.\n"))
				targetConn.Write([]byte("🏆 You win by having more towers!\n"))
				GainEXP(enemy, 30)
			} else {
				conn.Write([]byte("🤝 Draw! Equal number of towers.\n"))
				targetConn.Write([]byte("🤝 Draw! Equal number of towers.\n"))
				GainEXP(player, 10)
				GainEXP(enemy, 10)
			}

			// ✅ Gộp và ghi log
			fullLog := append(logs[0], logs[1]...)
			fullLog = append(fullLog, fmt.Sprintf("📊 Player %d Total Damage: %d", id+1, totalDamage))
			for troop, count := range troopUsage {
				fullLog = append(fullLog, fmt.Sprintf("🧍 %s used %d time(s)", troop, count))
			}
			utils.WriteLogs(fullLog)

			utils.SavePlayersToJSON([]data.Player{players[0], players[1]})
			gameOver.Store(true)
			return
		}

		if gameOver.Load() {
			return
		}

		// ✏️ Chọn troop
		conn.Write([]byte("🧠 Your turn! Type a troop name to deploy:\n"))
		for _, t := range player.Troops {
			conn.Write([]byte(fmt.Sprintf("- %s (ATK: %d, MANA: %d)\n", t.Name, t.ATK, t.MANA)))
		}
		manaBar := strings.Repeat("|", player.Mana) + strings.Repeat(".", 10-player.Mana)
		conn.Write([]byte(fmt.Sprintf("💧 Mana: [%s] (%d/10)\n", manaBar, player.Mana)))

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if gameOver.Load() {
			conn.Write([]byte("❌ Game already ended.\n"))
			return
		}

		var chosen *data.Troop
		for i := range player.Troops {
			if strings.EqualFold(player.Troops[i].Name, text) {
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

		player.Mana -= chosen.MANA
		if GainEXP(player, chosen.EXP) {
			conn.Write([]byte(fmt.Sprintf("🎉 Level UP! You are now Level %d\n", player.Level)))
		}

		// 👉 Nếu Queen → heal phe mình
		if strings.EqualFold(chosen.Name, "Queen") {
			conn.Write([]byte("💖 Select your tower to heal:\n"))
			for idx, t := range player.Towers {
				bar := strings.Repeat("█", t.HP/200)
				conn.Write([]byte(fmt.Sprintf("[%d] %-14s | HP: %4d | %s\n", idx, t.Type, t.HP, bar)))
			}
			conn.Write([]byte("Enter tower index to heal:\n"))
			towerIdxStr, _ := reader.ReadString('\n')
			towerIdxStr = strings.TrimSpace(towerIdxStr)
			towerIdx, err := strconv.Atoi(towerIdxStr)
			if err != nil || towerIdx < 0 || towerIdx >= len(player.Towers) {
				conn.Write([]byte("❌ Invalid tower index.\n"))
				continue
			}
			healTarget := &player.Towers[towerIdx]
			healTarget.HP += 300
			maxHP := utils.GetTowerMaxHPScaled(healTarget, player.Level)
			if healTarget.HP > maxHP {
				healTarget.HP = maxHP
			}
			conn.Write([]byte(fmt.Sprintf("💖 Healed %s by 300 HP. New HP: %d\n", healTarget.Type, healTarget.HP)))
			continue // Skip attack phase
		}

		// 📍Chọn enemy tower
		conn.Write([]byte("🛡️  Enemy Towers Status:\n"))
		for idx, t := range enemy.Towers {
			bar := strings.Repeat("█", t.HP/200)
			if t.HP <= 0 {
				conn.Write([]byte(fmt.Sprintf("[%d] [X] %-14s | DESTROYED ❌\n", idx, t.Type)))
			} else {
				conn.Write([]byte(fmt.Sprintf("[%d]     %-14s | HP: %4d | %s\n", idx, t.Type, t.HP, bar)))
			}
		}

		conn.Write([]byte("Enter tower index (0 = King, 1 = Guard1...):\n"))
		towerIdxStr, _ := reader.ReadString('\n')
		towerIdxStr = strings.TrimSpace(towerIdxStr)
		towerIdx, err := strconv.Atoi(towerIdxStr)
		if err != nil || towerIdx < 0 || towerIdx >= len(enemy.Towers) {
			conn.Write([]byte("❌ Invalid tower index.\n"))
			continue
		}
		targetTower := &enemy.Towers[towerIdx]
		if targetTower.HP <= 0 {
			conn.Write([]byte("❌ This tower has already been destroyed.\n"))
			continue
		}

		// 🔒 Check attack order restriction
		if (targetTower.Type == "Guard Tower 2" || targetTower.Type == "King Tower") && enemy.Towers[1].HP > 0 {
			conn.Write([]byte("🚫 Cannot attack this tower until Guard Tower 1 is destroyed.\n"))
			continue
		}

		// 🚀 Tấn công
		damage := utils.AttackTower(chosen, targetTower, id, enemy)
		troopUsage[chosen.Name]++
		totalDamage += damage
		msg := fmt.Sprintf("🔥 Player %d used %s. Dealt %d damage to Player %d's %s. HP left: %d",
			id+1, chosen.Name, damage, (1-id)+1, targetTower.Type, targetTower.HP)

		conn.Write([]byte(msg + "\n"))
		targetConn.Write([]byte(msg + "\n"))
		logs[id] = append(logs[id], msg)

		// 🎯 Kiểm tra King Tower
		if strings.Contains(targetTower.Type, "King") && targetTower.HP <= 0 {
			winMsg := fmt.Sprintf("🎉 Player %d (%s) wins by destroying the King Tower!", id+1, player.Username)
			conn.Write([]byte(winMsg + "\n"))
			targetConn.Write([]byte(winMsg + "\n"))
			logs[id] = append(logs[id], winMsg)

			GainEXP(player, 30)
			fullLog := append(logs[0], logs[1]...)
			fullLog = append(fullLog, fmt.Sprintf("📊 Player %d Total Damage: %d", id+1, totalDamage))
			for troop, count := range troopUsage {
				fullLog = append(fullLog, fmt.Sprintf("🧍 %s used %d time(s)", troop, count))
			}
			utils.WriteLogs(fullLog)
			utils.SavePlayersToJSON([]data.Player{players[0], players[1]})
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
		defer ticker.Stop() // ✅ Dừng ticker khi goroutine kết thúc
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
