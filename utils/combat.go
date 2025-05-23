package utils

import (
	"fmt"
	"math"
	"strings"
	"tcr/data"
)

var attackCount = make(map[string]int)

// 🧠 Troop attack logic including CRIT, special skills
func AttackTower(troop *data.Troop, tower *data.Tower, playerID int, enemy *data.Player) int {
	key := fmt.Sprintf("%d_%s", playerID, tower.Type)
	attackCount[key]++

	damage := troop.ATK - tower.DEF
	if damage < 0 {
		damage = 0
	}

	// 🎯 CRIT every 3rd attack
	if attackCount[key]%3 == 0 {
		damage *= 2
	}

	// 🧱 Rook: bypass DEF
	if troop.Name == "Rook" {
		damage = troop.ATK
	}

	// 👑 Prince: CRIT if tower HP < 20% of scaled HP
	if troop.Name == "Prince" && float64(tower.HP) < 0.2*float64(getTowerMaxHPScaled(tower, enemy.Level)) {
		damage *= 2
	}

	// ⚔️ Pawn: damage x2 to King if Guard1 & Guard2 destroyed
	if troop.Name == "Pawn" && strings.Contains(tower.Type, "King") && isGuardTowerDestroyed(enemy) {
		damage *= 2
	}

	tower.HP -= damage
	if tower.HP < 0 {
		tower.HP = 0
	}

	return damage
}

// 📐 Get max HP of tower scaled by level
func getTowerMaxHPScaled(t *data.Tower, level int) int {
	base := 1000
	if t.Type == "King Tower" {
		base = 2000
	}
	return int(float64(base) * math.Pow(1.1, float64(level-1)))
}

// 🔎 Check if both Guard Towers are destroyed
func isGuardTowerDestroyed(p *data.Player) bool {
	count := 0
	for _, t := range p.Towers {
		if strings.Contains(t.Type, "Guard") && t.HP <= 0 {
			count++
		}
	}
	return count >= 2
}
