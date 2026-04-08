package game

import (
	"oinakos/internal/engine"
	"testing"
)

func TestRestingImagePriority(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter

	// 1. Setup mock images
	staticImg := &engine.MockImage{W: 100, H: 100}
	restingImg := &engine.MockImage{W: 100, H: 100}
	
	pc.Config.StaticImage = staticImg
	pc.Config.RestingImage = restingImg
	
	// 2. Set to Resting state
	pc.ActionState = ActorResting
	
	// 3. Verify it picks RestingImage
	sprite := DrawActorGetSprite(&pc.Actor)
	if sprite != restingImg {
		t.Errorf("Expected resting image when resting, got something else")
	}
	
	// 4. Remove RestingImage, should fallback to StaticImage
	pc.Config.RestingImage = nil
	sprite = DrawActorGetSprite(&pc.Actor)
	if sprite != staticImg {
		t.Errorf("Expected static image when resting if resting image is missing, got something else")
	}
}

func TestCrouchingImagePriority(t *testing.T) {
	g := setupTestGame()
	pc := g.playableCharacter

	// 1. Setup mock images
	staticImg := &engine.MockImage{W: 100, H: 100}
	crouchImg := &engine.MockImage{W: 100, H: 100}
	
	pc.Config.StaticImage = staticImg
	pc.Config.CrouchImage = crouchImg
	
	// 2. Set to Crouching state
	pc.ActionState = ActorCrouching
	
	// 3. Verify it picks CrouchImage
	sprite := DrawActorGetSprite(&pc.Actor)
	if sprite != crouchImg {
		t.Errorf("Expected crouching image when crouching, got something else")
	}
}
