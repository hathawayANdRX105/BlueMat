package boid

import (
	"github.com/EngoEngine/engo"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo/common"
)

var (
	green                     = color.RGBA{R: 10, G: 255, B: 50, A: 255}
	maxSpeed          float32 = 1
	radius            float32 = 2
	viewRadius        float32 = 50
	riskRadius        float32 = 10
	alignWeight       float32 = 0.1
	cohesionWeight    float32 = 0.02
	separateWeight    float32 = -0.5
	noiseWeight       float32 = 0.3
	senseWeight       float32 = 0.1
	speedUpdateWeight float32 = 0.2
	loadFactor                = 1000
	core                      = 2
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
			speed: engo.Point{X: rand.Float32()*2 - 1, Y: rand.Float32()*2 - 1},
		}

		boids = append(boids, &boid)
	}

	return boids
}

type Boid struct {
	bid       int
	speed     engo.Point
	nextSpeed engo.Point
	//	neighbors      []int
	//	collisionRisks []int
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
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

type BoidSystem struct {
	step      int
	bx, by    float32 // the boundary of the X or Y axis
	boids     []*Boid
	speedChan chan *Boid
}

func (bs *BoidSystem) Config(boids []*Boid, bx, by float32) {
	log.Printf("boids count => %d \n", len(boids))

	bs.bx = bx - radius
	bs.by = by - radius
	bs.boids = boids
	bs.speedChan = make(chan *Boid, len(boids))

	step := (len(boids) / 10) + 1
	//	if step > loadFactor {
	//		step = (len(boids) >> (core + 2)) + 1
	//	}
	bs.step = step

	var goRoutinueCount int
	for i := 0; i < len(boids); i = i + step {
		if i+step >= len(boids) {
			step = len(boids) - i
		}
		go bs.clacAcceleration(boids[i : i+step])
		goRoutinueCount++
	}

	go bs.updateSpeed()
	go bs.updateSpeed()
	go bs.updateSpeed()
	go bs.updateSpeed()
	log.Printf("each step => %d, total gorountine for boid => %d \n", step, goRoutinueCount)
}

func distance(p1 engo.Point, p2 engo.Point) float32 {
	x, y := p1.X-p2.X, p1.Y-p2.Y
	return float32(math.Sqrt(float64(x*x + y*y)))
}

func (bs *BoidSystem) clacAcceleration(boids []*Boid) {
	sleepTime := 2 * time.Millisecond
	//log.Printf("sleep time for each goroutine => %v \n", sleepTime)

	for {
		for _, boid := range boids {
			var vCount, sCount int
			var align, cohesion, separation engo.Point
			for _, b := range bs.boids {
				if d := distance(boid.Position, b.Position); b.bid != boid.bid && d < viewRadius {
					vCount++
					align.Add(b.speed)
					cohesion.Add(b.Position)

					if d < riskRadius {
						separation.Add(b.Position)
						sCount++
					}
				}
			}

			if vCount > 0 {
				n1 := 1 / float32(vCount)
				align.MultiplyScalar(n1).Subtract(boid.speed).MultiplyScalar(alignWeight)
				cohesion.MultiplyScalar(n1).Subtract(boid.Position).MultiplyScalar(cohesionWeight)
				addNoise(&align, senseWeight)
				addNoise(&cohesion, senseWeight)

				if sCount > 0 {
					n2 := 1 / float32(sCount)
					separation.MultiplyScalar(n2).Subtract(boid.Position).MultiplyScalar(separateWeight)
					addNoise(&separation, senseWeight)
				}

				align.Add(cohesion).Add(separation)
			}

			nextSpeed := boid.speed
			nextSpeed.Add(align)
			boid.nextSpeed = nextSpeed

			bs.speedChan <- boid
		}

		// 让画面渲染更加平滑
		time.Sleep(sleepTime)
	}
}

func (bs *BoidSystem) updateSpeed() {

	for boid := range bs.speedChan {
		nextSpeed := boid.nextSpeed
		// 线性增长 nextSpeed = (1 - w) * oldSpeed + w * NewSpeed
		nextSpeed = *boid.speed.MultiplyScalar(1 - speedUpdateWeight).Add(*nextSpeed.MultiplyScalar(speedUpdateWeight))
		addNoise(&nextSpeed, noiseWeight)
		limitSpeed(&nextSpeed, -maxSpeed, maxSpeed)

		boid.Position.Add(nextSpeed)
		boid.speed = nextSpeed

		bs.boudnaryBounce(boid)
	}
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

func (bs *BoidSystem) New(*ecs.World) {}

func (*BoidSystem) Remove(ecs.BasicEntity) {}

func (bs *BoidSystem) Update(float32) {}
