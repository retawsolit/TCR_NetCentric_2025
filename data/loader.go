package data

import (
	"encoding/json"
	"os"
)

func LoadTroops(path string) ([]Troop, error) {
	var troops []Troop
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &troops); err != nil {
		return nil, err
	}
	return troops, nil
}

func LoadTowers(path string) ([]Tower, error) {
	var towers []Tower
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &towers); err != nil {
		return nil, err
	}
	return towers, nil
}

func LoadPlayer(path string) (*Player, error) {
	var player Player
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &player); err != nil {
		return nil, err
	}
	return &player, nil
}
