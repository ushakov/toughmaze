package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/soniakeys/graph"
)

var (
	outFile     = flag.String("out", "", "File to write svg ouput (if unset, output ascii art to stdout)")
	outSolution = flag.String("solution", "", "File to write solution svg")
)

const cellSize = 20
const N = 30

type xy struct {
	x int
	y int
}

func (p xy) String() string {
	return fmt.Sprintf("(%d, %d)", p.x, p.y)
}

var (
	dirs = []xy{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
	}

	weights = []int{
		1,
		5,
		1,
		5,
	}

	rots = []float64{
		0,
		-90,
		180,
		90,
	}
)

const (
	dirS = iota
	dirE
	dirN
	dirW
)

type Maze struct {
	N          int
	G          graph.LabeledUndirected
	visited    map[xy]bool
	extendable map[xy]bool
	path       []xy
	dists      []float64
	exit       xy
}

func NewMaze(size int) *Maze {
	m := &Maze{
		N:       size,
		G:       graph.LabeledUndirected{},
		visited: make(map[xy]bool),
		path:    []xy{{0, 0}},
	}
	return m
}

func (m *Maze) neighbors(x, y int) (ret []xy) {
	for i, d := range dirs {
		xx := x + d.x
		yy := y + d.y
		if xx >= 0 && xx < m.N && yy >= 0 && yy < m.N {
			for q := 0; q < weights[i]; q++ {
				ret = append(ret, xy{xx, yy})
			}
		}
	}
	return
}

func (m *Maze) coordinates(node graph.NI) (x, y int) {
	num := int(node)
	return num / m.N, num % m.N
}

func (m *Maze) node(x, y int) graph.NI {
	return graph.NI(x*m.N + y)
}

func (m *Maze) score() float64 {
	m.computeDistances()
	if m.dists == nil {
		return 0
	}
	var score float64
	for x := 0; x < m.N; x++ {
		for y := 0; y < m.N; y++ {
			n := m.node(x, y)
			dist := m.dists[n]
			localMaximum := true
			for _, neigh := range m.neighbors(x, y) {
				nn := m.node(neigh.x, neigh.y)
				if m.dists[nn] > dist {
					localMaximum = false
				}
			}
			if localMaximum {
				score += dist
			}
		}
	}
	return score
}

func (m *Maze) removeWall(from, to xy) {
	n1 := m.node(from.x, from.y)
	n2 := m.node(to.x, to.y)
	m.G.AddEdge(graph.Edge{n1, n2}, 0)
}

