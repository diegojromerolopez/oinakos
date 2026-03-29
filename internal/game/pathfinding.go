package game

import (
	"container/heap"
	"math"
	"oinakos/internal/engine"
)

// PathNode represents a grid cell for A* pathfinding.
type PathNode struct {
	X, Y    int
	G, H, F float64
	Parent  *PathNode
	Index   int // Heap index
}

type PriorityQueue []*PathNode

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].F < pq[j].F }
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *PriorityQueue) Push(x any) {
	n := x.(*PathNode)
	n.Index = len(*pq)
	*pq = append(*pq, n)
}
func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// FindPath returns a series of points from start to target avoiding obstacles.
func (a *Actor) FindPath(targetX, targetY float64, ctx *SystemContext) []engine.Point {
	const maxRange = 32 // Only search up to 32m away locally
	
	startX, startY := int(math.Floor(a.X)), int(math.Floor(a.Y))
	endX, endY := int(math.Floor(targetX)), int(math.Floor(targetY))
	
	if math.Abs(float64(startX-endX)) > maxRange || math.Abs(float64(startY-endY)) > maxRange {
		// Too far for local A*
		return nil
	}

	openSet := &PriorityQueue{}
	heap.Init(openSet)
	
	closedSet := make(map[uint64]bool)
	nodes := make(map[uint64]*PathNode)

	startNode := &PathNode{X: startX, Y: startY, G: 0, H: distance(startX, startY, endX, endY)}
	startNode.F = startNode.H
	heap.Push(openSet, startNode)
	nodes[nodeKey(startX, startY)] = startNode

	iterations := 0
	for openSet.Len() > 0 {
		iterations++
		if iterations > 500 { break } // Performance safety valve

		current := heap.Pop(openSet).(*PathNode)
		if current.X == endX && current.Y == endY {
			return reconstructPath(current)
		}
		
		key := nodeKey(current.X, current.Y)
		closedSet[key] = true

		// 8 directions
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 { continue }
				
				nx, ny := current.X+dx, current.Y+dy
				nKey := nodeKey(nx, ny)
				
				if closedSet[nKey] { continue }
				
				// Collision check
				if a.checkCollisionAt(float64(nx)+0.5, float64(ny)+0.5, ctx.World.Obstacles) {
					continue
				}

				moveCost := 1.0
				if dx != 0 && dy != 0 { moveCost = 1.414 } // Diagonal cost approximation
				
				gScore := current.G + moveCost
				neighbor, exists := nodes[nKey]
				
				if !exists || gScore < neighbor.G {
					if !exists {
						neighbor = &PathNode{X: nx, Y: ny}
						nodes[nKey] = neighbor
					}
					neighbor.Parent = current
					neighbor.G = gScore
					neighbor.H = distance(nx, ny, endX, endY)
					neighbor.F = neighbor.G + neighbor.H
					
					if !exists {
						heap.Push(openSet, neighbor)
					} else {
						heap.Fix(openSet, neighbor.Index)
					}
				}
			}
		}
	}
	
	return nil
}

func distance(ax, ay, bx, by int) float64 {
	return math.Sqrt(math.Pow(float64(ax-bx), 2) + math.Pow(float64(ay-by), 2))
}

func nodeKey(x, y int) uint64 {
	return uint64(uint32(x))<<32 | uint64(uint32(y))
}

func reconstructPath(node *PathNode) []engine.Point {
	var path []engine.Point
	for node != nil {
		path = append([]engine.Point{{X: float64(node.X) + 0.5, Y: float64(node.Y) + 0.5}}, path...)
		node = node.Parent
	}
	// Return the path excluding the start node if possible
	if len(path) > 1 {
		return path[1:]
	}
	return path
}
