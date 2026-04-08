package game

import (
	"testing"
)

func TestActor_SocialSentiment(t *testing.T) {
	a := NewCharacter(0, 0, nil, 1, true, nil)
	o := NewCharacter(1, 1, nil, 1, false, nil)
	o.Name = "Neighbor"

	t.Run("basic sentiment", func(t *testing.T) {
		a.ModifySentiment(o.Name, 50)
		s := a.GetRelationshipTier(o.Name)
		if s != "Acquaintance" {
			t.Errorf("expected Acquaintance for sentiment 50, got %s (raw=%v)", s, a.Relationships[o.Name])
		}
		
		a.ModifySentiment(o.Name, 30) // Result 80
		s = a.GetRelationshipTier(o.Name)
		if s != "Friendly" {
			t.Errorf("expected Friendly for sentiment 80, got %s (raw=%v)", s, a.Relationships[o.Name])
		}
		
		a.ModifySentiment(o.Name, -180) // Result -100 (clamped)
		s = a.GetRelationshipTier(o.Name)
		if s != "Enemy" {
			t.Errorf("expected Enemy for sentiment -100, got %s (raw=%v)", s, a.Relationships[o.Name])
		}
	})
	
	t.Run("romantic passion", func(t *testing.T) {
		// Reset
		a.RomanticInterest = make(map[string]float64)
		a.Relationships = make(map[string]float64)
		
		a.ModifyRomanticInterest(o.Name, 50)
		s := a.GetRelationshipTier(o.Name)
		if s != "Romantic" {
			t.Errorf("expected Romantic tier for passion 50, got %s (raw=%v)", s, a.RomanticInterest[o.Name])
		}
		
		a.ModifySentiment(o.Name, 95)
		a.ModifyRomanticInterest(o.Name, 45) // Total 95
		s = a.GetRelationshipTier(o.Name)
		if s != "Devoted" {
			t.Errorf("expected Devoted for sentiment 95/passion 95, got %s (raw s=%v, p=%v)", s, a.Relationships[o.Name], a.RomanticInterest[o.Name])
		}
	})
}
