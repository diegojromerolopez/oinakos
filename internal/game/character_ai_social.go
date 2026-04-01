package game

import (
	"fmt"
	"math"
	"math/rand"
)

// HandleSocial manages non-combat interactions with nearby NPCs.
func (c *Character) HandleSocial(ctx *SystemContext) {
	// Don't interact every tick
	if c.Tick%180 != 0 { return }

	for _, other := range ctx.World.Characters {
		if other == c || !other.IsAlive() { continue }
		
		dist := math.Sqrt(math.Pow(c.X-other.X, 2) + math.Pow(c.Y-other.Y, 2))
		if dist > 3.0 { continue }

		// 1. Check existing relationship (adjusted for hygiene)
		sentiment := c.GetEffectiveSentiment(&other.Actor)
		
		// 2. High Intellect NPCs are more likely to start interactions
		socialPromptChance := 0.05 + (float64(c.PrimaryAttributes.Intellect) * 0.001)
		
		if rand.Float64() < socialPromptChance {
			// Choose a social action
			action := "talk"
			
			// Social Hierarchy
			if (c.Behavior == BehaviorArtisan || c.GetAbilityYield("herbalism") > 60) && other.State.IsSeptic {
				action = "treat_infection"
			} else if c.State.Hunger > 75 || c.State.Thirst > 75 {
				action = "request_food"
			} else if sentiment < -30 && c.PrimaryAttributes.Strength > other.PrimaryAttributes.Strength {
				action = "intimidate"
			} else if (c.Behavior == BehaviorTrader || other.Behavior == BehaviorTrader || rand.Float64() < 0.1) && (c.Denarii > 0 || other.Denarii > 0 || len(c.Inventory) > 0 || len(other.Inventory) > 0) {
				action = "trade"
			} else if sentiment > 30 && c.Config.Gender != other.Config.Gender {
				action = "seduce"
			}

			switch action {
			case "treat_infection":
				other.State.IsSeptic = false
				other.ModifySentiment(c.Name, 15.0)
				if ctx.Log != nil && playerNear(c, ctx) {
					ctx.Log(fmt.Sprintf("%s treated %s's severe infection.", c.Name, other.Name), LogNPC)
				}
				ctx.World.FloatingTexts = append(ctx.World.FloatingTexts, &FloatingText{ Text: "Infection Cured!", X: other.X, Y: other.Y - 1.0, Life: 60, Color: ColorHeal })

			case "request_food":
				foundIdx := -1
				for i, it := range other.Inventory {
					if it != nil && it.Config != nil && (it.Config.Hunger > 0 || it.Config.Thirst > 0) {
						foundIdx = i; break
					}
				}
				if foundIdx >= 0 {
					item := other.Inventory[foundIdx]
					willGive := other.Relationships[c.ID] > 20 || other.Submission[c.ID] > 40
					if willGive {
						other.Inventory = append(other.Inventory[:foundIdx], other.Inventory[foundIdx+1:]...)
						c.Inventory = append(c.Inventory, item)
						c.ModifySentiment(other.Name, 10.0)
						other.ModifySentiment(c.Name, 5.0)
						if ctx.Log != nil && playerNear(c, ctx) {
							ctx.Log(fmt.Sprintf("%s gave some food to the hungry %s.", other.Name, c.Name), LogNPC)
						}
					} else {
						c.ModifySentiment(other.Name, -5.0)
					}
				}
			case "trade":
				if c.Denarii > 0 {
					// 1. C buys something from OTHER
					for i, it := range other.Inventory {
						if it == nil || it.Config == nil { continue }
						price := int(float64(it.Config.Value) * (1.2 - (c.GetAbilityYield("trade") * 0.002)))
						if price < 1 { price = 1 }
						if c.Denarii >= price {
							c.Denarii -= price; other.Denarii += price
							other.Inventory = append(other.Inventory[:i], other.Inventory[i+1:]...)
							c.Inventory = append(c.Inventory, it)
							c.ModifySentiment(other.Name, 2.0); other.ModifySentiment(c.Name, 2.0)
							if ctx.Log != nil && playerNear(c, ctx) { ctx.Log(fmt.Sprintf("%s bought %s from %s for %d denarii.", c.Name, it.Config.Name, other.Name, price), LogNPC) }
							break
						}
					}
				}
				if other.Denarii > 0 {
					// 2. OTHER buys something from C
					for i, it := range c.Inventory {
						if it == nil || it.Config == nil { continue }
						price := int(float64(it.Config.Value) * (0.8 + (other.GetAbilityYield("trade") * 0.002)))
						if price < 1 { price = 1 }
						if other.Denarii >= price {
							other.Denarii -= price; c.Denarii += price
							c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
							other.Inventory = append(other.Inventory, it)
							c.ModifySentiment(other.Name, 1.0); other.ModifySentiment(c.Name, 1.0)
							if ctx.Log != nil && playerNear(c, ctx) { ctx.Log(fmt.Sprintf("%s sold %s to %s for %d denarii.", c.Name, it.Config.Name, other.Name, price), LogNPC) }
							break
						}
					}
				}
			case "talk":
				c.ModifySentiment(other.Name, 1.0)
				other.ModifySentiment(c.Name, 1.0)
				c.ModifyGroupSentiment(ctx, other.Group, 0.1)
				if ctx.Log != nil && playerNear(c, ctx) {
					ctx.Log(fmt.Sprintf("%s and %s are chatting.", c.Name, other.Name), LogNPC)
				}
			case "intimidate":
				if c.CompetitiveAttributeRoll(&other.Actor, "culture") {
					other.ModifySentiment(c.Name, -10.0)
					other.ModifySubmission(c.Name, 15.0)
					c.ModifyGroupSentiment(ctx, other.Group, -0.5)
					if ctx.Log != nil && playerNear(c, ctx) {
						ctx.Log(fmt.Sprintf("%s cowed %s into submission.", c.Name, other.Name), LogNPC)
					}
				} else {
					other.ModifySentiment(c.Name, -5.0)
					other.State.IsAngry = true
					c.ModifyGroupSentiment(ctx, other.Group, -1.0)
				}
			case "seduce":
				if c.CompetitiveAttributeRoll(&other.Actor, "art") {
					if c.RomanticInterest == nil { c.RomanticInterest = make(map[string]float64) }
					if other.RomanticInterest == nil { other.RomanticInterest = make(map[string]float64) }
					c.RomanticInterest[other.ID] += 10.0
					other.RomanticInterest[c.ID] += 10.0
					c.ModifyGroupSentiment(ctx, other.Group, 0.5)
					if ctx.Log != nil && playerNear(c, ctx) {
						ctx.Log(fmt.Sprintf("%s shared a romantic moment with %s.", c.Name, other.Name), LogNPC)
					}
				}
			}
		}
	}
}

