package data

type Troop struct {
	Name    string `json:"Name"`
	HP      int    `json:"HP"`
	ATK     int    `json:"ATK"`
	DEF     int    `json:"DEF"`
	MANA    int    `json:"MANA"`
	EXP     int    `json:"EXP"`
	Special string `json:"Special"`
}

type Tower struct {
	Type string `json:"Type"`
	HP   int    `json:"HP"`
	ATK  int    `json:"ATK"`
	DEF  int    `json:"DEF"`
	CRIT int    `json:"CRIT"`
	EXP  int    `json:"EXP"`
}

type Player struct {
	Username string  `json:"Username"`
	Password string  `json:"Password"`
	EXP      int     `json:"EXP"`
	Level    int     `json:"Level"`
	Towers   []Tower `json:"Towers"`
	Troops   []Troop `json:"Troops"`
	Mana     int     `json:"Mana"`
}
