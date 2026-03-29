package game

import (
	"fmt"
	"image/color"
	"math"

	"oinakos/internal/engine"
)

func (g *Game) LogEvent(text string, category LogCategory) {
	if len(g.EventLog) > 0 && g.EventLog[len(g.EventLog)-1].Text == text {
		return
	}
	entry := LogEntry{Text: text, Category: category}
	if g.playableCharacter != nil { entry.Ticks = g.playableCharacter.Tick }
	g.EventLog = append(g.EventLog, entry)
	g.LogScrollOffset = 0
	if len(g.EventLog) > 100 { g.EventLog = g.EventLog[len(g.EventLog)-100:] }
}

func (g *Game) handleDialogueInput() {
	talkKey := g.settings.GetKey("talk")
	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) || g.input.IsKeyJustPressed(talkKey) {
		mx, my := g.input.MousePosition()
		
		// If clicking on fixed UI elements
		if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
			if mx >= g.width-110 && mx <= g.width-10 && my >= 20 && my <= 50 {
				g.isMenuOpen = true
				return
			}

			isDialogue := g.ActiveDialogue != nil
			boxH := 300
			if isDialogue && g.ActiveDialogue.UIState == DialogueMaximized { boxH = 600 } else if !isDialogue { boxH = 60 }
			
			bx, by := 10, g.height-boxH-10
			boxW := g.width - 20
			if mx >= bx && mx <= bx+boxW-20 && my >= by && my <= by+boxH {
				g.ToggleDialogueSize()
				return
			}
		}

		// Check for NPCs to talk to (Keyboard interaction only now, or via pinned menu)
		if g.input.IsKeyJustPressed(talkKey) {
			// Keyboard interaction (nearest NPC)
			var bestNPC *Character
			bestDist := 2.5 // Max interaction distance
			pc := g.playableCharacter

			for _, n := range g.characters {
				if !n.IsAlive() || n.Alignment == AlignmentEnemy || n.Config == nil || n.Config.Dialogues == nil { continue }
				dist := math.Sqrt(math.Pow(n.X-pc.X, 2) + math.Pow(n.Y-pc.Y, 2))
				if dist < bestDist {
					bestDist = dist
					bestNPC = n
				}
			}

			if bestNPC != nil {
				g.InitiateDialogue(bestNPC)
				return
			}
		}
	}
}

func (g *Game) handleLogScrolling() {
	mx, my := g.input.MousePosition()
	isDialogue := g.ActiveDialogue != nil
	boxH := 300
	maxLogEntries := 4
	if isDialogue {
		if g.ActiveDialogue.UIState == DialogueMaximized {
			boxH = 650
			maxLogEntries = 12
		} else {
			boxH = 350
			maxLogEntries = 5
		}
	} else {
		if g.LogUIState == DialogueMaximized {
			boxH = 300
			maxLogEntries = 15
		} else {
			boxH = 85
			maxLogEntries = 3
		}
	}
	
	bx, by := 10, g.height-boxH-10
	boxW := g.width - 20
	sbX, sbTrackY, sbTrackH := bx+boxW-10, by+5, boxH-10
	if isDialogue { sbTrackY = by + 25; sbTrackH = maxLogEntries * 15 + 10 }

	if g.input.IsMouseButtonJustPressed(engine.MouseButtonLeft) {
		if mx >= sbX-5 && mx <= sbX+10 && my >= int(sbTrackY) && my <= int(sbTrackY+sbTrackH) { g.IsDraggingLog = true }
	}

	if g.IsDraggingLog {
		if !g.input.IsMouseButtonPressed(engine.MouseButtonLeft) {
			g.IsDraggingLog = false
		} else {
			ratio := 1.0 - float32(my-int(sbTrackY))/float32(sbTrackH)
			if ratio < 0 { ratio = 0 } else if ratio > 1 { ratio = 1 }
			
			maxOffset := len(g.EventLog) - maxLogEntries
			if maxOffset < 0 { maxOffset = 0 }
			g.LogScrollOffset = int(float32(maxOffset) * ratio)
		}
	}

	_, wheelY := g.input.Wheel()
	if wheelY != 0 {
		if mx >= bx && mx <= bx+boxW && my >= by && my <= by+boxH {
			g.LogScrollOffset -= int(wheelY)
			if g.LogScrollOffset < 0 { g.LogScrollOffset = 0 }
			maxScroll := len(g.EventLog) - maxLogEntries
			if maxScroll < 0 { maxScroll = 0 }
			if g.LogScrollOffset > maxScroll { g.LogScrollOffset = maxScroll }
		}
	}
}

