package components

import "github.com/yohamta/donburi"

// MessagePointData stores the message ID and position for a message point entity
type MessagePointData struct {
	MessageID float64
	X, Y      float64
}

var MessagePoint = donburi.NewComponentType[MessagePointData]()

// MessageStateData is a singleton tracking the active message
type MessageStateData struct {
	ActiveMessageID float64 // Currently displayed message (0 = none)
	DisplayTimer    int     // Frames remaining to display current message
}

var MessageState = donburi.NewComponentType[MessageStateData]()
