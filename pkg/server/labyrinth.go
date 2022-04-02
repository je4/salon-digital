package server

import (
	"math/rand"
)

const (
	LABLeft  = 1 << 0
	LABRight = 1 << 1
	LABTop   = 1 << 2
	LABDown  = 1 << 3
)

var Field = map[string]int{
	"cross":      LABLeft | LABRight | LABTop | LABDown,
	"downleft":   LABLeft | LABDown,
	"downright":  LABRight | LABDown,
	"edown":      LABDown,
	"eleft":      LABLeft,
	"empty":      0,
	"eright":     LABRight,
	"etop":       LABTop,
	"full":       LABLeft | LABRight | LABTop | LABDown,
	"horizontal": LABLeft | LABRight,
	"lefttop":    LABLeft | LABTop,
	"rightdown":  LABRight | LABDown,
	"tdown":      LABLeft | LABRight | LABDown,
	"tleft":      LABLeft | LABTop | LABDown,
	"topleft":    LABTop | LABLeft,
	"topright":   LABTop | LABRight,
	"tright":     LABRight | LABTop | LABDown,
	"ttop":       LABTop | LABLeft | LABRight,
	"vertical":   LABTop | LABDown,
}

type Labyrinth struct {
	Fields                   [][]string
	Detail                   string
	Size                     int
	North, East, South, West bool
}

func getField(connect, noconnect int) string {
	var list = []string{}
	for fldname, fldval := range Field {
		if fldval&connect == connect &&
			fldval&^noconnect == fldval {
			list = append(list, fldname)
		}
	}
	if len(list) == 0 {
		return ""
	}
	return list[rand.Intn(len(list))]
}

func NewLabyrinth(size int) *Labyrinth {
	lab := &Labyrinth{
		Size: size,
	}
	lab.init()
	return lab
}

func (lab *Labyrinth) init() {
	lab.Fields = [][]string{}
	for i := 0; i < lab.Size; i++ {
		lab.Fields = append(lab.Fields, []string{})
		for j := 0; j < lab.Size; j++ {
			lab.Fields[i] = append(lab.Fields[i], "")
		}
	}
	lab.fill()

}

func (lab *Labyrinth) getConnections(i, j int) (connect, noconnect int) {
	/*
		if i == 0 {
			noconnect |= LABLeft
		}
		if i == lab.Size {
			noconnect |= LABRight
		}
		if j == 0 {
			noconnect |= LABTop
		}
		if j == lab.Size {
			noconnect |= LABDown
		}
	*/
	// top
	if i > 0 {
		if lab.Fields[i-1][j] != "" {
			if Field[lab.Fields[i-1][j]]&LABDown == LABDown {
				connect |= LABTop
			} else {
				noconnect |= LABTop
			}
		}
	}
	// down
	if i < lab.Size-1 {
		if lab.Fields[i+1][j] != "" {
			if Field[lab.Fields[i+1][j]]&LABTop == LABTop {
				connect |= LABDown
			} else {
				noconnect |= LABDown
			}
		}
	}

	// left
	if j > 0 {
		if lab.Fields[i][j-1] != "" {
			if Field[lab.Fields[i][j-1]]&LABRight == LABRight {
				connect |= LABLeft
			} else {
				noconnect |= LABLeft
			}
		}
	}
	// right
	if j < lab.Size-1 {
		if lab.Fields[i][j+1] != "" {
			if Field[lab.Fields[i][j+1]]&LABLeft == LABLeft {
				connect |= LABRight
			} else {
				noconnect |= LABRight
			}
		}
	}
	return connect, noconnect
}

func (lab *Labyrinth) fill() {
	for i := 0; i < lab.Size; i++ {
		for j := 0; j < lab.Size; j++ {
			if lab.Fields[i][j] == "" {
				conn, noconn := lab.getConnections(i, j)
				lab.Fields[i][j] = getField(conn, noconn)
			}
		}
	}
	lab.West, lab.East, lab.North, lab.South = lab.GetMovements()
}

func (lab *Labyrinth) GetMovements() (left, right, top, down bool) {
	middle := lab.Size / 2
	dir, _ := lab.getConnections(middle, middle)
	left = dir&LABLeft > 0
	right = dir&LABRight > 0
	top = dir&LABTop > 0
	down = dir&LABDown > 0
	return
}

func (lab *Labyrinth) Move(x, y int) {
	for i := 0; i < lab.Size; i++ {
		i0 := i
		if y > 0 {
			i0 = lab.Size - 1 - i
		}
		i1 := i0 + y
		for j := 0; j < lab.Size; j++ {
			j0 := j
			if x > 0 {
				j0 = lab.Size - 1 - j
			}
			j1 := j0 + x
			if j1 < lab.Size && j1 >= 0 &&
				i1 < lab.Size && i1 >= 0 {
				lab.Fields[i1][j1] = lab.Fields[i0][j0]
			}
			lab.Fields[i0][j0] = ""
		}
	}
	lab.fill()
}