func (g *Game) ToggleDialogueSize() {
	if g.ActiveDialogue != nil {
		if g.ActiveDialogue.UIState == DialogueMinimized { g.ActiveDialogue.UIState = DialogueMaximized } else { g.ActiveDialogue.UIState = DialogueMinimized }
	} else {
		if g.LogUIState == DialogueMinimized { g.LogUIState = DialogueMaximized } else { g.LogUIState = DialogueMinimized }
	}
}

func (g *Game) handleDialogueProximity() {
	if g.ActiveDialogue != nil { return }
	for _, n := range g.characters {
		if !n.IsAlive() || n.Alignment == AlignmentEnemy || n.HasInitiatedDialogue { continue }
		if n.Config != nil && n.Config.Dialogues != nil {
			for _, s := range n.Config.Dialogues.StartScenarios {
				if s.AutoInitiate {
					dist := math.Sqrt(math.Pow(n.X-g.playableCharacter.X, 2) + math.Pow(n.Y-g.playableCharacter.Y, 2))
					if dist < s.ProximityRange {
						g.InitiateDialogue(n)
						n.HasInitiatedDialogue = true
						break
					}
				}
			}
		}
	}
}

func (g *Game) InitiateDialogue(npc *Character) {
	if npc.Config == nil || npc.Config.Dialogues == nil { return }
	dr := npc.Config.Dialogues
	greeting := dr.PickGreeting()
	g.LogEvent(fmt.Sprintf("%s: %s", g.playableCharacter.Name, greeting), LogPlayer)
	start := dr.PickStart()
	if start == nil { return }
	g.ActiveDialogue = &DialogueState{ SpeakerNPC: npc, CurrentText: start.Text, Choices: start.Choices, IsActive: true, UIState: DialogueMaximized }
	if start.Next != "" {
		if node, ok := dr.Nodes[start.Next]; ok { g.ActiveDialogue.Choices = node.Choices }
	}
	
	if npc.Behavior == BehaviorTrader {
		g.ActiveTrader = npc
		g.isTradeOpen = true
	}

	// 1. Give a sentiment boost for talking!
	// Cooldown: 18000 ticks = 5 minutes in-game
	cooldown := 18000
	if g.Tick - npc.LastTalkTick > cooldown {
		npc.ModifySentiment(g.playableCharacter.Name, 5.0)
		npc.AddMemory(g.Tick, "conversation", g.playableCharacter.Name, 5.0)
		npc.LastTalkTick = g.Tick
		g.AddFloatingText("❤", npc.X, npc.Y+1.0, color.RGBA{255, 100, 100, 255})
	}

	g.LogEvent(fmt.Sprintf("%s: %s", npc.Name, start.Text), LogNPC)
}

func (g *Game) AdvanceDialogue() {
	if g.ActiveDialogue == nil || len(g.ActiveDialogue.Choices) == 0 {
		g.CloseDialogue()
		return
	}
	choice := g.ActiveDialogue.Choices[g.ActiveDialogue.SelectedChoice]
	g.LogEvent(fmt.Sprintf("%s: %s", g.playableCharacter.Name, choice.Text), LogPlayer)
	for _, effect := range choice.Effects { g.ApplyDialogueEffect(g.ActiveDialogue.SpeakerNPC, effect) }
	if choice.Next == "" || choice.Next == "exit" {
		g.CloseDialogue()
		return
	}
	dr := g.ActiveDialogue.SpeakerNPC.Config.Dialogues
	if node, ok := dr.Nodes[choice.Next]; ok {
		g.ActiveDialogue.CurrentText, g.ActiveDialogue.Choices, g.ActiveDialogue.SelectedChoice = node.Text, node.Choices, 0
		g.LogEvent(fmt.Sprintf("%s: %s", g.ActiveDialogue.SpeakerNPC.Name, node.Text), LogNPC)
	} else { g.CloseDialogue() }
}

func (g *Game) CloseDialogue() { g.ActiveDialogue = nil }

func (g *Game) ApplyDialogueEffect(npc *Character, effect DialogueEffect) {
	switch effect.Type {
	case "change_alignment":
		switch effect.Value {
		case "enemy": npc.Alignment, npc.TargetActor, npc.Behavior = AlignmentEnemy, &g.playableCharacter.Actor, BehaviorKnightHunter
		case "neutral": npc.Alignment, npc.TargetActor, npc.Behavior = AlignmentNeutral, nil, BehaviorWander
		case "ally": npc.Alignment, npc.TargetActor, npc.Behavior = AlignmentAlly, nil, BehaviorEscort
		}
	case "change_behavior":
		switch effect.Value {
		case "flee": npc.Behavior = BehaviorFlee
		case "patrol": npc.Behavior = BehaviorPatrol
		case "follow": npc.Behavior = BehaviorEscort
		}
	}
}
