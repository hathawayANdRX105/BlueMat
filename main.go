package main

import (
	"hathawayANdRX105/boids/boid"
	"image/color"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

var (
	boidCount         = 8000
	BoundaryX float32 = 1600
	BoundaryY float32 = 1200
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
	var (
		rs common.RenderSystem
		ms boid.BoidSystem
	)

	world, _ := updater.(*ecs.World)
	common.SetBackground(color.Black)

	boids := boid.NewBoidsSet(boidCount, BoundaryX, BoundaryY)

	world.AddSystem(&rs)
	world.AddSystem(&ms)

	for _, b := range boids {
		rs.Add(&b.BasicEntity, &b.RenderComponent, &b.SpaceComponent)
	}

	ms.Config(boids, BoundaryX, BoundaryY)

}

func main() {

	opts := engo.RunOptions{
		Title:  "boids model",
		Width:  int(BoundaryX),
		Height: int(BoundaryY),
	}
	engo.Run(opts, &myScene{})
}
