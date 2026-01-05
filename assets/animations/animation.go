package animations

type Animation struct {
	First          int
	Last           int
	Step           int     // how many indices do we move per frame
	SpeedInTps     float32 // how many ticks before next frame
	frameCounter   float32
	frame          int
	Looped         bool
	FreezeOnComplete bool // If true, stay on last frame instead of looping
}

func (a *Animation) Update() {
	a.frameCounter -= 1.0
	if a.frameCounter < 0.0 {
		a.frameCounter = a.SpeedInTps
		a.frame += a.Step
		if a.frame > a.Last {
			a.Looped = true
			if a.FreezeOnComplete {
				// Stay on last frame
				a.frame = a.Last
			} else {
				// loop back to the beginning
				a.frame = a.First
			}
		}
	}
}

func (a *Animation) Frame() int {
	return a.frame
}

func (a *Animation) Restart() {
	a.frame = a.First
	a.frameCounter = a.SpeedInTps
}

func NewAnimation(first, last, step int, speed float32) *Animation {
	return &Animation{
		First:        first,
		Last:         last,
		Step:         step,
		SpeedInTps:   speed,
		frameCounter: speed,
		frame:        first,
		Looped:       false,
	}
}
