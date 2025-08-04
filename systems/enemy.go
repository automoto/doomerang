package systems

import (
	"image"
	"image/color"
	"math"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const (
	enemyGravity  = 0.75 // Same as player
	enemyMaxSpeed = 6.0  // Same as player
	enemyFriction = 0.5  // Same as player
)

// AI state constants - should match factory
const (
	enemyStatePatrol = "patrol"
	enemyStateChase  = "chase"
	enemyStateAttack = "attack"
)

func UpdateEnemies(ecs *ecs.ECS) {
	// Get player position for AI decisions
	playerEntry, _ := components.Player.First(ecs.World)
	var playerObject *resolv.Object
	if playerEntry != nil {
		playerObject = cfg.GetObject(playerEntry)
	}

	tags.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		// Skip if enemy is in death sequence
		if e.HasComponent(components.Death) {
			if anim := components.Animation.Get(e); anim != nil && anim.CurrentAnimation != nil {
				anim.CurrentAnimation.Update()
			}
			return
		}

		// Skip if invulnerable
		enemy := components.Enemy.Get(e)
		if enemy.InvulnFrames > 0 {
			enemy.InvulnFrames--
		}

		// Update AI behavior
		updateEnemyAI(enemy, cfg.GetObject(e), playerObject)

		// Apply physics
		applyEnemyPhysics(enemy)

		// Resolve collisions
		resolveEnemyCollisions(enemy, cfg.GetObject(e))

		// Update animation state
		updateEnemyAnimation(enemy, components.Animation.Get(e))
	})
}

func updateEnemyAI(enemy *components.EnemyData, enemyObject, playerObject *resolv.Object) {
	enemy.StateTimer++

	// Update attack cooldown
	if enemy.AttackCooldown > 0 {
		enemy.AttackCooldown--
	}

	// No AI if no player
	if playerObject == nil {
		return
	}

	// Calculate distance to player
	distanceToPlayer := math.Abs(playerObject.X - enemyObject.X)

	// State machine
	switch enemy.CurrentState {
	case enemyStatePatrol:
		handlePatrolState(enemy, enemyObject, playerObject, distanceToPlayer)
	case enemyStateChase:
		handleChaseState(enemy, enemyObject, playerObject, distanceToPlayer)
	case enemyStateAttack:
		handleAttackState(enemy, enemyObject, playerObject, distanceToPlayer)
	}
}

func handlePatrolState(enemy *components.EnemyData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	// Check if should start chasing
	if distanceToPlayer <= enemy.ChaseRange {
		enemy.CurrentState = enemyStateChase
		enemy.StateTimer = 0
		return
	}

	// Patrol behavior - move back and forth
	if enemy.FacingRight {
		enemy.SpeedX = enemy.PatrolSpeed
		// Turn around if hit right boundary
		if enemyObject.X >= enemy.PatrolRight {
			enemy.FacingRight = false
		}
	} else {
		enemy.SpeedX = -enemy.PatrolSpeed
		// Turn around if hit left boundary
		if enemyObject.X <= enemy.PatrolLeft {
			enemy.FacingRight = true
		}
	}
}

func handleChaseState(enemy *components.EnemyData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	// Check if should attack
	if distanceToPlayer <= enemy.AttackRange && enemy.AttackCooldown == 0 {
		enemy.CurrentState = enemyStateAttack
		enemy.StateTimer = 0
		enemy.SpeedX = 0 // Stop moving when attacking
		return
	}

	// Check if should stop chasing (player too far)
	if distanceToPlayer > enemy.ChaseRange*1.5 { // Hysteresis to prevent flapping
		enemy.CurrentState = enemyStatePatrol
		enemy.StateTimer = 0
		return
	}

	// Chase player
	if playerObject.X > enemyObject.X {
		enemy.SpeedX = enemy.ChaseSpeed
		enemy.FacingRight = true
	} else {
		enemy.SpeedX = -enemy.ChaseSpeed
		enemy.FacingRight = false
	}
}

func handleAttackState(enemy *components.EnemyData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	// Attack animation duration (simplified - using timer)
	attackDuration := 30 // 30 frames for attack

	if enemy.StateTimer >= attackDuration {
		// Attack finished
		enemy.CurrentState = enemyStateChase
		enemy.StateTimer = 0
		enemy.AttackCooldown = 60 // 1 second cooldown
		return
	}

	// Don't move during attack
	enemy.SpeedX = 0
}