func (c *Character) ModifyGroupSentiment(ctx *SystemContext, otherGroup string, delta float64) {
	if c.Group == "" || otherGroup == "" || ctx.World.State.GroupSentiment == nil { return }
	if ctx.World.State.GroupSentiment[c.Group] == nil { ctx.World.State.GroupSentiment[c.Group] = make(map[string]float64) }
	ctx.World.State.GroupSentiment[c.Group][otherGroup] += delta
	if ctx.World.State.GroupSentiment[c.Group][otherGroup] > 100 { ctx.World.State.GroupSentiment[c.Group][otherGroup] = 100 }
	if ctx.World.State.GroupSentiment[c.Group][otherGroup] < -100 { ctx.World.State.GroupSentiment[c.Group][otherGroup] = -100 }
}

func playerNear(c *Character, ctx *SystemContext) bool {
	pc := ctx.World.PlayableCharacter
	if pc == nil { return false }
	return math.Sqrt(math.Pow(c.X-pc.X, 2) + math.Pow(c.Y-pc.Y, 2)) < 15.0
}
func (c *Character) PickIdleBark() string {
	barks := []string{"Beautiful day, isn't it?", "I'm quite hungry...", "Need to get back to work soon.", "Have you seen the latest trade prices?", "I hope the weather holds up."}
	if c.State.Hunger > 50 { barks = append(barks, "My stomach is growling.") }
	if c.State.Thirst > 50 { barks = append(barks, "A cup of wine would be lovely right now.") }
	if c.Denarii < 10 { barks = append(barks, "I'm running low on denarii.") }
	return barks[rand.Intn(len(barks))]
}
