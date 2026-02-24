package game

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math"
	"math/rand"
	"os"

	"oinakos/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

type PlayerState int

const (
	StateIdle PlayerState = iota
	StateWalking
	StateAttacking
	StateDead
)

type Direction int

const (
	DirSE Direction = iota
	DirSW
	DirNE
	DirNW
)

type Player struct {
	X, Y          float64
	Config        *EntityConfig
	Speed         float64
	SpriteSheet   *ebiten.Image
	AttackSprite  *ebiten.Image
	CorpseSprite  *ebiten.Image
	Facing        Direction
	State         PlayerState
	Tick          int
	Health        int
	MaxHealth     int
	Kills         int
	MapKills      map[string]int
	XP            int
	Level         int
	BaseAttack    int
	BaseDefense   int
	Weapon        *Weapon
	EquippedArmor map[ArmorSlot]*Armor
}

func loadPlayerImage(assets fs.FS, path string) (image.Image, error) {
	var f fs.File
	var err error

	if assets != nil {
		f, err = assets.Open(path)
	}
	if err != nil || assets == nil {
		f, err = os.Open(path)
	}

	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func NewPlayer(x, y float64, config *EntityConfig) *Player {
	p := &Player{
		X:            x,
		Y:            y,
		Config:       config,
		Facing:       DirSE,
		State:        StateIdle,
		Health:       config.Stats.HealthMin,
		MaxHealth:    config.Stats.HealthMin,
		Speed:        config.Stats.Speed,
		MapKills:     make(map[string]int),
		BaseAttack:   config.Stats.BaseAttack,
		BaseDefense:  config.Stats.BaseDefense,
		Weapon:       config.Weapon,
		SpriteSheet:  config.StaticImage,
		AttackSprite: config.AttackImage,
		CorpseSprite: config.CorpseImage,
		EquippedArmor: map[ArmorSlot]*Armor{
			SlotBody: ArmorLeather,
		},
	}
	// Random quality bonus for starting weapon
	if p.Weapon != nil {
		p.Weapon.Bonus = rand.Intn(4) // 0 to 3
	}
	return p
}

func (p *Player) GetTotalAttack() int {
	return p.calculateStat(p.BaseAttack, p.Level)
}

func (p *Player) GetTotalDefense() int {
	return p.calculateStat(p.BaseDefense, p.Level)
}

func (p *Player) GetTotalProtection() int {
	total := 0
	for _, a := range p.EquippedArmor {
		if a != nil {
			total += a.Protection
		}
	}
	return total
}

func (p *Player) calculateStat(base, level int) int {
	// Logarithmic scaling: stat = base + log2(level) * scalingFactor
	// scalingFactor = 10 for meaningful growth
	if level <= 1 {
		return base
	}
	bonus := int(math.Log2(float64(level)) * 10)
	return base + bonus
}

func (p *Player) AddXP(amount int) {
	p.XP += amount
	// Simple level up logic: level = XP / 100 + 1
	newLevel := p.XP/100 + 1
	if newLevel > p.Level {
		p.Level = newLevel
		// Optionally heal on level up
		p.Health = p.MaxHealth
	}
}

func (p *Player) TakeDamage(amount int) {
	if p.State == StateDead {
		return
	}
	p.Health -= amount
	// engine.PlaySound("knight_hit") // Knight remains silent as requested
	if p.Health <= 0 {
		p.Health = 0
		p.State = StateDead
		engine.PlaySound("knight_death")
	}
}

func (p *Player) IsAlive() bool {
	return p.State != StateDead
}

func (p *Player) GetFootprint() engine.Polygon {
	return engine.Polygon{Points: []engine.Point{
		{X: -0.2, Y: -0.1}, {X: 0.2, Y: -0.1}, {X: 0.3, Y: 0}, {X: 0.2, Y: 0.1}, {X: -0.2, Y: 0.1}, {X: -0.3, Y: 0},
	}}.Transformed(p.X, p.Y)
}

func (p *Player) checkCollisionAt(newX, newY float64, obstacles []*Obstacle) bool {
	pFootprint := engine.Polygon{Points: []engine.Point{
		{X: -0.2, Y: -0.1}, {X: 0.2, Y: -0.1}, {X: 0.3, Y: 0}, {X: 0.2, Y: 0.1}, {X: -0.2, Y: 0.1}, {X: -0.3, Y: 0},
	}}.Transformed(newX, newY)

	for _, o := range obstacles {
		if !o.Alive {
			continue
		}
		if engine.CheckCollision(pFootprint, o.GetFootprint()) {
			return true
		}
	}
	return false
}

func (p *Player) Update(obstacles []*Obstacle, npcs []*NPC, fts *[]*FloatingText) {
	if p.State == StateDead {
		return
	}

	if p.State == StateAttacking {
		p.Tick++
		if p.Tick == 15 {
			p.CheckAttackHits(npcs, obstacles, fts)
		}
		if p.Tick > 30 {
			p.State = StateIdle
			p.Tick = 0
		}
		return
	}

	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		dy -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		dy += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		dx -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		dx += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.State = StateAttacking
		p.Tick = 0
		engine.PlaySound("knight_attack") // Sword swing/woosh
		return
	}

	if dx != 0 || dy != 0 {
		p.State = StateWalking
		p.Tick++

		mag := math.Sqrt(dx*dx + dy*dy)
		dx /= mag
		dy /= mag

		moveX := dx * p.Speed
		moveY := dy * p.Speed

		if !p.checkCollisionAt(p.X+moveX, p.Y+moveY, obstacles) {
			p.X += moveX
			p.Y += moveY
		} else {
			if !p.checkCollisionAt(p.X+moveX, p.Y, obstacles) {
				p.X += moveX
			}
			if !p.checkCollisionAt(p.X, p.Y+moveY, obstacles) {
				p.Y += moveY
			}
		}

		if dx > 0 {
			if dy < 0 {
				p.Facing = DirNE
			} else if dy > 0 {
				p.Facing = DirSE
			} else {
				p.Facing = DirSE // Default for purely horizontal right
			}
		} else if dx < 0 {
			if dy < 0 {
				p.Facing = DirNW
			} else if dy > 0 {
				p.Facing = DirSW
			} else {
				p.Facing = DirSW // Default for purely horizontal left
			}
		} else {
			// Purely vertical movement
			if dy < 0 {
				p.Facing = DirNE // Up-Right in isometric
			} else if dy > 0 {
				p.Facing = DirSW // Down-Left in isometric
			}
		}
	} else {
		p.State = StateIdle
		p.Tick = 0
	}
}

