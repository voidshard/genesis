package types

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Dice represents a set of random numbers + a bonus min number.
// Making a range of random values with uneven distribution
// and a minumum value.
type Dice struct {
	roll []int
	add  int
}

// NewDice returns a new Dice struct.
// Ie.
// - 2d6 + 1d8 + 5
// - roll two six-sided dice, add one eight sided die then add five
// Would be NewDice(5, 6, 6, 8)
func NewDice(add int, rolls ...int) *Dice {
	accept := []int{}
	for _, r := range rolls {
		if r > 0 {
			accept = append(accept, r)
		}
	}
	return &Dice{
		roll: accept,
		add:  add,
	}
}

// Roll the dice!
func (d *Dice) Roll() int {
	v := d.add
	for _, x := range d.roll {
		v += rand.Intn(x)
	}
	return v
}
