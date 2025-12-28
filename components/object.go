package components

import (
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
)

type ObjectData struct {
	*resolv.Object
}

var Object = donburi.NewComponentType[ObjectData]()
