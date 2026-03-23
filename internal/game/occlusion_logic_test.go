package game

import (
	"testing"
)

func TestOcclusion_Logic(t *testing.T) {
	mc := &Actor{X: 10, Y: 10, State: ActorIdle}
	sortY := GetActorSortY(mc)
	if sortY != 20 { t.Errorf("Expected sortY 20, got %v", sortY) }
	
	mc.State = ActorDead
	sortY = GetActorSortY(mc)
	if sortY != -80 { t.Errorf("Expected sortY -80, got %v", sortY) }
	
	obs := &Obstacle{X: 5, Y: 5}
	sortY = GetObstacleSortY(obs)
	if sortY != 10 { t.Errorf("Expected sortY 10, got %v", sortY) }
	
	arch := &ObstacleArchetype{Type: "tree", Footprint: []FootprintPoint{{0, 0}, {2, 0}, {2, 2}, {0, 2}}}
	obs.Archetype = arch
	sortY = GetObstacleSortY(obs)
	if sortY != 12 { t.Errorf("Expected sortY 12, got %v", sortY) }
	
	// Coverage only for IsPointCoveredByObstacle
	IsPointCoveredByObstacle(obs, 0, 0)
}
