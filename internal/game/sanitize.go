package game

import (
	"log"
	"strconv"
)

// HexToRGBA converts a hex string like "#FF00FF" to [4]float32 RGBA components (0.0 to 1.0).
func HexToRGBA(hex string) [4]float32 {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return [4]float32{1, 1, 1, 1} // Fallback to white
	}

	r, errR := strconv.ParseUint(hex[0:2], 16, 8)
	g, errG := strconv.ParseUint(hex[2:4], 16, 8)
	b, errB := strconv.ParseUint(hex[4:6], 16, 8)

	if errR != nil || errG != nil || errB != nil {
		return [4]float32{1, 1, 1, 1}
	}

	return [4]float32{
		float32(r) / 255.0,
		float32(g) / 255.0,
		float32(b) / 255.0,
		1.0,
	}
}

// sanitizeEntityConfig validates and clamps all fields loaded from an archetype YAML.
// Any invalid value is fixed and a warning is logged.
func sanitizeEntityConfig(c *EntityConfig, source string) {
	changed := false

	if c.ID == "" {
		c.ID = "unknown"
		changed = true
	}
	if c.Name == "" {
		log.Printf("Warning [%s]: archetype %q has empty name, using id", source, c.ID)
		c.Name = c.ID
		changed = true
	}
	if c.Stats.HealthPoints.Min < 1 {
		c.Stats.HealthPoints.Min = 1
	}
	if c.Stats.HealthPoints.Max < c.Stats.HealthPoints.Min {
		c.Stats.HealthPoints.Max = c.Stats.HealthPoints.Min
	}
	if c.Stats.Speed.Min <= 0 {
		c.Stats.Speed = FloatInterval{Min: 0.01, Max: 0.01}
		changed = true
	}
	if c.Stats.Speed.Max > 2.0 {
		c.Stats.Speed.Max = 0.5
		if c.Stats.Speed.Min > 0.5 { c.Stats.Speed.Min = 0.5 }
		changed = true
	}
	if c.Stats.BaseAttack.Min < 0 {
		c.Stats.BaseAttack.Min = 0
		if c.Stats.BaseAttack.Max < 0 { c.Stats.BaseAttack.Max = 0 }
		changed = true
	}
	if c.Stats.BaseDefense.Min < 0 {
		c.Stats.BaseDefense.Min = 0
		if c.Stats.BaseDefense.Max < 0 { c.Stats.BaseDefense.Max = 0 }
		changed = true
	}
	if c.Stats.AttackCooldown.Min <= 0 {
		c.Stats.AttackCooldown = IntInterval{Min: 30, Max: 30}
		changed = true
	}
	if c.Stats.AttackRange.Min < 0 {
		c.Stats.AttackRange.Min = 0
		if c.Stats.AttackRange.Max < 0 { c.Stats.AttackRange.Max = 0 }
		changed = true
	}
	if c.Stats.ProjectileSpeed.Min < 0 {
		c.Stats.ProjectileSpeed.Min = 0
		if c.Stats.ProjectileSpeed.Max < 0 { c.Stats.ProjectileSpeed.Max = 0 }
		changed = true
	}
	if c.Stats.Age.Rate.IsZero() && c.Stats.Age.Current.IsZero() {
		c.Stats.Age.Rate = FloatInterval{Min: 1.0, Max: 1.0} 
	}
	_ = changed
}

// sanitizeObstacleArchetype validates and clamps all fields loaded from an obstacle YAML.
func sanitizeObstacleArchetype(c *ObstacleArchetype, source string) {
	if c.ID == "" {
		c.ID = "unknown"
	}
	if c.Name == "" {
		log.Printf("Warning [%s]: obstacle %q has empty name, using id", source, c.ID)
		c.Name = c.ID
	}
	if c.HealthPoints < -1 {
		log.Printf("Warning [%s]: obstacle %q has negative health (%d), clamping to 0", source, c.ID, c.HealthPoints)
		c.HealthPoints = 0
	}
	if c.CooldownTime < 0 {
		log.Printf("Warning [%s]: obstacle %q has cooldown_time=%v, clamping to 0", source, c.ID, c.CooldownTime)
		c.CooldownTime = 0
	}
}

