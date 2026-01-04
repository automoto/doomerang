package components

import "github.com/yohamta/donburi"

type LivesData struct {
	Lives    int
	MaxLives int
}

var Lives = donburi.NewComponentType[LivesData]()
