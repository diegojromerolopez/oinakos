package game

import (
	"log"
)

// sanitizeEntityConfig validates and clamps all fields loaded from an archetype YAML.
// Any invalid value is fixed and a warning is logged.
func sanitizeEntityConfig(c *EntityConfig, source string) {
	changed := false

	if c.ID == "" {
		log.Printf("Warning [%s]: archetype has empty id, using 'unknown'", source)
		c.ID = "unknown"
		changed = true
	}
	if c.Name == "" {
		log.Printf("Warning [%s]: archetype %q has empty name, using id", source, c.ID)
		c.Name = c.ID
		changed = true
	}
	if c.Stats.HealthMin <= 0 {
		log.Printf("Warning [%s]: archetype %q has health_min=%d, clamping to 1", source, c.ID, c.Stats.HealthMin)
		c.Stats.HealthMin = 1
		changed = true
	}
	if c.Stats.HealthMax < c.Stats.HealthMin {
		log.Printf("Warning [%s]: archetype %q has health_max=%d < health_min=%d, clamping to health_min",
			source, c.ID, c.Stats.HealthMax, c.Stats.HealthMin)
		c.Stats.HealthMax = c.Stats.HealthMin
		changed = true
	}
	if c.Stats.Speed <= 0 {
		log.Printf("Warning [%s]: archetype %q has speed=%v, clamping to 0.01", source, c.ID, c.Stats.Speed)
		c.Stats.Speed = 0.01
		changed = true
	}
	if c.Stats.Speed > 1.0 {
		log.Printf("Warning [%s]: archetype %q has speed=%v (suspiciously high), clamping to 0.5", source, c.ID, c.Stats.Speed)
		c.Stats.Speed = 0.5
		changed = true
	}
	if c.Stats.BaseAttack < 0 {
		log.Printf("Warning [%s]: archetype %q has base_attack=%d, clamping to 0", source, c.ID, c.Stats.BaseAttack)
		c.Stats.BaseAttack = 0
		changed = true
	}
	if c.Stats.BaseDefense < 0 {
		log.Printf("Warning [%s]: archetype %q has base_defense=%d, clamping to 0", source, c.ID, c.Stats.BaseDefense)
		c.Stats.BaseDefense = 0
		changed = true
	}
	if c.Stats.AttackCooldown <= 0 {
		log.Printf("Warning [%s]: archetype %q has attack_cooldown=%d, clamping to 30", source, c.ID, c.Stats.AttackCooldown)
		c.Stats.AttackCooldown = 30
		changed = true
	}
	if c.Stats.AttackRange < 0 {
		log.Printf("Warning [%s]: archetype %q has attack_range=%v, clamping to 0", source, c.ID, c.Stats.AttackRange)
		c.Stats.AttackRange = 0
		changed = true
	}
	if c.Stats.ProjectileSpeed < 0 {
		log.Printf("Warning [%s]: archetype %q has projectile_speed=%v, clamping to 0", source, c.ID, c.Stats.ProjectileSpeed)
		c.Stats.ProjectileSpeed = 0
		changed = true
	}
	if c.FootprintWidth <= 0 && len(c.Footprint) == 0 {
		log.Printf("Warning [%s]: archetype %q has no footprint_width and no custom footprint, clamping to 0.3", source, c.ID)
		c.FootprintWidth = 0.3
		changed = true
	}
	if c.FootprintHeight <= 0 && len(c.Footprint) == 0 {
		log.Printf("Warning [%s]: archetype %q has no footprint_height and no custom footprint, clamping to 0.3", source, c.ID)
		c.FootprintHeight = 0.3
		changed = true
	}
	_ = changed
}

