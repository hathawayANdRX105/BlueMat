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
	speed      float32 = 3
	radius     float32 = 10
	viewRadius float32 = 13
	adjRate    float32 = 0.015
)

func NewBoidsSet(boidCount int, boundaryX, boundaryY float32) []*Boid {
	// loop for building entities
	boids := make([]*Boid, 0, boidCount)
	for i := 0; i < boidCount; i++ {
		boid := Boid{
			bid:         i,
			BasicEntity: ecs.NewBasic(),
			SpaceComponent: common.SpaceComponent{
				Position: engo.Point{X: rand.Float32() * boundaryX, Y: rand.Float32() * boundaryY},
				Width:    radius,
				Height:   radius,
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
	bid          int
	speed        engo.Point
	nextPosition engo.Point
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type BoidSystem struct {
	bx, by  float32 // the boundary of the X or Y axis
	boids   []*Boid
	curMap  [][]*Boid
	nextMap [][]*Boid
}

func (bs *BoidSystem) Config(boids []*Boid, bx, by float32) {
	row, col := int(bx), int(by)
	curMap := make([][]*Boid, 0, row)
	nextMap := make([][]*Boid, 0, row)

	for row > 0 {
		row--
		curMap = append(curMap, make([]*Boid, col))
		nextMap = append(nextMap, make([]*Boid, col))
	}

	for _, b := range boids {
		x, y := b.Position.X+b.speed.X, b.Position.Y+b.speed.Y
		b.nextPosition = engo.Point{X: x, Y: y}
		curMap[int(b.Position.X)][int(b.Position.Y)] = b
		nextMap[int(x)][int(y)] = b
	}

	bs.bx = bx - radius
	bs.by = by - radius
	bs.boids = boids
	bs.curMap = curMap
	bs.nextMap = nextMap
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

func addV(p *engo.Point, y float32) {
	p.X += y
	p.Y += y
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

func minF32(a, b float32) float32 {
	if a < b {
		return a
	}

	return b
}

func maxF32(a, b float32) float32 {
	if a > b {
		return a
	}

	return b
}

func (bs *BoidSystem) addSeparateImpact() {

}

func (bs *BoidSystem) addCohesionImpact() {

}

func (bs *BoidSystem) clacAcceleration(boid *Boid) engo.Point {
	nextMap := bs.nextMap

	lx, ly := int(maxF32(0, boid.Position.X-viewRadius)), int(maxF32(0, boid.Position.Y-viewRadius))
	ux, uy := int(minF32(bs.bx, boid.Position.X+viewRadius)), int(minF32(bs.by, boid.Position.Y+viewRadius))

	var avgVec, avgPoi engo.Point

	var count int
	for lx <= ux {
		for ly <= uy {
			if nextMap[lx][ly] != nil {
				avgVec.Add(nextMap[lx][ly].speed)
				avgPoi.Add(nextMap[lx][ly].Position)
				count++
			}
			ly++
		}
		lx++
	}
	divideV(&avgVec, float32(count))
	divideV(&avgPoi, float32(count))
	multiplyV(avgVec.Add(avgPoi), adjRate)
	return avgVec
}

// change boid speed direction
func (bs *BoidSystem) boudnaryCollision(boid *Boid) {
	nextPosition, speed := &boid.nextPosition, &boid.speed

	if nextPosition.X < 0 || nextPosition.X > bs.bx {
		speed.X = -speed.X
	}

	if nextPosition.Y < 0 || nextPosition.Y > bs.by {
		speed.Y = -speed.Y
	}
}

func (bs *BoidSystem) Update(float32) {

	for _, boid := range bs.boids {
		// delete old position of boid
		bs.curMap[limit(0, boid.Position.X, bs.bx)][limit(0, boid.Position.Y, bs.by)] = nil

		accel := bs.clacAcceleration(boid)
		multiplyV(accel.Subtract(boid.speed), adjRate)
		boid.speed.Add(accel)

		bs.addSeparateImpact()
		bs.boudnaryCollision(boid)

		boid.Position = boid.nextPosition
		boid.nextPosition.Add(boid.speed)
		// add new position of boid
		bs.curMap[limit(0, boid.Position.X, bs.bx)][limit(0, boid.Position.Y, bs.by)] = boid
	}

	bs.curMap, bs.nextMap = bs.nextMap, bs.curMap

	time.Sleep(5 * time.Millisecond)
}