func (m *Maze) selectNext(current xy) (next xy, num_neighbors int) {
	var candidates []xy
	for _, n := range m.neighbors(current.x, current.y) {
		if !m.visited[n] {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return xy{}, 0
	}
	return candidates[rand.Intn(len(candidates))], len(candidates)
}

func (m *Maze) extend(current xy) bool {
	next, n := m.selectNext(current)
	if n == 0 {
		return false
	}
	m.removeWall(current, next)
	m.visited[next] = true
	m.path = append(m.path, next)
	return true
}

func (m *Maze) createMaze() {
	m.visited = make(map[xy]bool)
	m.path = []xy{{0, 0}}
	m.visited[m.path[0]] = true
	num := 0
	for len(m.path) > 0 {
		ok := m.extend(m.path[len(m.path)-1])
		num++
		if !ok {
			m.path = m.path[:len(m.path)-1]
		}
	}

	m.selectExit()
}

func (m *Maze) extendLine(p xy) int {
	num := 0
	for {
		nextp, n := m.selectNext(p)
		if n == 0 {
			m.extendable[p] = false
			return num
		}
		m.removeWall(p, nextp)
		p = nextp
		m.visited[p] = true
		m.extendable[p] = n > 1
		num++
	}
}

func randomKey(m map[xy]bool) (key xy, num int) {
	var r []xy
	for p, v := range m {
		if v {
			r = append(r, p)
		}
	}
	if len(r) == 0 {
		return xy{}, 0
	}
	return r[rand.Intn(len(r))], len(r)
}

func (m *Maze) createMaze3() {
	m.visited = map[xy]bool{}
	m.visited[xy{0, 0}] = true
	m.extendable = map[xy]bool{{0, 0}: true}
	iter := 0
	for {
		seed, nExtendable := randomKey(m.extendable)
		if nExtendable == 0 {
			break
		}
		ext := m.extendLine(seed)
		if ext == 0 {
			m.extendable[seed] = false
		}
		iter++
		log.Println("iter", iter, "seed =", seed, "extendables =", nExtendable, "extended for", ext)
	}

	m.selectExit()
}

func (m *Maze) createMaze2() {
	m.visited = make(map[xy]bool)
	gray := []xy{{0, 0}}
	iter := 0
	for len(gray) > 0 {
		isel := rand.Intn(len(gray))
		sel := gray[isel]
		gray[isel] = gray[len(gray)-1]
		gray = gray[0 : len(gray)-1]

		var visitedNeighbors []xy
		var unvisitedNeighbors []xy
		for _, n := range m.neighbors(sel.x, sel.y) {
			if m.visited[n] {
				visitedNeighbors = append(visitedNeighbors, n)
			} else {
				unvisitedNeighbors = append(unvisitedNeighbors, n)
			}
		}
		if len(visitedNeighbors) > 0 {
			conn := visitedNeighbors[rand.Intn(len(visitedNeighbors))]
			m.removeWall(sel, conn)
		}
		m.visited[sel] = true

		nextgray := map[xy]struct{}{}
		for _, p := range append(gray, unvisitedNeighbors...) {
			nextgray[p] = struct{}{}
		}
		gray = nil
		for p := range nextgray {
			gray = append(gray, p)
		}
		iter++
		//log.Println("iteration", iter, "-- gray=", gray)
	}

	m.selectExit()
}

func (m *Maze) selectExit() {
	m.computeDistances()
	bestExit := xy{0, 0}
	bestDist := 0

	tryExit := func(x, y int) {
		node := m.node(x, y)
		if int(node) >= len(m.dists) {
			return
		}
		d := int(m.dists[m.node(x, y)])
		if d > bestDist {
			bestDist = d
			bestExit = xy{x, y}
		}
	}

	for i := 0; i < m.N; i++ {
		tryExit(0, i)
		tryExit(i, 0)
		tryExit(m.N-1, i)
		tryExit(i, m.N-1)
	}
	m.exit = bestExit
}

func (m *Maze) computeDistances() {
	if m.dists != nil {
		return
	}
	w := func(label graph.LI) float64 { return 1 }
	start := m.node(0, 0)
	if int(start) >= len(m.G.LabeledAdjacencyList) {
		return
	}
	_, _, m.dists, _ = m.G.Dijkstra(start, -1, w)
}

func (m *Maze) solve() (path []xy) {
	w := func(label graph.LI) float64 { return 1 }
	exit := m.node(m.exit.x, m.exit.y)
	f, _, _, _ := m.G.Dijkstra(m.node(0, 0), exit, w)
	nodes := f.PathTo(exit, nil)
	for _, n := range nodes {
		x, y := m.coordinates(n)
		path = append(path, xy{x, y})
	}
	return
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().Unix())
	m := NewMaze(N)
	m.createMaze3()
	log.Printf("Score: %v", m.score())
	if *outFile == "" {
		m.computeDistances()
		end := m.node(m.exit.x, m.exit.y)
		w := func(label graph.LI) float64 { return 1 }
		start := m.node(0, 0)
		f, _, _, _ := m.G.Dijkstra(start, end, w)
		path := f.PathTo(end, nil)
		onpath := map[xy]struct{}{}
		for _, n := range path {
			x, y := m.coordinates(n)
			onpath[xy{x, y}] = struct{}{}
		}
		display := func(x, y int) string {
			_, ok := onpath[xy{x, y}]
			if ok {
				return "++"
			} else {
				return "  "
			}
		}
		// display = func(x, y int) string {
		// 	return "  "
		// }
		m.outputAscii(display)
	} else {
		err := m.outputSVG(*outFile, nil)
		if err != nil {
			log.Panicf("Writing SVG file: %v", err)
		}

		if *outSolution != "" {
			path := m.solve()
			err := m.outputSVG(*outSolution, path)
			if err != nil {
				log.Panicf("Writing solution SVG file: %v", err)
			}
		}
	}
}