func (p *Player) CheckAttackHits(npcs []*NPC, obstacles []*Obstacle, fts *[]*FloatingText) {
	attackDist := 0.9
	atX, atY := p.X, p.Y

	// Fix: Normalize attack center based on facing.
	// SE is "right" in isometric view generally (X increases, Y increases)
	// NE is X increases, Y decreases.
	// We want the attack to land in front of the player.
	switch p.Facing {
	case DirSE:
		atX += attackDist
		atY += attackDist * 0.5
	case DirSW:
		atX -= attackDist * 0.5
		atY += attackDist
	case DirNE:
		atX += attackDist
		atY -= attackDist * 0.5
	case DirNW:
		atX -= attackDist * 0.5
		atY -= attackDist
	}

	for _, n := range npcs {
		if !n.IsAlive() {
			continue
		}
		// Generous circle check around the attack center
		dist := math.Sqrt(math.Pow(atX-n.X, 2) + math.Pow(atY-n.Y, 2))
		if dist < 1.6 { // Increased range for better feel
			// HIT ROLL
			hitChance := p.GetTotalAttack() - n.GetTotalDefense()
			if hitChance < 5 {
				hitChance = 5 // Cap at 5% min
			}
			if hitChance > 95 {
				hitChance = 95 // Cap at 95% max
			}

			roll := rand.Intn(100) + 1
			if roll <= hitChance {
				// SUCCESSFUL HIT
				rawDmg := p.Weapon.RollDamage()
				protection := n.GetTotalProtection()
				finalDmg := int(math.Max(1, float64(rawDmg-protection)))
				n.TakeDamage(finalDmg, p, nil)

				*fts = append(*fts, &FloatingText{
					Text:  fmt.Sprintf("%d", finalDmg),
					X:     n.X,
					Y:     n.Y,
					Life:  45,
					Color: color.RGBA{255, 0, 0, 255},
				})
			} else {
				// MISS
				*fts = append(*fts, &FloatingText{
					Text:  "MISS",
					X:     n.X,
					Y:     n.Y,
					Life:  45,
					Color: color.RGBA{200, 200, 200, 255},
				})
			}
		}
	}
}

func (p *Player) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	isoX, isoY := engine.CartesianToIso(p.X, p.Y)

	drawSprite := p.SpriteSheet
	if p.State == StateAttacking {
		drawSprite = p.AttackSprite
	} else if p.State == StateDead {
		drawSprite = p.CorpseSprite
	}

	if drawSprite == nil {
		return
	}

	w, h := drawSprite.Size()
	op := &ebiten.DrawImageOptions{}
	scale := 0.25

	flip := 1.0
	if p.Facing == DirSE || p.Facing == DirNE {
		flip = -1.0
	}

	op.GeoM.Scale(scale*flip, scale)

	tx := isoX + offsetX
	if flip < 0 {
		tx += float64(w) * scale / 2
	} else {
		tx -= float64(w) * scale / 2
	}

	ty := isoY + offsetY - float64(h)*scale*0.85

	// Procedural Animation Overrides
	if p.State == StateDead {
		ty = isoY + offsetY - float64(h)*scale*0.5
	} else if p.State == StateWalking {
		// Bobbing effect
		bob := math.Sin(float64(p.Tick)*0.3) * 3.0
		ty += bob
	} else if p.State == StateAttacking {
		// Lunge effect (move forward then back)
		lungeAmt := 0.0
		if p.Tick < 15 {
			lungeAmt = (float64(p.Tick) / 15.0) * 5.0
		} else {
			lungeAmt = (float64(30-p.Tick) / 15.0) * 5.0
		}

		if flip < 0 {
			tx += lungeAmt // Lunge right
		} else {
			tx -= lungeAmt // Lunge left
		}
	}

	op.GeoM.Translate(tx, ty)
	screen.DrawImage(drawSprite, op)
}
