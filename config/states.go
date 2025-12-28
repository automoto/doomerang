package config

type StateID int

const (
	// Character animation states
	Idle StateID = iota
	Crouch
	Die
	Guard
	GuardImpact
	Hit
	Jump
	Kick01
	Kick02
	Kick03
	Knockback
	Ledge
	LedgeGrab
	Punch01
	Punch02
	Punch03
	Running
	Stunned
	Throw
	Walk
	WallSlide

	// Combat specific states
	StateAttackingPunch
	StateAttackingKick
	StateChargingAttack
	StateAttackingJump

	// Enemy AI states
	StatePatrol
	StateChase
)

// StateToFileName maps StateID to the corresponding filename prefix.
var StateToFileName = map[StateID]string{
	Idle:        "idle",
	Crouch:      "crouch",
	Die:         "die",
	Guard:       "guard",
	GuardImpact: "guardimpact",
	Hit:         "hit",
	Jump:        "jump",
	Kick01:      "kick01",
	Kick02:      "kick02",
	Kick03:      "kick03",
	Knockback:   "knockback",
	Ledge:       "ledge",
	LedgeGrab:   "ledgegrab",
	Punch01:     "punch01",
	Punch02:     "punch02",
	Punch03:     "punch03",
	Running:     "running",
	Stunned:     "stunned",
	Throw:       "throw",
	Walk:        "walk",
	WallSlide:   "wallslide",
	
	// Map combat states to animation files where appropriate
	StateAttackingPunch: "punch01",
	StateAttackingKick:  "kick01",
	StateAttackingJump:  "kick02", // Jump kick uses kick02 animation
	StateChargingAttack: "idle",   // Charging uses idle animation (or maybe a specific one later)
	
	// Enemy AI states map to movement animations
	StatePatrol: "walk",
	StateChase:  "running",
}

func (s StateID) String() string {
	if name, ok := StateToFileName[s]; ok {
		return name
	}
	return "unknown"
}
