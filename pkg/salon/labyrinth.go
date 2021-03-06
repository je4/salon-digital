package salon

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

type FieldEntry struct {
	Name string
	Work *Work
}

type Labyrinth struct {
	Fields                   [][]FieldEntry
	Detail                   string
	Size                     int
	North, East, South, West bool
	works                    map[string]*Work
	signatures               []string
}

func NewLabyrinth(size int, works map[string]*Work) *Labyrinth {
	lab := &Labyrinth{
		Size: size,
	}
	lab.initData(works)
	lab.init()
	return lab
}

func (lab *Labyrinth) initData(works map[string]*Work) {
	lab.works = works
	lab.signatures = []string{}
	for sig, _ := range lab.works {
		lab.signatures = append(lab.signatures, sig)
	}
}
func (lab *Labyrinth) init() {
	lab.Fields = [][]FieldEntry{}
	for i := 0; i < lab.Size; i++ {
		lab.Fields = append(lab.Fields, []FieldEntry{})
		for j := 0; j < lab.Size; j++ {
			lab.Fields[i] = append(lab.Fields[i], FieldEntry{})
		}
	}
	lab.fill()

}

func (lab *Labyrinth) getField(connect, noconnect int) FieldEntry {
	var list = []string{}
	for fldname, fldval := range Field {
		if fldval&connect == connect &&
			fldval&^noconnect == fldval {
			list = append(list, fldname)
		}
	}
	if len(list) == 0 {
		return FieldEntry{}
	}
	fe := FieldEntry{
		Name: list[rand.Intn(len(list))],
		Work: nil,
	}
	if fe.Name == "full" {
		fe.Work = lab.works[lab.signatures[rand.Intn(len(lab.signatures))]]
	}
	return fe
}

func (lab *Labyrinth) getConnections(i, j int) (connect, noconnect int) {
	// top
	if i > 0 {
		if lab.Fields[i-1][j].Name != "" {
			if Field[lab.Fields[i-1][j].Name]&LABDown == LABDown {
				connect |= LABTop
			} else {
				noconnect |= LABTop
			}
		}
	}
	// down
	if i < lab.Size-1 {
		if lab.Fields[i+1][j].Name != "" {
			if Field[lab.Fields[i+1][j].Name]&LABTop == LABTop {
				connect |= LABDown
			} else {
				noconnect |= LABDown
			}
		}
	}

	// left
	if j > 0 {
		if lab.Fields[i][j-1].Name != "" {
			if Field[lab.Fields[i][j-1].Name]&LABRight == LABRight {
				connect |= LABLeft
			} else {
				noconnect |= LABLeft
			}
		}
	}
	// right
	if j < lab.Size-1 {
		if lab.Fields[i][j+1].Name != "" {
			if Field[lab.Fields[i][j+1].Name]&LABLeft == LABLeft {
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
			if lab.Fields[i][j].Name == "" {
				conn, noconn := lab.getConnections(i, j)
				lab.Fields[i][j] = lab.getField(conn, noconn)
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
			lab.Fields[i0][j0] = FieldEntry{}
		}
	}
	lab.fill()
}