func applyEnemyPhysics(enemy *components.EnemyData) {
	// Apply friction
	if enemy.SpeedX > enemyFriction {
		enemy.SpeedX -= enemyFriction
	} else if enemy.SpeedX < -enemyFriction {
		enemy.SpeedX += enemyFriction
	} else {
		enemy.SpeedX = 0
	}

	// Limit horizontal speed
	if enemy.SpeedX > enemyMaxSpeed {
		enemy.SpeedX = enemyMaxSpeed
	} else if enemy.SpeedX < -enemyMaxSpeed {
		enemy.SpeedX = -enemyMaxSpeed
	}

	// Apply gravity
	enemy.SpeedY += enemyGravity
	if enemy.SpeedY > 16 {
		enemy.SpeedY = 16
	}
}

func resolveEnemyCollisions(enemy *components.EnemyData, enemyObject *resolv.Object) {
	// Horizontal collision - simplified (just stop on walls)
	dx := enemy.SpeedX
	if dx != 0 {
		if check := enemyObject.Check(dx, 0, "solid"); check != nil {
			enemy.SpeedX = 0
			dx = 0
		}
	}
	enemyObject.X += dx

	// Vertical collision - simplified ground detection
	enemy.OnGround = nil
	dy := enemy.SpeedY
	dy = math.Max(math.Min(dy, 16), -16)

	if check := enemyObject.Check(0, dy, "solid", "platform"); check != nil {
		if dy > 0 { // Falling down
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				dy = check.ContactWithObject(solids[0]).Y()
				enemy.OnGround = solids[0]
				enemy.SpeedY = 0
			} else if platforms := check.ObjectsByTags("platform"); len(platforms) > 0 {
				platform := platforms[0]
				if enemy.SpeedY >= 0 && enemyObject.Bottom() < platform.Y+4 {
					dy = check.ContactWithObject(platform).Y()
					enemy.OnGround = platform
					enemy.SpeedY = 0
				}
			}
		} else { // Moving up
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				dy = check.ContactWithObject(solids[0]).Y()
				enemy.SpeedY = 0
			}
		}
	}
	enemyObject.Y += dy
}

func updateEnemyAnimation(enemy *components.EnemyData, animData *components.AnimationData) {
	// Simple animation state based on movement and AI state
	var targetState string

	switch enemy.CurrentState {
	case enemyStateAttack:
		targetState = cfg.Punch01 // Use punch animation for attacks
	default:
		if enemy.OnGround == nil {
			targetState = cfg.Jump
		} else if enemy.SpeedX != 0 {
			targetState = cfg.Running
		} else {
			targetState = cfg.Idle
		}
	}

	// Update animation if changed
	if animData.CurrentAnimation != animData.Animations[targetState] {
		animData.SetAnimation(targetState)
	}

	if animData.CurrentAnimation != nil {
		animData.CurrentAnimation.Update()
	}
}

func DrawEnemies(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	tags.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		enemy := components.Enemy.Get(e)
		o := cfg.GetObject(e)
		animData := components.Animation.Get(e)

		if animData.CurrentAnimation != nil && animData.SpriteSheets[animData.CurrentSheet] != nil {
			// Calculate the source rectangle for the current frame
			frame := animData.CurrentAnimation.Frame()
			sx := frame * animData.FrameWidth
			sy := 0
			srcRect := image.Rect(sx, sy, sx+animData.FrameWidth, sy+animData.FrameHeight)

			// Create draw options
			op := &ebiten.DrawImageOptions{}

			// Anchor the sprite at its bottom-center
			op.GeoM.Translate(-float64(animData.FrameWidth)/2, -float64(animData.FrameHeight))

			// Flip sprite if facing left
			if !enemy.FacingRight {
				op.GeoM.Scale(-1, 1)
			}

			// Position sprite
			op.GeoM.Translate(o.X+o.W/2, o.Y+o.H)

			// Apply camera translation
			op.GeoM.Translate(float64(width)/2-camera.Position.X, float64(height)/2-camera.Position.Y)

			// Flicker effect if invulnerable
			if enemy.InvulnFrames > 0 && enemy.InvulnFrames%4 < 2 {
				op.ColorM.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
			}

			screen.DrawImage(animData.SpriteSheets[animData.CurrentSheet].SubImage(srcRect).(*ebiten.Image), op)
		} else {
			// Fallback rectangle (red for enemies)
			enemyColor := color.RGBA{255, 60, 60, 255}
			if enemy.OnGround == nil {
				enemyColor = color.RGBA{255, 0, 255, 255}
			}
			vector.DrawFilledRect(screen, float32(o.X), float32(o.Y), float32(o.W), float32(o.H), enemyColor, false)
		}
	})
}
