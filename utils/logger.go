package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"tcr/data"
	"time"
)

// 📝 Ghi toàn bộ log trận đấu vào file logs/game_timestamp.txt
func WriteLogs(logs []string) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("logs/game_%s.txt", timestamp)
	os.MkdirAll("logs", 0755)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("❌ Error creating log file:", err)
		return
	}
	defer file.Close()

	for _, line := range logs {
		file.WriteString(line + "\n")
	}

	fmt.Printf("✅ Full match log saved to %s\n", filename)
}

// 💾 Ghi trạng thái player sau trận vào JSON
func SavePlayersToJSON(players []data.Player) {
	dataBytes, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		fmt.Println("❌ Failed to marshal JSON:", err)
		return
	}

	err = os.WriteFile("data/players.json", dataBytes, 0644)
	if err != nil {
		fmt.Println("❌ Failed to write players.json:", err)
	} else {
		fmt.Println("✅ Saved players.json with updated EXP, Level, and HP.")
	}
}
