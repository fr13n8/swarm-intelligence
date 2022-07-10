package main

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	Pi  = math.Pi
	Tau = Pi * 2

	signalTimer       = 3
	signalRange       = 70
	maxResourcesCount = 3
	agentsCount       = 1500
)

type Circle struct {
	x, y, r float64
	c       color.RGBA
}

type Resources []*Circle
type Pedestal []*Circle

var (
	BorderColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	AgentColor  = color.RGBA{R: 56, G: 190, B: 255, A: 255}
	LineColor   = color.RGBA{R: 255, G: 255, B: 255, A: 10}

	resourcesCircles = Resources{
		&Circle{x: 550, y: 100, r: 20, c: color.RGBA{R: 223, G: 250, B: 90, A: 255}},
	}
	pedestalCircles = Pedestal{
		&Circle{x: 100, y: 550, r: 5, c: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	}
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type Agent struct {
	x, y, angle, speed float64
	counter            int
	angR, angP         float64
	disR, disP         int
	goal               int
}

func NewAgent(x float64, y float64, angle float64) *Agent {
	fRange := FloatRange{0, Tau}
	sRange := FloatRange{0.1, 1}
	cRange := FloatRange{0, signalTimer - 2}
	gRange := FloatRange{0, 10}
	goal := 0
	if gRange.NextRandom(r) > 5 {
		goal = 1
	}
	return &Agent{
		x:       x,
		y:       y,
		angle:   angle,
		speed:   sRange.NextRandom(r),
		angR:    fRange.NextRandom(r),
		angP:    fRange.NextRandom(r),
		disR:    100,
		disP:    100,
		counter: int(cRange.NextRandom(r)),
		goal:    goal,
	}
}

func (a *Agent) NextStep(world *World) {
	a.disR += 1
	a.disP += 1
	aRange := FloatRange{-0.05, 0.05}
	a.angle += aRange.NextRandom(r)

	nextX := a.x + a.speed*math.Sin(a.angle)
	nextY := a.y + a.speed*math.Cos(a.angle)

	if nextX < 0 || int(nextX) >= world.Width || nextY < 0 || int(nextY) >= world.Height {
		a.angle += Pi
		a.x += math.Sin(a.angle)
		a.y += math.Cos(a.angle)
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

	a.counter += 1
	if a.counter == signalTimer/2 {
		a.SignalR(world, a.disR+signalRange)
	}
	if a.counter == signalTimer {
		a.SignalP(world, a.disP+signalRange)
		a.counter = 0
	}
}

func (a *Agent) SignalR(world *World, distance int) {
	for _, agent := range world.Agents {
		if agent.disR > distance {
			if CheckIfPointInsideCircle(agent.x, agent.y, []*Circle{{x: a.x, y: a.y, r: signalRange}}) {
				dist := a.Distance(agent)
				if dist <= signalRange {
					agent.SetAngleToR(a.x, a.y, dist)
					agent.disR = distance
					if agent.goal == 0 {
						// ebitenutil.DrawLine(world.Screen, a.x, a.y, agent.x, agent.y, LineColor)
						agent.angle = agent.angR
					}
				}
			}
		}
	}
}

func (a *Agent) SignalP(world *World, distance int) {
	for _, agents := range world.Agents {
		if agents.disP > distance {
			if CheckIfPointInsideCircle(agents.x, agents.y, []*Circle{{x: a.x, y: a.y, r: signalRange}}) {
				dist := a.Distance(agents)
				if dist <= signalRange {
					agents.SetAngleToP(a.x, a.y, dist)
					agents.disP = distance
					if agents.goal == 1 {
						// ebitenutil.DrawLine(world.Screen, a.x, a.y, agents.x, agents.y, LineColor)
						agents.angle = agents.angP
					}
				}
			}
		}
	}
}

func (a *Agent) Distance(c2 *Agent) float64 {
	return math.Hypot(c2.x-a.x, c2.y-a.y)
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
	ebitenutil.DrawRect(screen, a.x, a.y, 1, 1, AgentColor)
}

func (c *Circle) SetPosition(x float64, y float64) {
	c.x = x
	c.y = y
}

type FloatRange struct {
	min, max float64
}

func (fr *FloatRange) NextRandom(r *rand.Rand) float64 {
	return fr.min + r.Float64()*(fr.max-fr.min)
}

func CheckIfPointInsideCircle(x, y float64, circles []*Circle) bool {
	for _, circle := range circles {
		if math.Pow(x-circle.x, 2)+math.Pow(y-circle.y, 2) < math.Pow(circle.r, 2) {
			return true
		}
	}
	return false
}

func GetCircleIn(x, y float64, circles []*Circle) *Circle {
	for _, circle := range circles {
		if math.Pow(x-circle.x, 2)+math.Pow(y-circle.y, 2) < math.Pow(circle.r, 2) {
			return circle
		}
	}
	return nil
}
