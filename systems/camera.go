package systems

import (
    "github.com/automoto/doomerang/components"
    "github.com/automoto/doomerang/tags"
    "github.com/yohamta/donburi/ecs"
)

func UpdateCamera(ecs *ecs.ECS) {
    cameraEntry, _ := components.Camera.First(ecs.World)
    camera := components.Camera.Get(cameraEntry)

    playerEntry, _ := tags.Player.First(ecs.World)
    playerObject := components.Object.Get(playerEntry)

    // Center the camera on the player, with some smoothing.
    camera.Position.X += (playerObject.X - camera.Position.X) * 0.1
    camera.Position.Y += (playerObject.Y - camera.Position.Y) * 0.1
}
