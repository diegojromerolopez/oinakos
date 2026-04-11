package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestDrawActorGetSprite_ActionPriorities(t *testing.T) {
	staticImg := &engine.MockImage{}
	attackImg := &engine.MockImage{}
	hitImg := &engine.MockImage{}
	crouchImg := &engine.MockImage{}
	corpseImg := &engine.MockImage{}
	chopImg := &engine.MockImage{}
	digImg := &engine.MockImage{}

	config := &EntityConfig{
		StaticImage:   staticImg,
		AttackImage:   attackImg,
		HitImage:      hitImg,
		CrouchImage:   crouchImg,
		CorpseImage:   corpseImg,
		ChoppingImage: chopImg,
		DiggingImage:  digImg,
	}

	actor := &Actor{
		Config: config,
	}

	t.Run("Dead priority", func(t *testing.T) {
		actor.ActionState = ActorDead
		if got := DrawActorGetSprite(actor, false); got != corpseImg {
			t.Errorf("expected corpse image, got %v", got)
		}
	})

	t.Run("Hit priority over attacking", func(t *testing.T) {
		actor.ActionState = ActorAttacking
		actor.HitTimer = 10
		if got := DrawActorGetSprite(actor, false); got != hitImg {
			t.Errorf("expected hit image to override attacking state, got %v", got)
		}
	})

	t.Run("Attacking sprite", func(t *testing.T) {
		actor.ActionState = ActorAttacking
		actor.HitTimer = 0
		if got := DrawActorGetSprite(actor, false); got != attackImg {
			t.Errorf("expected attack image, got %v", got)
		}
	})

	t.Run("Crouching/Resting sprite", func(t *testing.T) {
		actor.ActionState = ActorCrouching
		if got := DrawActorGetSprite(actor, false); got != crouchImg {
			t.Errorf("expected crouch image for crouching, got %v", got)
		}
		actor.ActionState = ActorResting
		if got := DrawActorGetSprite(actor, false); got != crouchImg {
			t.Errorf("expected crouch image for resting, got %v", got)
		}
	})

	t.Run("Chopping specific sprite", func(t *testing.T) {
		actor.ActionState = ActorChopping
		if got := DrawActorGetSprite(actor, false); got != chopImg {
			t.Errorf("expected chopping image, got %v", got)
		}
	})

	t.Run("Digging specific sprite", func(t *testing.T) {
		actor.ActionState = ActorDigging
		if got := DrawActorGetSprite(actor, false); got != digImg {
			t.Errorf("expected digging image, got %v", got)
		}
	})
}

func TestDrawActorGetSprite_Fallbacks(t *testing.T) {
	staticImg := &engine.MockImage{}
	
	t.Run("Fallback to static when action image is missing", func(t *testing.T) {
		config := &EntityConfig{
			StaticImage: staticImg,
			// Other images are nil
		}
		actor := &Actor{
			Config: config,
		}

		states := []ActorState{ActorAttacking, ActorChopping, ActorDigging, ActorCrouching, ActorResting, ActorCooking}
		for _, state := range states {
			actor.ActionState = state
			if got := DrawActorGetSprite(actor, false); got != staticImg {
				t.Errorf("state %v: expected fallback to static image, got %v", state, got)
			}
		}

		actor.ActionState = ActorIdle
		actor.HitTimer = 10
		if got := DrawActorGetSprite(actor, false); got != staticImg {
			t.Errorf("hit state: expected fallback to static image, got %v", got)
		}
	})

	t.Run("Chopping/Digging fallback to AttackImage", func(t *testing.T) {
		attackImg := &engine.MockImage{}
		config := &EntityConfig{
			StaticImage: staticImg,
			AttackImage: attackImg,
			// Chopping/Digging are nil
		}
		actor := &Actor{Config: config}

		actor.ActionState = ActorChopping
		if got := DrawActorGetSprite(actor, false); got != attackImg {
			t.Errorf("chopping: expected fallback to attack image, got %v", got)
		}

		actor.ActionState = ActorDigging
		if got := DrawActorGetSprite(actor, false); got != attackImg {
			t.Errorf("digging: expected fallback to attack image, got %v", got)
		}
	})
}