// sanitizeMapType validates and clamps all fields loaded from a map_type YAML.
func sanitizeMapType(m *MapType, source string) {
	if m.ID == "" {
		m.ID = "unknown"
	}
	if m.Name == "" {
		log.Printf("Warning [%s]: map_type %q has empty name, using id", source, m.ID)
		m.Name = m.ID
	}
	if m.Difficulty < 0 {
		log.Printf("Warning [%s]: map_type %q has difficulty=%d, clamping to 0", source, m.ID, m.Difficulty)
		m.Difficulty = 0
	}
	if m.TargetKillCount < 0 {
		log.Printf("Warning [%s]: map_type %q has target_kill_count=%d, clamping to 0", source, m.ID, m.TargetKillCount)
		m.TargetKillCount = 0
	}
	if m.TargetTime < 0 {
		log.Printf("Warning [%s]: map_type %q has target_time=%v, clamping to 0", source, m.ID, m.TargetTime)
		m.TargetTime = 0
	}
	for i := range m.Spawns {
		s := &m.Spawns[i]
		if s.Probability < 0 {
			s.Probability = 0
		} else if s.Probability > 1.0 {
			s.Probability = 1.0
		}
		if s.Frequency < 0 {
			s.Frequency = 0
		}
	}
	if m.TargetRadius < 0 {
		log.Printf("Warning [%s]: map_type %q has target_radius=%v, clamping to 0", source, m.ID, m.TargetRadius)
		m.TargetRadius = 0
	}
}

// sanitizePlayerSaveData validates fields from a player save block.
func sanitizePlayerSaveData(p *PlayerSaveData, source string) {
	if p.HealthPoints <= 0 {
		log.Printf("Warning [%s]: player health=%d is invalid, clamping to 1", source, p.HealthPoints)
		p.HealthPoints = 1
	}
	if p.MaxHealthPoints <= 0 {
		log.Printf("Warning [%s]: player max_health=%d is invalid, clamping to 100", source, p.MaxHealthPoints)
		p.MaxHealthPoints = 100
	}
	if p.HealthPoints > p.MaxHealthPoints {
		log.Printf("Warning [%s]: player health=%d exceeds max_health=%d, clamping to max", source, p.HealthPoints, p.MaxHealthPoints)
		p.HealthPoints = p.MaxHealthPoints
	}
	if p.Level <= 0 {
		log.Printf("Warning [%s]: player level=%d is invalid, clamping to 1", source, p.Level)
		p.Level = 1
	}
	if p.XP < 0 {
		log.Printf("Warning [%s]: player xp=%d is negative, clamping to 0", source, p.XP)
		p.XP = 0
	}
	if p.Kills < 0 {
		log.Printf("Warning [%s]: player kills=%d is negative, clamping to 0", source, p.Kills)
		p.Kills = 0
	}
	if p.BaseAttack < 0 {
		log.Printf("Warning [%s]: player base_attack=%d is negative, clamping to 0", source, p.BaseAttack)
		p.BaseAttack = 0
	}
	if p.BaseDefense < 0 {
		log.Printf("Warning [%s]: player base_defense=%d is negative, clamping to 0", source, p.BaseDefense)
		p.BaseDefense = 0
	}
}

// sanitizeNPCSaveData validates NPC save data fields.
func sanitizeNPCSaveData(n *NPCSaveData, idx int, source string) {
	if n.Archetype == "" && n.NPCID == "" {
		log.Printf("Warning [%s]: NPC[%d] has empty archetype and npc_id, will be skipped", source, idx)
	}
	if n.HealthPoints < 0 {
		log.Printf("Warning [%s]: NPC[%d] %q health=%d is negative, clamping to 0", source, idx, n.Name, n.HealthPoints)
		n.HealthPoints = 0
	}
	if n.MaxHealthPoints <= 0 {
		if n.HealthPoints > 0 {
			n.MaxHealthPoints = n.HealthPoints
			log.Printf("Warning [%s]: NPC[%d] %q max_health invalid, setting to health=%d", source, idx, n.Name, n.HealthPoints)
		} else {
			n.MaxHealthPoints = 1
			log.Printf("Warning [%s]: NPC[%d] %q max_health invalid, clamping to 1", source, idx, n.Name)
		}
	}
	if n.Level <= 0 {
		log.Printf("Warning [%s]: NPC[%d] %q level=%d is invalid, clamping to 1", source, idx, n.Name, n.Level)
		n.Level = 1
	}
}
