package game

import (
	"os"
	"oinakos/internal/engine"
	"sync/atomic"
	"testing"
)

func TestVenburgoDailyUptime(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long venburgo simulation")
	}

	// 1. Initialize real registries from filesystem
	// The test runs in internal/game, so we need to go up to the root
	assets := os.DirFS("../..")
	archetypes := NewArchetypeRegistry()
	characters := NewCharacterRegistry()
	obstacles := NewObstacleRegistry()
	objects := NewObjectRegistry()
	maps := NewMapTypeRegistry()

	if err := archetypes.LoadAll(assets); err != nil { t.Fatalf("failed to load archetypes: %v", err) }
	if err := characters.LoadAll(assets); err != nil { t.Fatalf("failed to load characters: %v", err) }
	characters.ProcessInheritance(archetypes)
	if err := obstacles.LoadAll(assets); err != nil { t.Fatalf("failed to load obstacles: %v", err) }
	if err := objects.LoadAll(assets); err != nil { t.Fatalf("failed to load objects: %v", err) }
	if err := maps.LoadAll(assets); err != nil { t.Fatalf("failed to load maps: %v", err) }

	// 2. Setup Game and Engine mocks
	pConfig, err := LoadPlayableCharacterConfig(assets)
	if err != nil { t.Fatalf("failed to load player config: %v", err) }

	g := &Game{
		width:             1280,
		height:            720,
		mapTypeRegistry:   maps,
		archetypeRegistry: archetypes,
		characterRegistry: characters,
		obstacleRegistry:  obstacles,
		Registries: &RegistryContainer{
			Archetypes: archetypes,
			Characters: characters,
			Obstacles:  obstacles,
			Objects:    objects,
			Maps:       maps,
		},
		settings: DefaultSettings(),
	}
	g.World = NewWorld()
	g.World.Game = g
	g.worldManager = NewWorldManager(g)
	g.mechanicsManager = NewMechanicsManager(g)
	g.input = NewMockInputManager()
	g.audio = NewMockAudioManager()
	g.camera = engine.NewCamera(0, 0)

	g.playableCharacter = NewCharacter(0, 0, pConfig, 1, true, objects)
	g.World.PlayableCharacter = g.playableCharacter
	g.characters = []*Character{g.playableCharacter}
	g.World.Characters = g.characters

	// 3. Load Venburgo Map
	g.initialMapID = "venburgo"
	g.worldManager.LoadMapLevel()

	oinakos := g.playableCharacter
	if oinakos == nil {
		t.Fatalf("Oinakos (playable character) not loaded in Venburgo")
	}

	// Initial State
	t.Logf("Initial State: Name=%s, HP=%d, Sanity=%.1f", oinakos.Name, oinakos.State.HealthPoints, oinakos.State.Sanity)

	// 4. Simulate 1 Day (17,280 ticks)
	g.simulationMode = true
	g.settings.AISimulationMode = true
	atomic.StoreInt32(&g.LoadingProgress, 1000)
	
	for i := 0; i < TicksPerDay; i++ {
		if err := g.Update(); err != nil {
			t.Fatalf("Game update failed at tick %d: %v", i, err)
		}
	}

	// 5. Assertions on Final State
	t.Logf("Final State (Day 1): HP=%d/%.0f, Sanity=%.1f, Fatigue=%.1f, Thirst=%.1f, Hunger=%.1f, HydrationBuffer=%d", 
		oinakos.State.HealthPoints, float64(oinakos.State.MaxHealthPoints), oinakos.State.Sanity, oinakos.State.Fatigue, oinakos.State.Thirst, oinakos.State.Hunger, oinakos.State.HydrationBuffer)

	// Health check: Should not be dying
	if float64(oinakos.State.HealthPoints) < float64(oinakos.State.MaxHealthPoints)*0.9 {
		t.Errorf("Oinakos health is too low: %d/%d", oinakos.State.HealthPoints, oinakos.State.MaxHealthPoints)
	}

	// Sanity check: Should stay above distress levels
	if oinakos.State.Sanity < 60 {
		t.Errorf("Oinakos is distressed: Sanity=%.1f", oinakos.State.Sanity)
	}

	// Fatigue check: With proactive sleeping at 70 and ShiftRest at 16:00, he should be rested
	if oinakos.State.Fatigue >= 50 {
		t.Errorf("Oinakos fatigue management failed, too tired after 1 day: %.1f", oinakos.State.Fatigue)
	}

	// Thirst check: Proactive seeking (30) + HydrationBuffer should keep this low
	if oinakos.State.Thirst > 60 {
		t.Errorf("Oinakos is too thirsty in Venburgo: %.2f (Buffer: %d)", oinakos.State.Thirst, oinakos.State.HydrationBuffer)
	}

	// Hunger check: Proactive seeking (70) should keep this manageable
	if oinakos.State.Hunger > 60 {
		t.Errorf("Oinakos is too hungry in Venburgo: %.2f", oinakos.State.Hunger)
	}
}
