package boid

import (
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/EngoEngine/engo"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo/common"
)

var (
	green              = color.RGBA{10, 255, 50, 255}
	speed      float32 = 0.5
	radius     float32 = 10
	viewRadius float32 = 150
	adjRate    float32 = 0.0015
)

func NewBoidsSet(boidCount int, boundaryX, boundaryY float32) []*Boid {
	// loop for building entities
	var safeGap float32 = 70
	safeX, safeY := boundaryX-2*safeGap, boundaryY-2*safeGap
	boids := make([]*Boid, 0, boidCount)

	for i := 0; i < boidCount; i++ {
		boid := Boid{
			bid:         i,
			BasicEntity: ecs.NewBasic(),
			SpaceComponent: common.SpaceComponent{
				Position: engo.Point{X: rand.Float32()*safeX + safeGap, Y: rand.Float32()*safeY + safeGap},
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
	accel        engo.Point
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

func distance(p1 engo.Point, p2 engo.Point) float32 {
	x, y := p1.X-p2.X, p1.Y-p2.Y
	return float32(math.Sqrt(float64(x*x + y*y)))
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

// borderBounce ...
func borderBounce(pos float32, maxBorder float32) float32 {
	if pos < 0 {
		// bouce it and give it positive direction acceleration
		return 1 / pos
	}

	if pos+viewRadius > maxBorder {
		// bouce it and give it negative direction acceleration
		return 1 / (pos - maxBorder)
	}

	return 0
}

func (bs *BoidSystem) clacAcceleration(boid *Boid) engo.Point {
	nextMap := bs.nextMap

	// lx, ly := int(maxF32(0, boid.Position.X-viewRadius)), int(maxF32(0, boid.Position.Y-viewRadius))
	// ux, uy := int(minF32(bs.bx, boid.Position.X+viewRadius)), int(minF32(bs.by, boid.Position.Y+viewRadius))
	lx, ly := 0, 0
	ux, uy := int(bs.bx), int(bs.by)

	var avgVec, avgPoi, separateVec engo.Point

	var count int
	for lx <= ux {
		for ly <= uy {
			if nextMap[lx][ly] != nil {
				otherBoid := nextMap[lx][ly]
				d := maxF32(distance(boid.Position, otherBoid.Position), 1)
				if d < viewRadius {
					// algin
					avgVec.Add(otherBoid.speed)
					// cohesion
					avgPoi.Add(otherBoid.Position)

					// separate
					separateVec.X += (boid.Position.X - otherBoid.Position.X) / d
					separateVec.Y += (boid.Position.Y - otherBoid.Position.Y) / d

					count++
				}
			}
			ly++
		}
		lx++
	}
	accel := engo.Point{X: borderBounce(boid.Position.X, bs.bx), Y: borderBounce(boid.Position.Y, bs.by)}
	if count > 0 {
		// log.Println(borderBounce(boid.Position.X, bs.bx), borderBounce(boid.Position.Y, bs.by))
		// log.Println(avgPoi, avgVec, separateVec)
		divideV(&avgVec, float32(count))
		divideV(&avgPoi, float32(count))
		avgPoi.Subtract(boid.Position)
		avgVec.Subtract(boid.speed)

		accel.Add(avgVec).Add(avgPoi).Add(separateVec)
		log.Println(accel, boid.speed, avgPoi, avgVec, separateVec)
	}
	accel.MultiplyScalar(adjRate)

	return accel
}

func (bs *BoidSystem) Update(float32) {

	for _, boid := range bs.boids {
		// delete old position of boid
		bs.curMap[limit(0, boid.Position.X, bs.bx)][limit(0, boid.Position.Y, bs.by)] = nil

		// log.Println(accel)
		accel := bs.clacAcceleration(boid)
		boid.speed.Add(accel)

		bs.boudnaryCollision(boid)

		boid.Position = boid.nextPosition
		boid.nextPosition.Add(boid.speed)
		// log.Println(boid.speed, boid.Position, boid.nextPosition)

		// add new position of boid
		// log.Println(limit(0, boid.Position.X, bs.bx), limit(0, boid.Position.Y, bs.by))
		bs.curMap[limit(0, boid.Position.X, bs.bx)][limit(0, boid.Position.Y, bs.by)] = boid
	}

	bs.curMap, bs.nextMap = bs.nextMap, bs.curMap

	// time.Sleep(5 * time.Millisecond)
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
