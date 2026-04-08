package game

import (
	"fmt"
	"image/color"
	"math"
	"oinakos/internal/engine"
)

func (g *Game) handleInteractionInput() {
	if g.ActiveDialogue != nil { return }
	g.handleDialogueInput(); g.handleDialogueProximity(); g.handleLogScrolling()
	mx, my := g.input.MousePosition(); offsetX, offsetY := g.camera.GetOffsets(g.width, g.height)
	if g.pinnedCharacter != nil {
		boxW, boxH := 320, 480
		if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			if mx >= g.pinnedUIX && mx <= g.pinnedUIX+boxW && my >= g.pinnedUIY && my <= g.pinnedUIY+40 {
				g.isDraggingPinnedUI, g.dragPinnedOffsetX, g.dragPinnedOffsetY = true, mx-g.pinnedUIX, my-g.pinnedUIY
			}
		}
		if g.isDraggingPinnedUI {
			if g.input.IsMouseButtonPressed(engine.MouseButtonLeft) {
				g.pinnedUIX, g.pinnedUIY = mx-g.dragPinnedOffsetX, my-g.dragPinnedOffsetY
				if g.pinnedUIX < 0 { g.pinnedUIX = 0 }; if g.pinnedUIY < 0 { g.pinnedUIY = 0 }
				if g.pinnedUIX+boxW > g.width { g.pinnedUIX = g.width - boxW }; if g.pinnedUIY+boxH > g.height { g.pinnedUIY = g.height - boxH }
			} else { g.isDraggingPinnedUI = false }
		}
	}
	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		if g.pinnedCharacter != nil && !g.isDraggingPinnedUI {
			bx, by, boxW, boxH := g.pinnedUIX, g.pinnedUIY, 320, 480
			yCmds := by + boxH - 120 + 20
			commands := []string{"TALK", "ATTACK", "TRADE", "SEDUCE", "INTIMIDATE", "STEAL", "RESTRAIN", "HEAL", "GIVE ITEM", "SEX", "TORTURE"}
			for i, cmd := range commands {
				cx, cy := bx+10+(i%3)*100, yCmds+(i/3)*25
				if mx >= cx && mx <= cx+85 && my >= cy-15 && my <= cy+10 { g.handlePinnedCommand(cmd, g.pinnedCharacter); return }
			}
			if mx >= bx && mx <= bx+boxW && my >= by && my <= by+boxH { return }
		}
		var found *Character
		for _, n := range g.characters {
			if !n.IsAlive() { continue }
			isoX, isoY := engine.CartesianToIso(n.X, n.Y); scrX, scrY := isoX+offsetX, isoY+offsetY
			if math.Sqrt(math.Pow(float64(mx)-scrX, 2)+math.Pow(float64(my)-(scrY-40), 2)) < 40 { found = n; break }
		}
		if found != nil {
			g.pinnedCharacter = found
			if g.pinnedUIX == 0 && g.pinnedUIY == 0 {
				isoX, isoY := engine.CartesianToIso(found.X, found.Y)
				g.pinnedUIX, g.pinnedUIY = int(isoX+offsetX)+20, int(isoY+offsetY)+20
				if g.pinnedUIX+320 > g.width { g.pinnedUIX = g.width - 340 }; if g.pinnedUIY+480 > g.height { g.pinnedUIY = g.height - 500 }
				if g.pinnedUIX < 0 { g.pinnedUIX = 10 }; if g.pinnedUIY < 0 { g.pinnedUIY = 10 }
			}
		} else if ! (my < 50 || (my > g.height-180 && mx < g.width-20) || (mx < 360 && my < 300)) { g.pinnedCharacter = nil }
	}
	if g.input.IsKeyJustPressed(engine.KeyEscape) { if g.pinnedCharacter != nil { g.pinnedCharacter = nil } else { g.isMenuOpen = true } }
}

func (g *Game) handlePinnedCommand(cmd string, n *Character) {
	pc, ctx := g.playableCharacter, g.GetContext()
	dist := math.Sqrt(math.Pow(pc.X-n.X, 2) + math.Pow(pc.Y-n.Y, 2))
	if dist > 3.0 && cmd != "ATTACK" { g.LogEvent(fmt.Sprintf("%s is too far away!", n.Name), LogInfo); return }
	switch cmd {
	case "TALK": g.InitiateDialogue(n)
	case "ATTACK": pc.TargetActor = &n.Actor; g.LogEvent("Attacking!", LogPlayer); n.Say("Wait! What are you doing?!", ctx); n.AddMemory(g.Tick, "attacked", pc.Name, -5.0)
	case "TRADE": g.ActiveTrader, g.isTradeOpen = n, true; n.LastReaction = "Let's see what you have."; n.AddMemory(g.Tick, "trade", pc.Name, 1.0)
	case "SEDUCE":
		if n.ActionState == ActorIncapacitated || pc.CheckAbilitySuccess("seduce", 0) {
			n.ModifyRomanticInterest(pc.Name, 10.0); g.AddFloatingText("❤", n.X, n.Y-1, color.RGBA{255, 105, 180, 255}); n.Say("You... you are quite charming.", ctx); n.AddMemory(g.Tick, "seduce", pc.Name, 10.0)
		} else { n.Say("Stop that. It's embarrassing.", ctx); n.AddMemory(g.Tick, "failed_seduction", pc.Name, -5.0) }
	case "INTIMIDATE":
		if n.ActionState == ActorIncapacitated || pc.CheckAbilitySuccess("intimidate", 0) {
			n.ModifySubmission(pc.Name, 15.0); n.Say("P-please... I'll do whatever you want.", ctx); n.AddMemory(g.Tick, "intimidate", pc.Name, -10.0)
		} else { n.Say("You don't scare me.", ctx); n.AddMemory(g.Tick, "failed_intimidation", pc.Name, -10.0) }
	case "STEAL":
		if n.ActionState == ActorIncapacitated || pc.CheckAbilitySuccess("steal", 0) { n.LastReaction = "(They haven't noticed yet...)"; n.AddMemory(g.Tick, "theft", pc.Name, 0)
		} else { n.Alignment, n.LastReaction = AlignmentEnemy, "THIEF! HELP!"; n.Say("THIEF! GUARDS! HELP!", ctx); n.AddMemory(g.Tick, "caught_stealing", pc.Name, -30.0) }
	case "RESTRAIN":
		if pc.CompetitiveContest(&n.Actor, "dexterity", "strength") {
			n.ActionState = ActorIncapacitated; g.AddFloatingText("Bound!", n.X, n.Y-1, color.RGBA{200, 200, 200, 255}); n.Say("Get these off me!", ctx); n.AddMemory(g.Tick, "restrain", pc.Name, -15.0)
		} else { n.Alignment = AlignmentEnemy; n.Say("Nice try!", ctx); n.AddMemory(g.Tick, "failed_restrain", pc.Name, -10.0) }
	case "HEAL": if pc.CheckAbilitySuccess("heal", 0) { n.Heal(20); n.Say("Thank you...", ctx); n.AddMemory(g.Tick, "heal", pc.Name, 20.0) }
	case "GIVE ITEM": g.LogEvent("Use Trade screen.", LogInfo)
	case "SEX": if dist < 2.0 { pc.Actor.haveSex(ctx, &n.Actor, "vaginal") }
	case "TORTURE": if dist < 2.0 { pc.Actor.Torture(&n.Actor, ctx) }
	}
}
