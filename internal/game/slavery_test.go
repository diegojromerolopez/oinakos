package game

import (
	"testing"
)

func TestSlaveryMechanics(t *testing.T) {
	ctx := NewTestContext()
	ctx.Settings.AdultMode = true

	master := NewCharacter(10, 10, &EntityConfig{ID: "master", Name: "Player"}, 25, true, nil)
	master.UID = "player_uid"
	slave := NewCharacter(11, 11, &EntityConfig{ID: "slave", Behavior: "slave"}, 25, false, nil)
	slave.UID = "slave_uid"
	slave.Behavior = BehaviorSlave
	slave.MasterID = "player_uid"
	slave.Alignment = AlignmentAlly
	ctx.World.PlayableCharacter = master
	ctx.World.Characters = []*Character{master, slave}

	t.Run("Slave defends Master", func(t *testing.T) {
		enemy := NewCharacter(12, 12, &EntityConfig{ID: "enemy"}, 25, false, nil)
		enemy.Alignment = AlignmentEnemy
		enemy.TargetActor = &master.Actor
		ctx.World.Characters = append(ctx.World.Characters, enemy)

		slave.updateAI(ctx)
		if slave.TargetActor != &enemy.Actor {
			t.Errorf("Slave did not target enemy attacking their master")
		}
	})

	t.Run("Slave resignation when attacked by Master", func(t *testing.T) {
		slave.TargetActor = nil
		slave.ActionState = ActorIdle
		slave.TakeDamage(1, master, ctx)
		
		if slave.TargetActor != nil {
			t.Errorf("Slave targeted master after being attacked")
		}
		if slave.ActionState != ActorCrouching {
			t.Errorf("Slave did not adopt submissive pose")
		}
	})

	t.Run("Slave Breeding/Inheritance", func(t *testing.T) {
		mother := NewCharacter(10, 10, &EntityConfig{ID: "peasant_female", Gender: "female"}, 20, false, nil)
		mother.Behavior = BehaviorSlave
		mother.MasterID = "player_uid"
		mother.IsPregnant = true
		mother.GestationTicks = 0
		mother.FatherID = "someone"
		
		mother.giveBirth(ctx)
		
		foundBaby := false
		for _, c := range ctx.World.Characters {
			if c.ParentID == mother.Name {
				foundBaby = true
				if c.Behavior != BehaviorSlave || c.MasterID != "player_uid" {
					t.Errorf("Baby did not inherit slave status or master")
				}
			}
		}
		if !foundBaby { t.Errorf("No baby was born") }
	})

	t.Run("Slave Income Redirection", func(t *testing.T) {
		startMasterDenarii := master.Denarii
		slave.AddDenarii(10, ctx.World)
		if master.Denarii != startMasterDenarii+10 {
			t.Errorf("Income was not redirected to master")
		}
	})

	t.Run("Adult Mode Gating in Maps", func(t *testing.T) {
		wm := &WorldManager{game: &Game{settings: &Settings{AdultMode: false}, characterRegistry: ctx.Registries.Characters, archetypeRegistry: ctx.Registries.Archetypes, obstacleRegistry: ctx.Registries.Obstacles, Registries: ctx.Registries}}
		
		mt := &MapType{
			ID: "test_map",
			Inhabitants: []Inhabitant{
				{Archetype: "slaver", AdultMode: true},
				{Archetype: "peasant_male", AdultMode: false},
			},
		}
		
		game := wm.game
		game.currentMapType = *mt
		game.playableCharacter = master
		
		wm.LoadMapLevel() // Indirectly tests filtering
		
		slaverCount := 0
		for _, c := range game.characters {
			if c.Config != nil && c.Config.ID == "slaver" { slaverCount++ }
		}
		if slaverCount > 0 {
			t.Errorf("Slaver spawned even though AdultMode is OFF")
		}
	})

	t.Run("Selling Slave via Dialogue", func(t *testing.T) {
		g := &Game{playableCharacter: master, characters: []*Character{master, slave}, settings: &Settings{AdultMode: true}}
		slaver := NewCharacter(10, 11, &EntityConfig{ID: "slaver"}, 25, false, nil)
		
		master.Denarii = 0
		g.ApplyDialogueEffect(slaver, DialogueEffect{Type: "sell_slave"})
		
		if master.Denarii != 50 {
			t.Errorf("Selling slave did not award Denarii")
		}
		if slave.MasterID != "" {
			t.Errorf("Sold slave still has a MasterID")
		}
	})

	t.Run("NPC Autonomous Slave Buying", func(t *testing.T) {
		ctx.Settings.AdultMode = true
		ctx.World.PlayableCharacter = nil // Clear player to avoid targeting interference
		
		wealthyNPC := NewCharacter(10, 10, &EntityConfig{ID: "rich_noble", Name: "WealthyNPC"}, 25, false, nil)
		wealthyNPC.State.HealthPoints = 100
		wealthyNPC.UID = "rich_uid"
		wealthyNPC.Denarii = 500
		wealthyNPC.Behavior = BehaviorWander
		wealthyNPC.Alignment = AlignmentNeutral
		
		slaver := NewCharacter(10, 10, &EntityConfig{ID: "slaver", Name: "SlaverNPC"}, 25, false, nil)
		slaver.State.HealthPoints = 100
		slaver.Behavior = BehaviorSlaver
		slaver.UID = "slaver_uid"
		slaver.Alignment = AlignmentNeutral
		
		targetSlave := NewCharacter(10, 10, &EntityConfig{ID: "slave/female", Name: "FemaleSlave", Gender: "female"}, 25, false, nil)
		targetSlave.State.HealthPoints = 100
		targetSlave.Behavior = BehaviorSlave
		targetSlave.MasterID = ""
		targetSlave.Alignment = AlignmentNeutral
		
		ctx.World.Characters = []*Character{wealthyNPC, slaver, targetSlave}
		ctx.World.Game.characters = ctx.World.Characters
		ctx.World.Game.obstacles = ctx.World.Obstacles
		
		// Run AI - should move towards slaver (if far) or buy (if near)
		wealthyNPC.updateAI(ctx)
		
		if wealthyNPC.Denarii != (500 - 150) {
			t.Errorf("Wealthy NPC did not autonomously purchase the slave when close (Denarii=%d)", wealthyNPC.Denarii)
		}
		
		if targetSlave.MasterID != wealthyNPC.UID {
			t.Errorf("Slave MasterID did not update after autonomous purchase")
		}
	})

	t.Run("Manumission on Master Death", func(t *testing.T) {
		master := NewCharacter(0, 0, &EntityConfig{ID: "master"}, 25, false, nil)
		master.UID = "dead_master_uid"
		slave := NewCharacter(1, 1, &EntityConfig{ID: "slave"}, 25, false, nil)
		slave.Behavior = BehaviorSlave
		slave.MasterID = master.UID
		
		ctx.World.Characters = []*Character{master, slave}
		master.die(nil, ctx)
		
		if slave.MasterID != "" || slave.Behavior != BehaviorWander {
			t.Errorf("Slave was not freed upon master's death")
		}
	})

	t.Run("Adult Mode Toggle Mid-Game", func(t *testing.T) {
		slave := NewCharacter(0, 0, &EntityConfig{ID: "slave"}, 25, false, nil)
		slave.Behavior = BehaviorSlave
		slave.MasterID = "some_master"
		
		ctx.Settings.AdultMode = false
		slave.updateAI(ctx)
		
		if slave.Behavior != BehaviorWander {
			t.Errorf("Slave behavior did not revert to Wander when AdultMode was toggled OFF")
		}
	})
}
