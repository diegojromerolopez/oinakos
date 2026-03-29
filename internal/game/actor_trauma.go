package game

import (
	"fmt"
	"math/rand"
	"strings"
)

func (a *Actor) acquireRandomTrauma(attacker ActorInterface) {
	r := rand.Intn(7)
	switch r {
	case 0: if !a.Trauma.LeftArmLost { a.Trauma.LeftArmLost = true; DebugLog("Actor [%s] %s lost LEFT ARM!", a.Alignment, a.Name) }
	case 1: if !a.Trauma.RightArmLost { a.Trauma.RightArmLost = true; DebugLog("Actor [%s] %s lost RIGHT ARM!", a.Alignment, a.Name) }
	case 2: if !a.Trauma.LeftLegLost { a.Trauma.LeftLegLost = true; DebugLog("Actor [%s] %s lost LEFT LEG!", a.Alignment, a.Name) }
	case 3: if !a.Trauma.RightLegLost { a.Trauma.RightLegLost = true; DebugLog("Actor [%s] %s lost RIGHT LEG!", a.Alignment, a.Name) }
	case 4: if a.Trauma.EyesLost < 2 { a.Trauma.EyesLost++; DebugLog("Actor [%s] %s lost an EYE!", a.Alignment, a.Name) }
	case 5: if !a.Trauma.BurnedAlive { a.Trauma.BurnedAlive = true; DebugLog("Actor [%s] %s was BURNED ALIVE!", a.Alignment, a.Name) }
	case 6: if !a.Trauma.SpineBroken { a.Trauma.SpineBroken = true; DebugLog("Actor [%s] %s suffered BROKEN SPINE!", a.Alignment, a.Name) }
	}
}

func (a *Actor) GetTraumaDescription() string {
	res := []string{}
	if a.Trauma.BurnedAlive { res = append(res, "Severely Burned") }
	if a.Trauma.LeftArmLost { res = append(res, "Left Arm Amputated") }
	if a.Trauma.RightArmLost { res = append(res, "Right Arm Amputated") }
	if a.Trauma.LeftLegLost { res = append(res, "Left Leg Amputated") }
	if a.Trauma.RightLegLost { res = append(res, "Right Leg Amputated") }
	if a.Trauma.EyesLost >= 2 { res = append(res, "Permanently Blind") } else if a.Trauma.EyesLost == 1 { res = append(res, "One Eye Lost") }
	if a.Trauma.SpineBroken { res = append(res, "Broken Spine (Paralyzed)") }
	if len(res) == 0 { return "No permanent traumas." }
	return strings.Join(res, ", ")
}

func (a *Actor) GetActiveTraumas() []string {
	var traumas []string
	if a.Trauma.LeftArmLost { traumas = append(traumas, "Left Arm Lost") }
	if a.Trauma.RightArmLost { traumas = append(traumas, "Right Arm Lost") }
	if a.Trauma.LeftLegLost { traumas = append(traumas, "Left Leg Lost") }
	if a.Trauma.RightLegLost { traumas = append(traumas, "Right Leg Lost") }
	if a.Trauma.EyesLost > 0 { traumas = append(traumas, fmt.Sprintf("%d Eyes Lost", a.Trauma.EyesLost)) }
	if a.Trauma.BurnedAlive { traumas = append(traumas, "Burned") }
	if a.Trauma.SpineBroken { traumas = append(traumas, "Spine Broken") }
	return traumas
}

func (a *Actor) GetDeathReason() string {
	if a.State.Age.Max > 0 && a.State.Age.Current >= a.State.Age.Max { return "died of extreme old age" }
	if a.State.Hunger >= 100 { return "starved to death" }
	if a.State.Thirst >= 100 { return "died of dehydration" }
	if a.State.Fatigue >= 100 { return "died of extreme exhaustion" }
	if a.BodyTemperature > 41.5 { return "died of heatstroke" }
	if a.BodyTemperature < 29.5 { return "died of severe hypothermia" }
	if a.State.IsSeptic { return "succumbed to sepsis" }
	if a.Trauma.BurnedAlive { return "succumbed to severe burns" }
	return "was killed in combat"
}