// sanitizeObstacleArchetype validates and clamps all fields loaded from an obstacle YAML.
func sanitizeObstacleArchetype(c *ObstacleArchetype, source string) {
	if c.ID == "" {
		log.Printf("Warning [%s]: obstacle has empty id, using 'unknown'", source)
		c.ID = "unknown"
	}
	if c.Name == "" {
		log.Printf("Warning [%s]: obstacle %q has empty name, using id", source, c.ID)
		c.Name = c.ID
	}
	if c.Health < 0 {
		log.Printf("Warning [%s]: obstacle %q has health=%d (negative), clamping to 0 (indestructible)", source, c.ID, c.Health)
		c.Health = 0
	}
	if c.Scale <= 0 {
		log.Printf("Warning [%s]: obstacle %q has scale=%v, clamping to 1.0", source, c.ID, c.Scale)
		c.Scale = 1.0
	}
	if c.Scale > 10 {
		log.Printf("Warning [%s]: obstacle %q has scale=%v (suspiciously large), clamping to 10", source, c.ID, c.Scale)
		c.Scale = 10
	}
	if c.FootprintWidth <= 0 && len(c.Footprint) == 0 {
		log.Printf("Warning [%s]: obstacle %q has no footprint_width and no custom footprint, clamping to 0.3", source, c.ID)
		c.FootprintWidth = 0.3
	}
	if c.FootprintHeight <= 0 && len(c.Footprint) == 0 {
		log.Printf("Warning [%s]: obstacle %q has no footprint_height and no custom footprint, clamping to 0.3", source, c.ID)
		c.FootprintHeight = 0.3
	}
	if c.CooldownTime < 0 {
		log.Printf("Warning [%s]: obstacle %q has cooldown_time=%v, clamping to 0", source, c.ID, c.CooldownTime)
		c.CooldownTime = 0
	}
}

// sanitizeMapType validates and clamps all fields loaded from a map_type YAML.
func sanitizeMapType(m *MapType, source string) {
	if m.ID == "" {
		log.Printf("Warning [%s]: map_type has empty id, using 'unknown'", source)
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
	if m.SpawnFreq < 0 {
		log.Printf("Warning [%s]: map_type %q has spawn_frequency=%v, clamping to 0", source, m.ID, m.SpawnFreq)
		m.SpawnFreq = 0
	}
	if m.SpawnAmount < 0 {
		log.Printf("Warning [%s]: map_type %q has spawn_amount=%d, clamping to 0", source, m.ID, m.SpawnAmount)
		m.SpawnAmount = 0
	}
	if m.TargetRadius < 0 {
		log.Printf("Warning [%s]: map_type %q has target_radius=%v, clamping to 0", source, m.ID, m.TargetRadius)
		m.TargetRadius = 0
	}
}

// sanitizePlayerSaveData validates fields from a player save block.
func sanitizePlayerSaveData(p *PlayerSaveData, source string) {
	if p.Health < 0 {
		log.Printf("Warning [%s]: player health=%d is negative, clamping to 1", source, p.Health)
		p.Health = 1
	}
	if p.MaxHealth <= 0 {
		log.Printf("Warning [%s]: player max_health=%d is invalid, clamping to 100", source, p.MaxHealth)
		p.MaxHealth = 100
	}
	if p.Health > p.MaxHealth {
		log.Printf("Warning [%s]: player health=%d exceeds max_health=%d, clamping to max", source, p.Health, p.MaxHealth)
		p.Health = p.MaxHealth
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
	if n.ArchetypeID == "" {
		log.Printf("Warning [%s]: NPC[%d] has empty archetype_id, will be skipped", source, idx)
	}
	if n.Health < 0 {
		log.Printf("Warning [%s]: NPC[%d] %q health=%d is negative, clamping to 0", source, idx, n.Name, n.Health)
		n.Health = 0
	}
	if n.MaxHealth <= 0 {
		if n.Health > 0 {
			n.MaxHealth = n.Health
			log.Printf("Warning [%s]: NPC[%d] %q max_health invalid, setting to health=%d", source, idx, n.Name, n.Health)
		} else {
			n.MaxHealth = 1
			log.Printf("Warning [%s]: NPC[%d] %q max_health invalid, clamping to 1", source, idx, n.Name)
		}
	}
	if n.Level <= 0 {
		log.Printf("Warning [%s]: NPC[%d] %q level=%d is invalid, clamping to 1", source, idx, n.Name, n.Level)
		n.Level = 1
	}
}
