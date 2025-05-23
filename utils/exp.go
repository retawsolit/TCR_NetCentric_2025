package utils

import (
	"math"
	"tcr/data"
)

// Cộng EXP cho player và kiểm tra lên cấp
func GainEXP(p *data.Player, amount int) (leveledUp bool) {
	p.EXP += amount
	required := p.Level * 100
	if p.EXP >= required {
		p.Level++
		p.EXP = 0
		BuffPlayerStats(p)
		return true
	}
	return false
}

// Tăng chỉ số troops + towers 10% mỗi cấp
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
