package game

import (
	"container/heap"
	"fmt"
	"math"
	"oinakos/internal/engine"
)

type PathNode struct {
	X, Y     int
	G, H, F  float64
	Parent   *PathNode
	index    int
}

type PriorityQueue []*PathNode

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].F < pq[j].F }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index, pq[j].index = i, j }
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*PathNode)
	node.index = n
	*pq = append(*pq, node)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	node.index = -1
	*pq = old[0 : n-1]
	return node
}

func (a *Actor) FindPath(targetX, targetY float64, ctx *SystemContext) []engine.Point {
	if ctx == nil || ctx.World == nil || ctx.World.Game == nil { return nil }
	return ctx.World.Game.FindAStarPath(a.X, a.Y, targetX, targetY)
}

func (g *Game) FindAStarPath(startX, startY, endX, endY float64) []engine.Point {
	gridSize := 1.0 
	sX, sY := int(math.Floor(startX/gridSize)), int(math.Floor(startY/gridSize))
	eX, eY := int(math.Floor(endX/gridSize)), int(math.Floor(endY/gridSize))
	if sX == eX && sY == eY { return nil }

	openSet := &PriorityQueue{}
	heap.Init(openSet)
	closedSet := make(map[string]bool)
	nodes := make(map[string]*PathNode)

	startNode := &PathNode{X: sX, Y: sY, G: 0, H: heuristic(sX, sY, eX, eY)}
	startNode.F = startNode.G + startNode.H
	heap.Push(openSet, startNode)
	nodes[fmt.Sprintf("%d,%d", sX, sY)] = startNode

	limit := 500 
	count := 0
	
	for openSet.Len() > 0 && count < limit {
		count++
		current := heap.Pop(openSet).(*PathNode)
		if current.X == eX && current.Y == eY { return reconstructPath(current, gridSize) }
		closedSet[fmt.Sprintf("%d,%d", current.X, current.Y)] = true
		
		for _, neighbor := range g.getNeighbors(current, eX, eY, gridSize) {
			key := fmt.Sprintf("%d,%d", neighbor.X, neighbor.Y)
			if closedSet[key] { continue }
			
			tentativeG := current.G + 1.0
			if existing, ok := nodes[key]; !ok || tentativeG < existing.G {
				neighbor.G, neighbor.H = tentativeG, heuristic(neighbor.X, neighbor.Y, eX, eY)
				neighbor.F, neighbor.Parent = neighbor.G + neighbor.H, current
				if !ok {
					heap.Push(openSet, neighbor)
					nodes[key] = neighbor
				} else {
					heap.Fix(openSet, existing.index)
				}
			}
		}
	}
	return nil
}

func heuristic(x1, y1, x2, y2 int) float64 {
	return math.Sqrt(float64((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)))
}

func (g *Game) getNeighbors(n *PathNode, goalX, goalY int, gridSize float64) []*PathNode {
	neighbors := []*PathNode{}
	dirs := [][]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for _, d := range dirs {
		nX, nY := n.X+d[0], n.Y+d[1]
		posX, posY := float64(nX)*gridSize, float64(nY)*gridSize
		
		isBlocked := false
		// 1. Check Obstacles
		for _, o := range g.obstacles {
			if o.Alive && o.IsColliding(posX, posY, 0.5) { isBlocked = true; break }
		}
		if isBlocked { continue }
		
		// 2. Check Miasma Avoidance (Rotten corpses)
		for _, char := range g.characters {
			if char.ActionState == ActorDead && char.Actor.RotTicks > TicksPerDay {
				dist := math.Sqrt(math.Pow(posX-char.X, 2) + math.Pow(posY-char.Y, 2))
				if dist < 4.0 { isBlocked = true; break }
			}
		}
		
		if !isBlocked || (nX == goalX && nY == goalY) {
			neighbors = append(neighbors, &PathNode{X: nX, Y: nY})
		}
	}
	return neighbors
}

func reconstructPath(node *PathNode, gridSize float64) []engine.Point {
	path := []engine.Point{}
	for node != nil {
		path = append([]engine.Point{{X: float64(node.X)*gridSize, Y: float64(node.Y)*gridSize}}, path...)
		node = node.Parent
	}
	return path
}
