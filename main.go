package main

import (
	entities "hathawayANdRX105/boids/entites"
	"hathawayANdRX105/boids/systems"
	"image/color"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

var (
	boidCount         = 10
	BoundaryY float32 = 600
	BoundaryX float32 = 1200
)

type myScene struct{}

// Type uniquely defines your game type
func (*myScene) Type() string { return "boids" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*myScene) Preload() {}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (*myScene) Setup(updater engo.Updater) {
	world, _ := updater.(*ecs.World)
	common.SetBackground(color.Black)

	boids := entities.NewBoidsSet(boidCount, BoundaryX, BoundaryY)
	var rs common.RenderSystem
	var ms systems.MovementSystem
	ms.Config(boids, BoundaryX, BoundaryY)

	world.AddSystem(&rs)
	world.AddSystem(&ms)

	for _, b := range boids {
		rs.Add(&b.BasicEntity, &b.RenderComponent, &b.SpaceComponent)
		ms.FillComponent(b)
	}

}

func main() {

	opts := engo.RunOptions{
		Title:  "boids model",
		Width:  int(BoundaryX),
		Height: int(BoundaryY),
	}
	engo.Run(opts, &myScene{})
}
