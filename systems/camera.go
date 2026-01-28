package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateCamera(e *ecs.ECS) {
	cameraEntry, _ := components.Camera.First(e.World)
	camera := components.Camera.Get(cameraEntry)

	// Process screen shake
	updateScreenShake(cameraEntry, camera)

	playerEntry, ok := tags.Player.First(e.World)
	if !ok {
		return // no player (could be dead), skip camera update
	}
	playerObject := components.Object.Get(playerEntry)
	playerData := components.Player.Get(playerEntry)
	physics := components.Physics.Get(playerEntry)

	// Get level dimensions for camera bounds
	levelEntry, ok := components.Level.First(e.World)
	if !ok {
		return
	}
	levelData := components.Level.Get(levelEntry)
	if levelData.CurrentLevel == nil {
		return
	}

	// Only update look-ahead when player is moving - freeze offset when idle
	if math.Abs(physics.SpeedX) > config.Camera.LookAheadSpeedThreshold {
		targetLookAhead := playerData.Direction.X * config.Camera.LookAheadDistanceX * config.Camera.LookAheadMovingScale
		camera.LookAheadX += (targetLookAhead - camera.LookAheadX) * config.Camera.LookAheadSmoothing
	}

	// Calculate target camera position (following the player with look-ahead)
	targetX := playerObject.X + camera.LookAheadX
	targetY := playerObject.Y

	// Calculate camera bounds based on screen and level dimensions
	screenWidth := float64(config.C.Width)
	screenHeight := float64(config.C.Height)
	levelWidth := float64(levelData.CurrentLevel.Width)
	levelHeight := float64(levelData.CurrentLevel.Height)

	// Camera bounds: ensure the level always fills the screen
	minCameraX := screenWidth / 2
	maxCameraX := levelWidth - screenWidth/2
	minCameraY := screenHeight / 2
	maxCameraY := levelHeight - screenHeight/2

	// Constrain target position to camera bounds
	targetX = math.Max(minCameraX, math.Min(maxCameraX, targetX))
	targetY = math.Max(minCameraY, math.Min(maxCameraY, targetY))

	// Center the camera on the constrained target position, with some smoothing.
	camera.Position.X += (targetX - camera.Position.X) * config.Camera.FollowSmoothing
	camera.Position.Y += (targetY - camera.Position.Y) * config.Camera.FollowSmoothing
}

// updateScreenShake applies screen shake offset to camera and decrements duration
func updateScreenShake(cameraEntry *donburi.Entry, camera *components.CameraData) {
	if !cameraEntry.HasComponent(components.ScreenShake) {
		return
	}

	shake := components.ScreenShake.Get(cameraEntry)
	shake.Elapsed++

	// Calculate decaying intensity
	progress := float64(shake.Duration-shake.Elapsed) / float64(shake.Duration)
	if progress < 0 {
		progress = 0
	}
	currentIntensity := shake.Intensity * progress

	// Apply oscillating offset using sine/cosine for smooth shake
	offsetX := math.Sin(float64(shake.Elapsed)*1.1) * currentIntensity
	offsetY := math.Cos(float64(shake.Elapsed)*1.3) * currentIntensity

	camera.Position.X += offsetX
	camera.Position.Y += offsetY

	// Remove component when shake is complete
	if shake.Elapsed >= shake.Duration {
		cameraEntry.RemoveComponent(components.ScreenShake)
	}
}

// TriggerScreenShake starts a screen shake effect
func TriggerScreenShake(ecs *ecs.ECS, intensity float64, duration int) {
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return
	}

	// Add or update screen shake component
	if cameraEntry.HasComponent(components.ScreenShake) {
		shake := components.ScreenShake.Get(cameraEntry)
		// Only override if new shake is stronger
		if intensity > shake.Intensity {
			shake.Intensity = intensity
			shake.Duration = duration
			shake.Elapsed = 0
		}
	} else {
		cameraEntry.AddComponent(components.ScreenShake)
		components.ScreenShake.Set(cameraEntry, &components.ScreenShakeData{
			Intensity: intensity,
			Duration:  duration,
			Elapsed:   0,
		})
	}
}
