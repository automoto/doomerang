package components

import "github.com/yohamta/donburi"

type FinishLineData struct {
	Activated bool
}

var FinishLine = donburi.NewComponentType[FinishLineData]()
