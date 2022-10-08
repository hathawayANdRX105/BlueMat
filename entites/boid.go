package entities

import (
	"hathawayANdRX105/boids/components"
	"image/color"
	"math/rand"

	"github.com/EngoEngine/engo"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo/common"
)

var (
	green          = color.RGBA{10, 255, 50, 255}
	Radius float32 = 10
)

type Boid struct {
	BId int
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	components.MovementComponent
}

func NewBoidsSet(boidCount int, boundaryX, boundaryY float32) []*Boid {
	// loop for building entities
	boids := make([]*Boid, 0, boidCount)
	for i := 0; i < boidCount; i++ {
		boid := Boid{
			BId: i,
			BasicEntity: ecs.NewBasic(),
			SpaceComponent: common.SpaceComponent{
				Position: engo.Point{X: rand.Float32() * boundaryX, Y: rand.Float32() * boundaryY},
				Width:    Radius,
				Height:   Radius,
			},
			RenderComponent: common.RenderComponent{
				Drawable: common.Circle{},
				Color:    green,
			},
		}

		boids = append(boids, &boid)
	}

	return boids
}
