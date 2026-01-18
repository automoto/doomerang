package config

type AnimationDef struct {
	First int
	Last  int
	Step  int
	Speed float32
}

// CharacterAnimations maps a character key (e.g., "player")
// to its specific set of animation definitions.
var CharacterAnimations = map[string]map[StateID]AnimationDef{
	"player": {
		Crouch:                 {First: 0, Last: 5, Step: 1, Speed: 5},
		Die:                    {First: 0, Last: 8, Step: 1, Speed: 5},
		Guard:                  {First: 0, Last: 0, Step: 1, Speed: 10},
		GuardImpact:            {First: 0, Last: 2, Step: 1, Speed: 5},
		Hit:                    {First: 0, Last: 2, Step: 1, Speed: 5},
		Idle:                   {First: 0, Last: 6, Step: 1, Speed: 5},
		Jump:                   {First: 0, Last: 2, Step: 1, Speed: 10},
		Kick01:                 {First: 0, Last: 8, Step: 1, Speed: 4},
		Kick02:                 {First: 0, Last: 7, Step: 1, Speed: 3},
		Kick03:                 {First: 0, Last: 8, Step: 1, Speed: 5},
		Knockback:              {First: 0, Last: 5, Step: 1, Speed: 5},
		Ledge:                  {First: 0, Last: 7, Step: 1, Speed: 5},
		LedgeGrab:              {First: 0, Last: 4, Step: 1, Speed: 5},
		Punch01:                {First: 0, Last: 5, Step: 1, Speed: 4},
		Punch02:                {First: 0, Last: 3, Step: 1, Speed: 5},
		Punch03:                {First: 0, Last: 6, Step: 1, Speed: 5},
		Running:                {First: 0, Last: 7, Step: 1, Speed: 5},
		Stunned:                {First: 0, Last: 6, Step: 1, Speed: 5},
		Throw:                  {First: 0, Last: 4, Step: 1, Speed: 5},
		Walk:                   {First: 0, Last: 7, Step: 1, Speed: 5},
		WallSlide:              {First: 0, Last: 5, Step: 1, Speed: 5},
		StateChargingBoomerang: {First: 0, Last: 0, Step: 1, Speed: 0},
		StateChargingAttack:    {First: 0, Last: 0, Step: 1, Speed: 0}, // Use idle frame while charging
		StateSliding:           {First: 0, Last: 3, Step: 1, Speed: 6}, // Custom slide animation (4 frames)
		// Dust effects (in player spritesheet directory, 96x84 frames)
		StateJumpDust:  {First: 0, Last: 6, Step: 1, Speed: 3},
		StateLandDust:  {First: 0, Last: 5, Step: 1, Speed: 3},
		StateSlideDust: {First: 0, Last: 5, Step: 1, Speed: 3},
	},
	// SFX animations (in sfx spritesheet directory, various frame sizes)
	"sfx": {
		StateExplosionShort: {First: 0, Last: 24, Step: 1, Speed: 2}, // 57x56 frames
		StatePlasma:         {First: 0, Last: 23, Step: 1, Speed: 2}, // 32x43 frames
		StateGunshot:        {First: 0, Last: 12, Step: 1, Speed: 2}, // 46x26 frames
		HitExplosion:        {First: 0, Last: 64, Step: 4, Speed: 1}, // 47x57 frames, fast playback
		ChargeUp:            {First: 25, Last: 49, Step: 1, Speed: 2}, // 102x135 frames, second half only
	},
	// Obstacle animations (in obstacle spritesheet directory)
	"obstacle": {
		FirePulsing:    {First: 0, Last: 44, Step: 1, Speed: 3}, // 65x43 frames, 45 frames total
		FireContinuous: {First: 0, Last: 24, Step: 1, Speed: 3}, // 66x47 frames, 25 frames total
	},
}
