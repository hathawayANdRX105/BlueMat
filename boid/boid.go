package boid

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/EngoEngine/engo"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo/common"
)

var (
	green              = color.RGBA{10, 255, 50, 255}
	Radius     float32 = 10
	speed      float32 = 8
	viewRadius float32 = 13
	adjRate    float32 = 0.015
)

func NewBoidsSet(boidCount int, boundaryX, boundaryY float32) []*Boid {
	// loop for building entities
	boids := make([]*Boid, 0, boidCount)
	for i := 0; i < boidCount; i++ {
		boid := Boid{
			BId:         i,
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
			speed: engo.Point{X: (rand.Float32()*2 - 1) * speed, Y: (rand.Float32()*2 - 1) * speed},
		}

		boids = append(boids, &boid)
	}

	return boids
}

type Boid struct {
	BId   int
	speed engo.Point
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type BoidSystem struct {
	bx, by  float32 // the boundary of the X or Y axis
	boids   []*Boid
	viewMap [][]*Boid
}

func (bs *BoidSystem) Config(boids []*Boid, bx, by float32) {
	row, col := int(bx), int(by)
	viewMap := make([][]*Boid, 0, row)

	for row > 0 {
		viewMap = append(viewMap, make([]*Boid, col))
		row--
	}

	bs.bx = bx
	bs.by = by
	bs.viewMap = viewMap
	bs.boids = boids
}

func (bs *BoidSystem) New(*ecs.World) {}

func (*BoidSystem) Remove(ecs.BasicEntity) {}

func limit(x, y, z float32) int {
	if y < x {
		return int(x)
	}

	if y > z {
		return int(z)
	}

	return int(y)
}

func divideV(p *engo.Point, y float32) *engo.Point {
	if y < 1 {
		return p
	}
	p.X /= y
	p.Y /= y
	return p
}

func multiplyV(p *engo.Point, y float32) *engo.Point {
	p.X *= y
	p.Y *= y
	return p
}

func (bs *BoidSystem) addSeparateImpact() {

}

func (bs *BoidSystem) addCohesionImpact() {

}

func (bs *BoidSystem) addAlignImpact() {

}

func (bs *BoidSystem) boudnaryCollision() {
	for _, boid := range bs.boids {
		space, speed := &boid.Position, &boid.speed

		space.Add(*speed)

		if space.X < 0 || space.X > bs.bx {
			speed.X = -speed.X
		}

		if space.Y < 0 || space.Y > bs.by {
			speed.Y = -speed.Y
		}

	}

}

func (bs *BoidSystem) Update(float32) {
	bs.addAlignImpact()
	bs.addCohesionImpact()
	bs.addSeparateImpact()
	bs.boudnaryCollision()
	time.Sleep(5 * time.Millisecond)
}
