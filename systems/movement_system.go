package systems

import (
	"hathawayANdRX105/boids/components"
	entities "hathawayANdRX105/boids/entites"
	"log"
	"math/rand"
	"time"

	"github.com/EngoEngine/engo"

	"github.com/EngoEngine/ecs"
)

var (
	speed      float32 = 8
	viewRadius float32 = 13
	adjRate    float32 = 0.015
)

type MovementSystem struct {
	world   *ecs.World
	boids   []*entities.Boid
	x, y    float32
	viewMap [][]*entities.Boid
}

func (ms *MovementSystem) New(world *ecs.World) {
	ms.world = world
}

func (ms *MovementSystem) FillComponent(b *entities.Boid) {
	b.MovementComponent = components.MovementComponent{
		Point: engo.Point{X: (rand.Float32()*2 - 1) * speed, Y: (rand.Float32()*2 - 1) * speed},
	}

	ms.viewMap[int(b.SpaceComponent.Position.X)][int(b.SpaceComponent.Position.Y)] = b
}

func (ms *MovementSystem) Config(boids []*entities.Boid, x, y float32) {
	ms.boids = boids
	ms.x = x
	ms.y = y

	row, col := int(x), int(y)
	viewMap := make([][]*entities.Boid, 0, row)

	for row > 0 {
		row--
		viewMap = append(viewMap, make([]*entities.Boid, col))
	}
	ms.viewMap = viewMap
}

func (*MovementSystem) Remove(ecs.BasicEntity) {}

func min(x, y float32) int {
	if x < y {
		return int(x)
	}
	return int(y)
}

func max(x, y float32) int {
	if x > y {
		return int(x)
	}
	return int(y)
}

func limit(x, y, z float32) int {
	if y < x {
		return int(x)
	}

	if y > z {
		return int(z)
	}

	return int(y)
}

type accelImpact struct {
	engo.Point
	count int
}

// clacAcceleration ...
func (ms *MovementSystem) calcAcceleration() [][]accelImpact {
	x, y := int(ms.x), int(ms.y)
	accels := make([][]accelImpact, 0, x)
	for i := 0; i < x; i++ {
		accels = append(accels, make([]accelImpact, y))
	}

	for _, b := range ms.boids {
		sx, sy := max(b.Position.X-viewRadius, 0), max(b.Position.Y-viewRadius, 0)
		ex, ey := min(ms.x-1, b.Position.X+viewRadius), min(ms.y-1, b.Position.Y+viewRadius)

		for ; sx <= ex; sx++ {
			for ; sy <= ey; sy++ {
				// log.Println(b.Point, b.MovementComponent.Point)
				accels[sx][sy].Point = *accels[sx][sy].Point.Add(b.Point).Add(b.MovementComponent.Point)
				// accels[sx][sy].Point.Add(b.Point).Add(b.MovementComponent.Point)
				accels[sx][sy].count++
				// log.Print(accels[sx][sy], sx, sy)
			}
			log.Println(accels[sx])
		}
	}

	return accels
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

func (ms *MovementSystem) Update(float32) {

	ms.calcAcceleration()

	// for _, boid := range ms.boids {
	// 	space, speed := &boid.Position, &boid.Point

	// 	ms.viewMap[limit(0, space.X, ms.x)][limit(0, space.Y, ms.y)] = nil
	// 	accel := accels[limit(0, space.X, ms.x)][limit(0, space.Y, ms.y)]
	// 	log.Println(accel, accel.count)
	// 	multiplyV(divideV(&accel.Point, float32(accel.count)), adjRate)
	// 	log.Println(accel, accel.count)
	// 	space.Add(*speed)

	// 	if space.X < 0 || space.X > ms.x {
	// 		speed.X = -speed.X
	// 	}

	// 	if space.Y < 0 || space.Y > ms.y {
	// 		speed.Y = -speed.Y
	// 	}

	// 	ms.viewMap[limit(0, space.X, ms.x)][limit(0, space.Y, ms.y)] = boid
	// }

	time.Sleep(5 * time.Millisecond)
}
