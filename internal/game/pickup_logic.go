package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"oinakos/internal/engine"
)

func (g *Game) UpdatePickups() {
	if g == nil || g.World == nil {
		return
	}
	pc := g.playableCharacter
	if pc == nil || !pc.IsAlive() {
		return
	}

	// Check input: Spacebar or Mouse left click
	wantsToPick := false
	mx, my := 0, 0
	isClick := false
	if g.input != nil {
		wantsToPick = g.input.IsKeyJustPressed(engine.KeySpace)
		mx, my = g.input.MousePosition()
		isClick = g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft)
	}

	// We'll iterate all items. If spacebar is pressed, pick the closest one in range.
	// If mouse clicked, pick if clicked on it AND in range.

	pickupRange := 2.5
	var remainingItems []*ItemInstance

	pickedUp := false

	// Calculate mouse Cartesian coordinates once if needed
	mxCart, myCart := 0.0, 0.0
	if isClick {
		offX, offY := g.camera.GetOffsets(g.width, g.height)
		mxIso := float64(mx) - offX
		myIso := float64(my) - offY
		mxCart, myCart = engine.IsoToCartesian(mxIso, myIso)
	}

	for _, it := range g.World.Items {
		dist := math.Sqrt(math.Pow(pc.X-it.X, 2) + math.Pow(pc.Y-it.Y, 2))
		
		canPickThis := false

		if wantsToPick && dist < pickupRange && !pickedUp {
			canPickThis = true
		} else if isClick && dist < pickupRange && !pickedUp {
			if it.GetFootprint().Contains(mxCart, myCart) {
				canPickThis = true
			}
		}

		if canPickThis && it.Pickable {
			// Try to pick up
			if g.TryPickup(&pc.Actor, it) {
				pickedUp = true
				continue // Item removed
			}
		}
		remainingItems = append(remainingItems, it)
	}
	g.World.Items = remainingItems
}

func (g *Game) TryPickup(a *Actor, it *ItemInstance) bool {
	if it == nil || !it.Pickable {
		return false
	}

	if a.GetTotalWeight()+it.Weight > a.MaxWeight {
		if g.playableCharacter != nil && a == &g.playableCharacter.Actor {
			g.AddFloatingText("Too heavy!", a.X, a.Y, ColorHarm)
		}
		return false
	}

	capacity := a.Config.MaxItems
	if capacity == 0 {
		capacity = 20 // Default
	}
	if len(a.Inventory) >= capacity {
		if g.playableCharacter != nil && a == &g.playableCharacter.Actor {
			g.AddFloatingText("Inventory full!", a.X, a.Y, ColorHarm)
		}
		return false
	}

	// Success
	it.Pickable = false // Mark as unpickable immediately to prevent race conditions
	
	// Gift logic: if NPC picks up something dropped by another
	if (g.playableCharacter == nil || a != &g.playableCharacter.Actor) && it.DroppedBy != "" && it.DroppedBy != a.Name {
		if a.Relationships == nil { a.Relationships = make(map[string]float64) }
		a.Relationships[it.DroppedBy] += 10.0
		a.AddMemory(a.Tick, "gift", it.DroppedBy, 10.0)
		if g.World != nil {
			g.AddFloatingText("❤", a.X, a.Y, color.RGBA{255, 100, 100, 255})
		}
	}

	a.Inventory = append(a.Inventory, it)
	a.UpdateEffects()
	
	// Set animation state
	// CrouchImage is now loaded in registries
	if a.Config.CrouchImage != nil {
		a.ActionState = ActorCrouching
		a.CrouchTimer = 30 // Show crouch for 0.5s
	}
	
	if g.playableCharacter != nil && a == &g.playableCharacter.Actor {
		g.AddFloatingText(fmt.Sprintf("Picked up %s", it.Config.Name), a.X, a.Y, ColorHeal)
		if g.audio != nil {
			g.audio.PlayRandomSound("pickup") // Fallback to generic if not exist
		}
	}
	
	return true
}

