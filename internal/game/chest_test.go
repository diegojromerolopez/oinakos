package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestChestMechanics(t *testing.T) {
	// Setup
	reg := NewObjectRegistry()
	keyConfig := &ObjectConfig{ID: "key", Name: "Key", Type: "key"}
	reg.Objects["key"] = keyConfig
	
	itemConfig := &ObjectConfig{ID: "gold", Name: "Gold Nugget", Weight: 1.0}
	reg.Objects["gold"] = itemConfig

	chestArch := &ObstacleArchetype{
		ID:             "chest",
		Name:           "Chest",
		MaxCapacity:    100.0,
		LockResistance: 50,
	}

	g := &Game{
		Registries: &RegistryContainer{
			Objects: reg,
		},
		World: &World{
			Items:         []*ItemInstance{},
			FloatingTexts: []*FloatingText{},
		},
	}
	
	owner := &Actor{ID: "owner", Name: "Owner", MaxWeight: 100, Config: &EntityConfig{}}
	g.playableCharacter = &Character{}
	g.playableCharacter.Actor.ID = "thief"
	g.playableCharacter.Actor.Name = "Thief"
	g.playableCharacter.Actor.MaxWeight = 100
	thief := &g.playableCharacter.Actor
	
	// Give thief locksmith ability
	thief.Config = &EntityConfig{
		Abilities: map[string]Ability{
			"locksmith": {Yield: "dexterity * 1.0"},
		},
	}
	thief.PrimaryAttributes.Dexterity = 10

	chest := NewObstacle("chest_1", 0, 0, chestArch)
	chest.OwnerID = "owner"
	chest.Locked = true
	chest.LockHealth = 50

	// 1. Unlocking with the particular key
	t.Run("UnlockWithKey", func(t *testing.T) {
		key := NewItemInstance("key_1", keyConfig, 0, 0)
		key.TargetID = "chest_1"
		thief.Inventory = []*ItemInstance{key}
		
		if !g.TryUnlock(thief, chest) {
			t.Error("Should be able to unlock with correct key")
		}
		if chest.Locked {
			t.Error("Chest should be unlocked")
		}
	})

	// 2. Not being able to unlock because the key does not open that chest
	t.Run("UnlockWithWrongKey", func(t *testing.T) {
		chest.Locked = true
		key := NewItemInstance("key_2", keyConfig, 0, 0)
		key.TargetID = "wrong_chest"
		thief.Inventory = []*ItemInstance{key}
		
		// Since thief has locksmith ability (dexterity 10), we reset dexterity to 0 to test just the key failure
		thief.PrimaryAttributes.Dexterity = 0
		if g.TryUnlock(thief, chest) {
			t.Error("Should not be able to unlock with wrong key")
		}
		if !chest.Locked {
			t.Error("Chest should still be locked")
		}
		thief.PrimaryAttributes.Dexterity = 10 // Restore
	})

	// 3. Opening a locked chest (locksmith)
	t.Run("UnlockLocksmith", func(t *testing.T) {
		chest.Locked = true
		chest.LockHealth = 10
		thief.Inventory = []*ItemInstance{}
		
		if !g.TryUnlock(thief, chest) {
			t.Error("Locksmith should be able to pick the lock")
		}
		if chest.Locked {
			t.Error("Chest should be unlocked")
		}
		if chest.LockBroken {
			t.Error("Lock should not be broken by picking")
		}
	})

	// 4. Breaking the lock of a closed chest (brute force)
	t.Run("BreakLock", func(t *testing.T) {
		chest.Locked = true
		chest.LockHealth = 5
		chest.LockBroken = false
		// Actor with no locksmith ability
		brute := &Actor{ID: "brute", Name: "Brute", BaseAttack: 10, Config: &EntityConfig{}}
		
		if !g.TryUnlock(brute, chest) {
			t.Error("Brute should be able to break the lock")
		}
		if chest.Locked {
			t.Error("Chest should be unlocked")
		}
		if !chest.LockBroken {
			t.Error("Lock should be broken")
		}
	})

	// 5. Not being able to open an already open chest
	t.Run("AlreadyOpen", func(t *testing.T) {
		chest.Locked = false
		if !g.TryUnlock(owner, chest) {
			t.Error("TryUnlock on open chest should return true (already open)")
		}
	})

	// 6. Dropping off objects to an open chest
	t.Run("DropToOpenChest", func(t *testing.T) {
		chest.Locked = false
		gold := NewItemInstance("gold_1", itemConfig, 0, 0)
		owner.Inventory = []*ItemInstance{gold}
		
		if !g.TryDropToObstacle(owner, 0, chest) {
			t.Error("Should be able to drop item to open chest")
		}
		if len(chest.Inventory) != 1 {
			t.Errorf("Chest should have 1 item, got %d", len(chest.Inventory))
		}
		if len(owner.Inventory) != 0 {
			t.Error("Owner inventory should be empty")
		}
	})

	// 7. Not being able to drop off objects to a locked chest
	t.Run("DropToLockedChest", func(t *testing.T) {
		chest.Locked = true
		chest.LockHealth = 1000 // Too high to break in one go
		gold := NewItemInstance("gold_2", itemConfig, 0, 0)
		thief.Inventory = []*ItemInstance{gold}
		thief.PrimaryAttributes.Dexterity = 0 // No picking
		
		if g.TryDropToObstacle(thief, 0, chest) {
			t.Error("Should not be able to drop item to locked chest")
		}
		if len(thief.Inventory) != 1 {
			t.Error("Thief should still have the item")
		}
	})

	// 8. Picking up an object from an open chest
	t.Run("PickupFromOpenChest", func(t *testing.T) {
		chest.Locked = false
		// Already has gold_1 from previous test
		if !g.TryPickupFromObstacle(owner, chest, 0) {
			t.Error("Should be able to pickup from open chest")
		}
		if len(chest.Inventory) != 0 {
			t.Error("Chest should be empty")
		}
		if len(owner.Inventory) != 1 {
			t.Error("Owner should have 1 item")
		}
	})

	// 9. Not being able to pick objects from a locked chest
	t.Run("PickupFromLockedChest", func(t *testing.T) {
		gold := NewItemInstance("gold_3", itemConfig, 0, 0)
		chest.Inventory = []*ItemInstance{gold}
		chest.Locked = true
		chest.LockHealth = 1000
		
		if g.TryPickupFromObstacle(thief, chest, 0) {
			t.Error("Should not be able to pickup from locked chest")
		}
		if len(chest.Inventory) != 1 {
			t.Error("Chest should still have the item")
		}
	})

	// 10. Locking an open chest
	t.Run("LockChest", func(t *testing.T) {
		chest.Locked = false
		chest.LockBroken = false
		key := NewItemInstance("key_1", keyConfig, 0, 0)
		key.TargetID = "chest_1"
		owner.Inventory = []*ItemInstance{key}
		
		if !g.TryLock(owner, chest) {
			t.Error("Owner should be able to lock the chest with key")
		}
		if !chest.Locked {
			t.Error("Chest should be locked")
		}
	})

	// 11. Not being able to lock a locked chest
	t.Run("LockAlreadyLocked", func(t *testing.T) {
		// chest is already locked from previous test
		if g.TryLock(owner, chest) {
			t.Error("Should return false when locking an already locked chest")
		}
	})
	
	// 12. Repairing a broken lock
	t.Run("RepairLock", func(t *testing.T) {
		chest.LockBroken = true
		chest.Locked = false
		
		thief.PrimaryAttributes.Dexterity = 10
		if !g.TryRepairObstacleLock(thief, chest) {
			t.Error("Locksmith should be able to repair the lock")
		}
		if chest.LockBroken {
			t.Error("Lock should be repaired (not broken)")
		}
	})

	// 13. Trying to open a locked chest without key and fails (insufficient skill/strength)
	t.Run("UnlockFailsWithoutKey", func(t *testing.T) {
		chest.Locked = true
		chest.LockHealth = 1000
		chest.LockBroken = false
		thief.Inventory = []*ItemInstance{}
		thief.PrimaryAttributes.Dexterity = 0 // Remove locksmith skill
		
		if g.TryUnlock(thief, chest) {
			t.Error("Should fail to unlock without key and with high resistance")
		}
		if !chest.Locked {
			t.Error("Chest should still be locked")
		}
	})

	// 14. Dropping an object that exceeds weight capacity
	t.Run("DropExceedsCapacity", func(t *testing.T) {
		chest.Locked = false
		chest.Inventory = []*ItemInstance{}
		chest.TotalWeight = 0
		
		heavyItemConfig := &ObjectConfig{ID: "heavy_rock", Name: "Heavy Rock", Weight: 200.0}
		heavyItem := NewItemInstance("rock_1", heavyItemConfig, 0, 0)
		owner.Inventory = []*ItemInstance{heavyItem}
		
		if g.TryDropToObstacle(owner, 0, chest) {
			t.Error("Should not be able to stash item that exceeds capacity")
		}
		if len(chest.Inventory) != 0 {
			t.Error("Chest should be empty")
		}
	})

	// 15. Picking up an object that exceeds character's weight capacity
	t.Run("PickupExceedsOwnerWeight", func(t *testing.T) {
		chest.Locked = false
		heavyRock := NewItemInstance("heavy_rock_2", &ObjectConfig{ID: "heavy_rock", Weight: 150.0}, 0, 0)
		chest.Inventory = []*ItemInstance{heavyRock}
		owner.MaxWeight = 100 // Character is weaker than the rock
		
		if g.TryPickupFromObstacle(owner, chest, 0) {
			t.Error("Should not be able to pickup item that exceeds personal capacity")
		}
		if len(chest.Inventory) != 1 {
			t.Error("Item should still be in the chest")
		}
	})

	// 16. Owner can access their own locked chest without a key
	t.Run("OwnerAccessWithoutKey", func(t *testing.T) {
		chest.Locked = true
		owner.Inventory = []*ItemInstance{} // No key
		
		if !g.TryUnlock(owner, chest) {
			t.Error("Owner should always be able to access their own chest")
		}
	})

	// 17. Cannot lock a chest with a broken lock
	t.Run("LockBrokenChestFailure", func(t *testing.T) {
		chest.Locked = false
		chest.LockBroken = true
		key := NewItemInstance("key_1", &ObjectConfig{ID: "key", Type: "key"}, 0, 0)
		key.TargetID = "chest_1"
		owner.Inventory = []*ItemInstance{key}
		
		if g.TryLock(owner, chest) {
			t.Error("Should not be able to lock a chest with a broken lock mechanism")
		}
	})

	// 18. Dropping an object when the chest already has items and the new one doesn't fit
	t.Run("PartialCapacityFailure", func(t *testing.T) {
		chest.Locked = false
		chest.Inventory = []*ItemInstance{
			NewItemInstance("gold_1", &ObjectConfig{ID: "gold", Weight: 60.0}, 0, 0),
		}
		chest.TotalWeight = 60.0
		chest.Archetype.MaxCapacity = 100.0 // Remaining space: 40.0
		
		heavyItem := NewItemInstance("gold_2", &ObjectConfig{ID: "gold", Weight: 50.0}, 0, 0)
		owner.Inventory = []*ItemInstance{heavyItem}
		
		if g.TryDropToObstacle(owner, 0, chest) {
			t.Error("Should not be able to drop item that exceeds remaining capacity (60 + 50 > 100)")
		}
		if len(chest.Inventory) != 1 {
			t.Errorf("Chest should still have only 1 item, got %d", len(chest.Inventory))
		}
	})

	// 19. Cumulative progress: a strong lock takes multiple attempts
	t.Run("CumulativeUnlockProgress", func(t *testing.T) {
		chest.Locked = true
		chest.LockHealth = 25
		brute := &Actor{ID: "brute", Name: "Brute", BaseAttack: 10, Config: &EntityConfig{}}
		
		// Attempt 1: 25 - 10 = 15 (Still locked)
		if g.TryUnlock(brute, chest) {
			t.Error("Should not unlock in one hit")
		}
		if !chest.Locked || chest.LockHealth != 15 {
			t.Errorf("Should be locked with 15 health, got %d", chest.LockHealth)
		}
		
		// Attempt 2: 15 - 10 = 5 (Still locked)
		g.TryUnlock(brute, chest)
		if !chest.Locked || chest.LockHealth != 5 {
			t.Error("Should still be locked after second hit")
		}
		
		// Attempt 3: 5 - 10 = -5 (Unlocked!)
		if !g.TryUnlock(brute, chest) {
			t.Error("Should unlock on final hit")
		}
		if chest.Locked {
			t.Error("Chest should be open")
		}
	})

	// 20. Crouching animation state trigger
	t.Run("CrouchAnimationState", func(t *testing.T) {
		chest.Locked = false
		chest.Inventory = []*ItemInstance{}
		gold := NewItemInstance("gold_1", &ObjectConfig{ID: "gold"}, 0, 0)
		thief.Inventory = []*ItemInstance{gold}
		
		// Setup crouch image in config so it triggers
		thief.Config.CrouchImage = &engine.MockImage{} 
		thief.ActionState = ActorIdle
		
		g.TryDropToObstacle(thief, 0, chest)
		
		if thief.ActionState != ActorCrouching {
			t.Error("Actor should be in crouching state after stashing item")
		}
		if thief.CrouchTimer <= 0 {
			t.Error("Crouch timer should be set")
		}
	})
}
