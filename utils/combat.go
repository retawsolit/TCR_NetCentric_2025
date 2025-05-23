package utils

import (
	"fmt"
	"math/rand"
	"tcr/data"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CRIT theo phần trăm cho Tower vs Troop
func CalculateDamage(attackerATK, critChance, defenderDEF int) int {
	crit := rand.Intn(100) < critChance // true if CRIT
	actualATK := attackerATK
	if crit {
		actualATK = int(float64(attackerATK) * 1.2)
	}
	dmg := actualATK - defenderDEF
	if dmg < 0 {
		return 0
	}
	return dmg
}

// Đếm lượt đánh để tính CRIT theo lượt
var attackCount = make(map[string]int)

// Troop đánh vào Tower (CRIT mỗi 3 đòn)
func AttackTower(troop *data.Troop, tower *data.Tower, playerID int) int {
	key := fmt.Sprintf("%d_%s", playerID, tower.Type)
	attackCount[key]++

	damage := troop.ATK - tower.DEF
	if damage < 0 {
		damage = 0
	}

	// Every 3rd attack is CRIT (×2 damage)
	if attackCount[key]%3 == 0 {
		damage *= 2
	}

	tower.HP -= damage
	if tower.HP < 0 {
		tower.HP = 0
	}

	return damage
}

// Tower phản công troop (có CRIT theo phần trăm)
func AttackTroop(tower *data.Tower, troop *data.Troop) int {
	damage := CalculateDamage(tower.ATK, tower.CRIT, troop.DEF)
	troop.HP -= damage
	if troop.HP < 0 {
		troop.HP = 0
	}
	return damage
}
