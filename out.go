package main

import (
	"fmt"
	"os"

	"github.com/ajstarks/svgo"
)

func printWall(enabled bool, sym string) {
	if !enabled {
		for range sym {
			fmt.Print(" ")
		}
		return
	}
	fmt.Print(sym)
}

func (m *Maze) outputAscii(display func(x, y int) string) {
	for y := 0; y < m.N; y++ {
		if y == 0 {
			for x := 0; x < m.N; x++ {
				if m.exit.y == 0 && m.exit.x == x {
					fmt.Print("-EE")
				} else if x == 0 {
					fmt.Print("--V")
				} else {
					fmt.Print("---")
				}
			}
			fmt.Print("-\n")
		} else {
			for x := 0; x < m.N; x++ {
				cross := "|"
				if x > 0 {
					cross = "+"
				}
				fmt.Print(cross)
				up := m.canPass(x, y, x, y-1)
				printWall(!up, "--")
			}
			fmt.Print("|\n")
		}
		for x := 0; x < m.N; x++ {
			if x == 0 {
				if m.exit.y == y && m.exit.x == x {
					fmt.Print("E")
				} else {
					fmt.Print("|")
				}
			} else {
				left := m.canPass(x, y, x-1, y)
				printWall(!left, "|")
			}
			fmt.Print(display(x, y))
		}
		if m.exit.y == y && m.exit.x == m.N-1 {
			fmt.Print("E\n")
		} else {
			fmt.Print("|\n")
		}
	}
	for x := 0; x < m.N; x++ {
		if m.exit.y == m.N-1 && m.exit.x == x {
			fmt.Print("-EE")
		} else {
			fmt.Print("---")
		}
	}
	fmt.Print("-\n")
}

func drawArrow(s *svg.SVG, cellx, celly, dir int) {
	// s.TranslateRotate(cellx*cellSize, celly*cellSize, 90)

	bx := cellx*cellSize + cellSize/2
	by := celly*cellSize + cellSize/2
	mx := bx + dirs[dir].x*cellSize/2
	my := by + dirs[dir].y*cellSize/2

	s.TranslateRotate(mx, my, rots[dir])

	s.Line(0, -cellSize/2, 0, cellSize/2)
	s.Line(0, cellSize/2, cellSize/6, 0)
	s.Line(0, cellSize/2, -cellSize/6, 0)

	s.Line(-cellSize/2, 0, -cellSize/3, 0)
	s.Line(-cellSize/3, cellSize/10, -cellSize/3, -cellSize/10)

	s.Line(cellSize/2, 0, cellSize/3, 0)
	s.Line(cellSize/3, cellSize/10, cellSize/3, -cellSize/10)

	s.Gend()
}

func (m *Maze) outputSVG(fname string, path []xy) error {
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("cannot open %s: %v", *outFile, err)
	}
	defer f.Close()
	s := svg.New(f)

	s.Start(cellSize*(m.N+2), cellSize*(m.N+2))
	s.Translate(cellSize, cellSize)
	s.Gstyle("stroke:black")

	if m.exit.y == 0 {
		drawArrow(s, 0, -1, dirS)
		s.Line(cellSize, 0, m.exit.x*cellSize, 0)
		drawArrow(s, m.exit.x, m.exit.y, dirN)
		s.Line((m.exit.x+1)*cellSize, 0, m.N*cellSize, 0)
	} else {
		drawArrow(s, 0, -1, dirS)
		s.Line(cellSize, 0, m.N*cellSize, 0)
	}

	if m.exit.y == m.N-1 {
		s.Line(0, m.N*cellSize, m.exit.x*cellSize, m.N*cellSize)
		drawArrow(s, m.exit.x, m.exit.y, dirS)
		s.Line((m.exit.x+1)*cellSize, m.N*cellSize, m.N*cellSize, m.N*cellSize)
	} else {
		s.Line(0, m.N*cellSize, m.N*cellSize, m.N*cellSize)
	}

	if m.exit.x == 0 && m.exit.y != 0 && m.exit.y != m.N-1 {
		s.Line(0, 0, 0, m.exit.y*cellSize)
		drawArrow(s, m.exit.x, m.exit.y, dirW)
		s.Line(0, (m.exit.y+1)*cellSize, 0, m.N*cellSize)
	} else {
		s.Line(0, 0, 0, m.N*cellSize)
	}

	if m.exit.x == m.N-1 && m.exit.y != 0 && m.exit.y != m.N-1 {
		s.Line(m.N*cellSize, 0, m.N*cellSize, m.exit.y*cellSize)
		drawArrow(s, m.exit.x, m.exit.y, dirE)
		s.Line(m.N*cellSize, (m.exit.y+1)*cellSize, m.N*cellSize, m.N*cellSize)
	} else {
		s.Line(m.N*cellSize, 0, m.N*cellSize, m.N*cellSize)
	}

	for x := 0; x < m.N; x++ {
		for y := 0; y < m.N; y++ {
			if x > 0 {
				horiz := m.canPass(x, y, x-1, y)
				if !horiz {
					s.Line(x*cellSize, y*cellSize, x*cellSize, y*cellSize+cellSize)
				}
			}
			if y > 0 {
				vert := m.canPass(x, y, x, y-1)
				if !vert {
					s.Line(x*cellSize, y*cellSize, x*cellSize+cellSize, y*cellSize)
				}
			}
		}
	}

	s.Gstyle("stroke:red")
	for i := 1; i < len(path); i++ {
		a := path[i-1]
		b := path[i]
		s.Line(a.x*cellSize+cellSize/2, a.y*cellSize+cellSize/2, b.x*cellSize+cellSize/2, b.y*cellSize+cellSize/2)
	}
	s.Gend()

	s.Gend()
	s.Gend()
	s.End()
	return nil
}

func (m *Maze) canPass(x, y, x1, y1 int) bool {
	n1 := m.node(x, y)
	n2 := m.node(x1, y1)
	if int(n1) >= len(m.G.LabeledAdjacencyList) || int(n2) >= len(m.G.LabeledAdjacencyList) {
		return false
	}
	arc, _ := m.G.HasArc(n1, n2)
	return arc
}
