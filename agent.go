package main

import (
	"image/color"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	Pi                = math.Pi
	Tau               = Pi * 2
	signalTimer       = 3
	signalRange       = 70
	maxResourcesCount = 3
	agentsCount       = 1500

	signalRangeSq = signalRange * signalRange
)

type Circle struct {
	x, y, r  float64
	rSquared float64
	c        color.RGBA
}

func NewCircle(x, y, r float64, c color.RGBA) *Circle {
	return &Circle{
		x:        x,
		y:        y,
		r:        r,
		rSquared: r * r,
		c:        c,
	}
}

type Resources []*Circle
type Pedestal []*Circle

var (
	BorderColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	AgentColor  = color.RGBA{R: 56, G: 190, B: 255, A: 255}
	LineColor   = color.RGBA{R: 255, G: 255, B: 255, A: 10}

	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	randMutex  sync.Mutex

	resourcesCircles = Resources{
		NewCircle(550, 100, 20, color.RGBA{R: 223, G: 250, B: 90, A: 255}),
	}
	pedestalCircles = Pedestal{
		NewCircle(100, 350, 5, color.RGBA{R: 255, G: 255, B: 255, A: 255}),
	}
)

var circleSyncPool = sync.Pool{
	New: func() interface{} {
		return &Circle{r: signalRange, rSquared: signalRangeSq}
	},
}

type Agent struct {
	x, y, angle, speed float64
	counter            int
	angR, angP         float64
	disR, disP         int
	goal               int
}

func GetNextRandom(min, max float64) float64 {
	randMutex.Lock()
	defer randMutex.Unlock()
	return min + globalRand.Float64()*(max-min)
}

func NewAgent(x float64, y float64, angle float64) *Agent {
	goal := 0
	if GetNextRandom(0, 10) > 5 {
		goal = 1
	}
	return &Agent{
		x:       x,
		y:       y,
		angle:   angle,
		speed:   GetNextRandom(0.1, 1),
		angR:    GetNextRandom(0, Tau),
		angP:    GetNextRandom(0, Tau),
		disR:    100,
		disP:    100,
		counter: int(GetNextRandom(0, signalTimer-2)),
		goal:    goal,
	}
}

func (a *Agent) NextStep(world *World) {
	a.disR++
	a.disP++
	a.angle += GetNextRandom(-0.05, 0.05)

	sin, cos := math.Sincos(a.angle)
	nextX := a.x + a.speed*sin
	nextY := a.y + a.speed*cos

	if nextX < 0 || int(nextX) >= world.Width || nextY < 0 || int(nextY) >= world.Height {
		a.angle += Pi
		sin, cos = math.Sincos(a.angle)
		a.x += sin
		a.y += cos
	} else {
		a.x = nextX
		a.y = nextY
	}

	if CheckIfPointInsideCircle(a.x, a.y, resourcesCircles) {
		a.goal = 1
		a.angle = a.angP
		a.disR = 0
	}
	if CheckIfPointInsideCircle(a.x, a.y, pedestalCircles) {
		a.goal = 0
		a.angle = a.angR
		a.disP = 0
	}

	a.counter++
	if a.counter == signalTimer/2 {
		a.SignalR(world, a.disR+signalRange)
	}
	if a.counter == signalTimer {
		a.SignalP(world, a.disP+signalRange)
		a.counter = 0
	}
}

func (a *Agent) SignalR(world *World, distance int) {
	tempCircle := circleSyncPool.Get().(*Circle)
	tempCircle.x = a.x
	tempCircle.y = a.y

	world.signalWorkerPool.Add(1)
	go func() {
		defer world.signalWorkerPool.Done()
		defer circleSyncPool.Put(tempCircle)

		for _, agent := range world.Agents {
			if agent.disR > distance {
				dx, dy := agent.x-a.x, agent.y-a.y
				distSq := dx*dx + dy*dy
				if distSq <= signalRangeSq {
					dist := math.Sqrt(distSq)
					agent.SetAngleToR(a.x, a.y, dist)
					agent.disR = distance
					if agent.goal == 0 {
						agent.angle = agent.angR
					}
				}
			}
		}
	}()
}

func (a *Agent) SignalP(world *World, distance int) {
	tempCircle := circleSyncPool.Get().(*Circle)
	tempCircle.x = a.x
	tempCircle.y = a.y

	world.signalWorkerPool.Add(1)
	go func() {
		defer world.signalWorkerPool.Done()
		defer circleSyncPool.Put(tempCircle)

		for _, agent := range world.Agents {
			if agent.disP > distance {
				dx, dy := agent.x-a.x, agent.y-a.y
				distSq := dx*dx + dy*dy
				if distSq <= signalRangeSq {
					dist := math.Sqrt(distSq)
					agent.SetAngleToP(a.x, a.y, dist)
					agent.disP = distance
					if agent.goal == 1 {
						agent.angle = agent.angP
					}
				}
			}
		}
	}()
}

func (a *Agent) Distance(c2 *Agent) float64 {
	dx, dy := c2.x-a.x, c2.y-a.y
	return math.Sqrt(dx*dx + dy*dy)
}

func (a *Agent) SetAngleToR(x float64, y float64, dist float64) {
	an := math.Acos((y - a.y) / dist)
	if x > a.x {
		a.angR = an
	} else {
		a.angR = Tau - an
	}
}

func (a *Agent) SetAngleToP(x float64, y float64, dist float64) {
	an := math.Acos((y - a.y) / dist)
	if x > a.x {
		a.angP = an
	} else {
		a.angP = Tau - an
	}
}

func (a *Agent) Render(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, float32(a.x), float32(a.y), 1, 1, AgentColor, true)
}

func (c *Circle) SetPosition(x float64, y float64) {
	c.x = x
	c.y = y
}

func CheckIfPointInsideCircle(x, y float64, circles []*Circle) bool {
	for _, circle := range circles {
		dx, dy := x-circle.x, y-circle.y
		if dx*dx+dy*dy < circle.rSquared {
			return true
		}
	}
	return false
}

func GetCircleIn(x, y float64, circles []*Circle) *Circle {
	for _, circle := range circles {
		dx, dy := x-circle.x, y-circle.y
		if dx*dx+dy*dy < circle.rSquared {
			return circle
		}
	}
	return nil
}