// DropAllItems Drops all inventory and equipped items around the actor
func (g *Game) DropAllItems(a *Actor) {
	radius := 1.0

	dropObject := func(it *ItemInstance) {
		if it == nil || it.Config == nil {
			return
		}
		safeX, safeY := findSafePosition(a.X+(rand.Float64()*radius-radius/2.0), a.Y+(rand.Float64()*radius-radius/2.0), engine.Circle{Radius: 0.5}, g.obstacles)
		
		it.X = safeX
		it.Y = safeY
		it.Pickable = true
		
		if g.World != nil {
			g.World.Items = append(g.World.Items, it)
		}
	}

	for slot, it := range a.Slots {
		if it != nil {
			dropObject(it)
			a.Slots[slot] = nil
		}
	}
	a.UpdateEffects()

	for _, it := range a.Inventory {
		if it != nil {
			dropObject(it)
		}
	}
	a.Inventory = nil

	// Drop meat if this is an animal/entity with meat
	if a.Config != nil && a.Config.Meat > 0 {
		if meatConfig, ok := g.Registries.Objects.Objects["raw_meat"]; ok {
			for i := 0; i < a.Config.Meat; i++ {
				// Each piece of meat is around 2-5 units
				w := 2.0 + rand.Float64()*3.0
				dropX := a.X + (rand.Float64()*radius - radius/2.0)
				dropY := a.Y + (rand.Float64()*radius - radius/2.0)
				safeX, safeY := findSafePosition(dropX, dropY, engine.Circle{Radius: 0.5}, g.obstacles)
				
				item := NewItemInstance(fmt.Sprintf("meat_%s_%d", a.Config.ID, rand.Int()), meatConfig, safeX, safeY)
				item.Weight = w
				if g.World != nil {
					g.World.Items = append(g.World.Items, item)
				}
			}
		}
	}
}

func (g *Game) DropEquippedItem(a *Actor, slot string) bool {
	it, ok := a.Slots[slot]
	if !ok || it == nil {
		return false
	}
	
	radius := 1.0
	safeX, safeY := findSafePosition(a.X+(rand.Float64()*radius-radius/2.0), a.Y+(rand.Float64()*radius-radius/2.0), engine.Circle{Radius: 0.5}, g.obstacles)
	
	it.X = safeX
	it.Y = safeY
	it.DroppedBy = a.Name
	it.Pickable = true
	
	if g.World != nil {
		g.World.Items = append(g.World.Items, it)
	}

	delete(a.Slots, slot)
	a.UpdateEffects()

	if a == &g.playableCharacter.Actor {
		g.AddFloatingText(fmt.Sprintf("Dropped %s", it.Config.Name), a.X, a.Y, color.RGBA{150, 150, 150, 255})
	}

	return true
}

func (g *Game) TryDrop(a *Actor, index int) bool {
	if index < 0 || index >= len(a.Inventory) {
		return false
	}
	
	it := a.Inventory[index]
	
	// Check if equipped
	for slot, equipped := range a.Slots {
		if equipped == it {
			delete(a.Slots, slot)
			break
		}
	}
	
	// Remove from inventory
	a.Inventory = append(a.Inventory[:index], a.Inventory[index+1:]...)
	a.UpdateEffects()
	
	// Add to world with a slight offset so it appears "besides" the character
	offsetX := (rand.Float64() - 0.5) * 0.5
	offsetY := (rand.Float64() - 0.5) * 0.5
	safeX, safeY := findSafePosition(a.X+offsetX, a.Y+offsetY, engine.Circle{Radius: 0.5}, g.obstacles)
	
	it.X = safeX
	it.Y = safeY
	it.DroppedBy = a.Name
	it.Pickable = true
	g.World.Items = append(g.World.Items, it)
	
	if a == &g.playableCharacter.Actor {
		g.AddFloatingText(fmt.Sprintf("Dropped %s", it.Config.Name), a.X, a.Y, ColorHarm)
	}
	
	return true
}

func (g *Game) DropSpecificItem(a *Actor, it *ItemInstance) bool {
	index := -1
	for i, item := range a.Inventory {
		if item == it {
			index = i
			break
		}
	}
	if index != -1 {
		return g.TryDrop(a, index)
	}
	// If it's not in the inventory, it might just be in a slot (although we usually add to inventory first).
	// But TryDrop only pops from Inventory right now. So we must put it in inventory first, then drop it.
	a.Inventory = append(a.Inventory, it)
	return g.TryDrop(a, len(a.Inventory)-1)
}


func (g *Game) AddFloatingText(text string, x, y float64, col color.Color) {
	// Utility for picking logic
	g.World.FloatingTexts = append(g.World.FloatingTexts, &FloatingText{
		Text: text, X: x, Y: y, Life: 60, Color: col,
	})
}
