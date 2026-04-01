package game
import "testing"
import "fmt"
func TestPrint(t *testing.T) {
	arch := &EntityConfig{
		ID: "test_npc",
		Attributes: PrimaryAttributeConfig{ Strength: IntInterval{Min: 5, Max: 5} },
		Stats: EntityStatsConfig{
			Age: AgeConfig{Current: FloatInterval{Mean: 25.0, SD: 0.0, Mode: "normal"}},
		},
	}
	n := NewCharacter(0, 0, arch, 1, false, nil)
	fmt.Printf("DEBUG: n.BaseAttack=%v str=%v ConfigNil=%v Age=%v\n", n.BaseAttack, n.PrimaryAttributes.Strength, n.Config==nil, n.AgeTicks)
}
