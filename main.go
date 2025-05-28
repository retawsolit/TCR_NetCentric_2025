package main

import (
	"fmt"
	"log"
	"tcr/data"
)

func main() {
	troops, err := data.LoadTroops("data/troops.json")
	if err != nil {
		log.Fatal(err)
	}

	towers, err := data.LoadTowers("data/towers.json")
	if err != nil {
		log.Fatal(err)
	}

	player, err := data.LoadPlayer("data/players.json")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loaded", len(troops), "troops")
	fmt.Println("Loaded", len(towers), "towers")
	fmt.Println("Player:", player.Username, "Level:", player.Level, "Mana:", player.Mana)

	//DEBUG TEST COMBAT
	// // Lấy Knight và Guard Tower để test
	// var knight *data.Troop
	// var guard *data.Tower
	// for i := range player.Troops {
	// 	if player.Troops[i].Name == "Knight" {
	// 		knight = &player.Troops[i]
	// 		break
	// 	}
	// }
	// for i := range player.Towers {
	// 	if player.Towers[i].Type == "Guard Tower" {
	// 		guard = &player.Towers[i]
	// 		break
	// 	}
	// }

	// // Test combat
	// if knight != nil && guard != nil {
	// 	fmt.Println("Before attack - Guard Tower HP:", guard.HP)
	// 	dmg := utils.AttackTower(knight, guard)
	// 	fmt.Println("Knight attacked Guard Tower!")
	// 	fmt.Println("Damage dealt:", dmg)
	// 	fmt.Println("After attack - Guard Tower HP:", guard.HP)
	// } else {
	// 	fmt.Println("Knight or Guard Tower not found!")
	// }
}
