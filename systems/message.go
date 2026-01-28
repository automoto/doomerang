package systems

import (
	"math"
	"strings"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"golang.org/x/image/font"
)

// Cached font face for message rendering (lazy initialized)
var messageFontFace font.Face

// UpdateMessage checks player proximity to message points and activates messages
func UpdateMessage(ecs *ecs.ECS) {
	state := getOrCreateMessageState(ecs)

	// Decrement display timer if active
	if state.DisplayTimer > 0 {
		state.DisplayTimer--
		if state.DisplayTimer == 0 {
			state.ActiveMessageID = 0
		}
		// Don't check for new messages while one is displaying
		return
	}

	// Get player position (use center of collision box)
	playerEntry, ok := tags.Player.First(ecs.World)
	if !ok {
		return
	}
	playerObj := components.Object.Get(playerEntry)
	playerCenterX := playerObj.X + playerObj.W/2
	playerCenterY := playerObj.Y + playerObj.H/2

	components.MessagePoint.Each(ecs.World, func(entry *donburi.Entry) {
		msg := components.MessagePoint.Get(entry)

		// Skip if already showing a message
		if state.ActiveMessageID != 0 {
			return
		}

		// Calculate distance to player center
		dx := playerCenterX - msg.X
		dy := playerCenterY - msg.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist <= cfg.Message.ActivationRadius {
			// Activate this message
			state.ActiveMessageID = msg.MessageID
			state.DisplayTimer = cfg.Message.DisplayDuration
		}
	})
}

// DrawMessage renders the active message at the top center of the screen
func DrawMessage(ecs *ecs.ECS, screen *ebiten.Image) {
	state := getOrCreateMessageState(ecs)

	if state.ActiveMessageID == 0 {
		return
	}

	// Get message text
	messageText, ok := cfg.Message.Messages[state.ActiveMessageID]
	if !ok {
		return
	}

	// Resolve placeholders based on input method
	input := getOrCreateInput(ecs)
	resolvedText := resolvePlaceholders(messageText, input.LastInputMethod)

	// Lazy initialize cached font face
	if messageFontFace == nil {
		messageFontFace = fonts.ExcelBold.Get()
	}

	// Measure text
	bounds := text.BoundString(messageFontFace, resolvedText)
	textWidth := bounds.Dx()
	textHeight := bounds.Dy()

	// Calculate box dimensions
	padding := cfg.Message.BoxPadding
	boxWidth := float32(textWidth) + float32(padding)*2
	boxHeight := float32(textHeight) + float32(padding)*2

	// Position at top center
	screenWidth := float64(screen.Bounds().Dx())
	boxX := float32((screenWidth - float64(boxWidth)) / 2)
	boxY := float32(cfg.Message.TopMargin)

	// Draw semi-transparent background box
	vector.DrawFilledRect(
		screen,
		boxX, boxY,
		boxWidth, boxHeight,
		cfg.Message.BoxColor,
		false,
	)

	// Draw text centered in box
	textX := int(boxX + float32(padding))
	textY := int(boxY + float32(padding) + float32(textHeight))
	text.Draw(screen, resolvedText, messageFontFace, textX, textY, cfg.Message.TextColor)
}

// resolvePlaceholders replaces {placeholder} tokens with input-specific labels
func resolvePlaceholders(text string, inputMethod components.InputMethod) string {
	var labels map[string]string

	switch inputMethod {
	case components.InputPlayStation:
		labels = cfg.Message.PlayStationLabels
	case components.InputXbox:
		labels = cfg.Message.XboxLabels
	default:
		labels = cfg.Message.KeyboardLabels
	}

	result := text
	for placeholder, label := range labels {
		result = strings.ReplaceAll(result, "{"+placeholder+"}", label)
	}

	return result
}

// ResetMessageState clears the active message (call on respawn)
func ResetMessageState(ecs *ecs.ECS) {
	state := getOrCreateMessageState(ecs)
	state.ActiveMessageID = 0
	state.DisplayTimer = 0
}

// getOrCreateMessageState returns the singleton MessageState component
func getOrCreateMessageState(ecs *ecs.ECS) *components.MessageStateData {
	entry, ok := components.MessageState.First(ecs.World)
	if !ok {
		entry = ecs.World.Entry(ecs.World.Create(components.MessageState))
		components.MessageState.SetValue(entry, components.MessageStateData{
			ActiveMessageID: 0,
		})
	}
	return components.MessageState.Get(entry)
}
