package boid

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/EngoEngine/engo"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo/common"
)

var (
	green            = color.RGBA{10, 255, 50, 255}
	maxSpeed float32 = 1
	adjRate  float32 = 0.015
	// initRate          float32 = 0.25
	radius            float32 = 3
	viewRadius        float32 = 50
	riskRadius        float32 = 10
	alignWeight       float32 = 0.1
	cohesionWeight    float32 = 0.02
	separateWeight    float32 = -0.5
	noiseWeight       float32 = 0.3
	senseWeight       float32 = 0.1
	speedUpdateWeight float32 = 0.2
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
			speed: engo.Point{X: (rand.Float32()*2 - 1), Y: (rand.Float32()*2 - 1)},
		}

		boids = append(boids, &boid)
	}

	return boids
}

type Boid struct {
	bid            int
	speed          engo.Point
	neighbors      []int
	collisionRisks []int
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// lookAround ...
func (b *Boid) lookAround(boids []*Boid) {
	b.neighbors = b.neighbors[:0]
	b.collisionRisks = b.collisionRisks[:0]

	for _, otherBoid := range boids {
		if otherBoid.bid != b.bid {
			dist := distance(otherBoid.Position, b.Position)
			if dist < viewRadius {
				b.neighbors = append(b.neighbors, otherBoid.bid)
			}
			if dist < riskRadius {
				b.collisionRisks = append(b.collisionRisks, otherBoid.bid)
			}
		}
	}
}

func limitSpeed(speed *engo.Point, lower, high float32) {
	if speed.X < lower {
		speed.X = lower
	}

	if speed.X > high {
		speed.X = high
	}

	if speed.Y < lower {
		speed.Y = lower
	}

	if speed.Y > high {
		speed.Y = high
	}
}

func addNoise(p *engo.Point, weight float32) {
	p.X += ((rand.Float32() * 2) - 1) * weight
	p.Y += ((rand.Float32() * 2) - 1) * weight
}

func (b *Boid) getAvgVecAndAvgPoi(boids []*Boid) (engo.Point, engo.Point) {
	var avgVec, avgPoi engo.Point

	for _, id := range b.neighbors {
		avgVec.Add(boids[id].speed)
		avgPoi.Add(boids[id].Position)
	}

	n := float32(len(b.neighbors))
	avgVec.MultiplyScalar(1 / n)
	avgPoi.MultiplyScalar(1 / n)
	addNoise(&avgVec, senseWeight)
	addNoise(&avgPoi, senseWeight)

	return avgVec, avgPoi
}

func (b *Boid) getAvgSep(boids []*Boid) engo.Point {
	var avgSep engo.Point
	for _, id := range b.collisionRisks {
		avgSep.Add(boids[id].Position)
	}

	n := float32(len(b.collisionRisks))
	avgSep.MultiplyScalar(1 / n)
	addNoise(&avgSep, senseWeight)

	return avgSep
}

func (b *Boid) Start(bs *BoidSystem) {
	// time.Sleep(30 * time.Millisecond)
	boids := bs.boids

	for {
		b.lookAround(boids)

		nextSpeed := b.speed

		if len(b.neighbors) > 0 {
			avgVec, avgPoi := b.getAvgVecAndAvgPoi(boids)
			avgVec.Subtract(b.speed).MultiplyScalar(alignWeight)
			avgPoi.Subtract(b.Position).MultiplyScalar(cohesionWeight)
			nextSpeed.Add(avgVec).Add(avgPoi)
			// log.Printf("bid:%d => a:%v, c:%v, n:%v\n", b.bid, avgVec, avgPoi, nextSpeed)
		}
		// log.Printf("bid:%d => p: %v, s: %v \n", b.bid, nextSpeed, b.speed)

		if len(b.collisionRisks) > 0 {
			avgSep := b.getAvgSep(boids)
			avgSep.Subtract(b.Position).MultiplyScalar(separateWeight)
			nextSpeed.Add(avgSep)
			// log.Printf("bid:%d => s:%v n:%v\n", b.bid, avgSep, nextSpeed)
		}

		addNoise(&nextSpeed, noiseWeight)
		// newSpeed = (1 - w) * oldSpeed + w * nextSpeed
		nextSpeed = *b.speed.MultiplyScalar(1 - speedUpdateWeight).Add(*nextSpeed.MultiplyScalar(speedUpdateWeight))
		limitSpeed(&nextSpeed, -maxSpeed, maxSpeed)
		// log.Printf("bid:%d => p: %v, s: %v \n", b.bid, nextSpeed, b.speed)

		b.Position.Add(nextSpeed)
		b.speed = nextSpeed
		bs.boudnaryBounce(b)

		// bs.boudnaryMirror(b)

		// log.Printf("bid:%d => p: %v, s: %v \n", b.bid, b.Position, b.speed)
		time.Sleep(5 * time.Millisecond)
	}
}

type BoidSystem struct {
	bx, by float32 // the boundary of the X or Y axis
	boids  []*Boid
}

func (bs *BoidSystem) Config(boids []*Boid, bx, by float32) {

	bs.bx = bx - radius
	bs.by = by - radius
	bs.boids = boids
	log.Printf("boids count => %d \n", len(boids))

	for _, b := range boids {
		go b.Start(bs)
	}
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

func (bs *BoidSystem) clacAcceleration(boid *Boid) {
	var avgPoi, separateVec engo.Point
	avgVec := engo.Point{X: 0, Y: 0}

	var count int
	for _, b := range bs.boids {
		if b.bid != boid.bid {
			if d := distance(boid.Position, b.Position); d < viewRadius {
				// algin
				avgVec.Add(b.speed)
				// cohesion
				avgPoi.Add(b.Position)

				// separate
				separateVec.X += (b.Position.X - boid.Position.X) / d
				separateVec.Y += (b.Position.Y - boid.Position.Y) / d

				count++
			}
		}
	}

	if count > 0 {
		divideV(&avgVec, float32(count))
		divideV(&avgPoi, float32(count))
		avgPoi.Subtract(boid.Position)
		avgVec.Subtract(boid.speed)

		avgVec.Add(avgPoi).Add(separateVec)
	}

	avgVec.MultiplyScalar(adjRate)

	boid.speed.Add(avgVec)
	// log.Println(boid.speed, avgPoi, avgVec, separateVec)
}

func (bs *BoidSystem) Update(float32) {
}

func (bs *BoidSystem) boudnaryMirror(boid *Boid) {
	position := &boid.Position

	if position.X < 0 {
		position.X = bs.bx
	}

	if position.X > bs.bx {
		position.X = 0
	}

	if position.Y < 0 {
		position.Y = bs.by
	}

	if position.Y > bs.by {
		position.Y = 0
	}
}

func (bs *BoidSystem) boudnaryBounce(boid *Boid) {
	position := boid.Position

	if position.X < viewRadius {
		boid.speed.X += viewRadius / position.X
	}

	if position.X > bs.bx-viewRadius {
		boid.speed.X -= viewRadius / (bs.bx - position.X)
	}

	if position.Y < viewRadius {
		boid.speed.Y += viewRadius / position.Y
	}

	if position.Y > bs.by-viewRadius {
		boid.speed.Y -= viewRadius / (bs.by - position.Y)
	}
}
