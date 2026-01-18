package tags

import "github.com/yohamta/donburi"

var (
	Player           = donburi.NewTag().SetName("Player")
	Platform         = donburi.NewTag().SetName("Platform")
	FloatingPlatform = donburi.NewTag().SetName("FloatingPlatform")
	Wall             = donburi.NewTag().SetName("Wall")
	Enemy            = donburi.NewTag().SetName("Enemy")
	Hitbox           = donburi.NewTag().SetName("Hitbox")
	Boomerang        = donburi.NewTag().SetName("Boomerang")
	Checkpoint       = donburi.NewTag().SetName("Checkpoint")
)

// Resolv tags for physics collision
const (
	ResolvSolid      = "solid"
	ResolvRamp       = "ramp"
	ResolvPlayer     = "Player"
	ResolvEnemy      = "Enemy"
	ResolvBoomerang  = "Boomerang"
	ResolvDeadZone   = "deadzone"
	ResolvCheckpoint = "checkpoint"

	// Slope type tags
	Slope45UpRight = "45_up_right"
	Slope45UpLeft  = "45_up_left"
)
