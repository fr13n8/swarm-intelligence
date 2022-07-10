package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"golang.org/x/image/math/f64"

	"github.com/fr13n8/swarm-intelligence/camera"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type World struct {
	Agents   []*Agent
	Width    int
	Height   int
	camera   camera.Camera
	LifeTime int
	Paused   bool
	Screen   *ebiten.Image
}

func NewWorld(width int, height int) *World {
	world := &World{
		Width:  width,
		Height: height,
		camera: camera.Camera{ViewPort: f64.Vec2{ScreenWidth, ScreenHeight}},
		Paused: true,
	}
	return world
}

func (w *World) GetCursorCoordinates() (int, int) {
	worldX, worldY := w.camera.ScreenToWorld(ebiten.CursorPosition())
	x, y := int(worldX), int(worldY)
	return x, y
}

func (w *World) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := w.GetCursorCoordinates()
		if y <= w.Height && x <= w.Width && y >= 0 && x >= 0 {
			if c := GetCircleIn(float64(x), float64(y), resourcesCircles); c != nil {
				c.SetPosition(float64(x), float64(y))
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := w.GetCursorCoordinates()

			if y <= w.Height && x <= w.Width && y >= 0 && x >= 0 {
				if len(resourcesCircles) < maxResourcesCount {
					resourcesCircles = append(resourcesCircles, &Circle{x: float64(x), y: float64(y), r: 20, c: color.RGBA{R: 223, G: 250, B: 90, A: 255}})
				}
				return nil
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyControlLeft) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := w.GetCursorCoordinates()

			if y <= w.Height && x <= w.Width && y >= 0 && x >= 0 {
				if CheckIfPointInsideCircle(float64(x), float64(y), resourcesCircles) {
					resourcesCircles = RemoveCircle(resourcesCircles, GetCircleIn(float64(x), float64(y), resourcesCircles))
				}
			}
		}
		return nil
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		x, y := w.GetCursorCoordinates()
		if y <= w.Height && x <= w.Width && y >= 0 && x >= 0 {
			if c := GetCircleIn(float64(x), float64(y), pedestalCircles); c != nil {
				c.SetPosition(float64(x), float64(y))
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		w.Next()
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		w.Paused = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeyNumpadEnter) {
		w.Paused = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		w.camera.Position[0] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		w.camera.Position[0] += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		w.camera.Position[1] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		w.camera.Position[1] += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		if w.camera.ZoomFactor > -2400 {
			w.camera.ZoomFactor -= 1
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		if w.camera.ZoomFactor < 2400 {
			w.camera.ZoomFactor += 1
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		w.camera.Rotation += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		w.camera.Reset()
	}

	return nil
}

func (w *World) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	background := ebiten.NewImage(ScreenWidth, ScreenHeight)

	w.Print(background)
	w.Screen = background
	if !w.Paused {
		w.Next()
	}
	w.camera.Render(background, screen)

	worldX, worldY := w.camera.ScreenToWorld(ebiten.CursorPosition())
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("TPS: %0.2f\n"+
			"Pause/Play (ESC)(Enter)\n"+
			"Move (WASD/Arrows)\n"+
			"Zoom (QE)\n"+
			"Rotate (R)\n"+
			"Reset camera (Space)\n"+
			"Add/Remove Recource (LShift+LCM)/(LCntrl+LCM)\n"+
			"Move Recource/Pedestal (LCM/RCM)\n"+
			"Next step (N)\n", ebiten.CurrentTPS()),
		2, 1,
	)

	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("%s\nCursor World Pos: %.2f,%.2f\nIterations count: %d",
			w.camera.String(),
			worldX, worldY, w.LifeTime),
		2, ScreenHeight-48,
	)
}

func Agents(count int) []*Agent {
	agents := make([]*Agent, count)
	return agents
}

func (w *World) Init() {
	w.Agents = Agents(agentsCount)
	w.InitAgents()
}

func (w *World) InitAgents() {
	for i := 0; i < agentsCount; i++ {
		xRange := FloatRange{0, float64(w.Width)}
		yRange := FloatRange{0, float64(w.Height)}
		fRange := FloatRange{0, Tau}
		w.Agents[i] = NewAgent(xRange.NextRandom(r), yRange.NextRandom(r), fRange.NextRandom(r))
	}
}

func (w *World) Print(screen *ebiten.Image) {
	for _, circle := range resourcesCircles {
		w.drawCircle(screen, circle.x, circle.y, circle.r, circle.c, true)
	}
	for _, circle := range pedestalCircles {
		w.drawCircle(screen, circle.x, circle.y, circle.r, circle.c, true)
	}

	for i := 0; i < w.Height; i++ {
		for j := 0; j < w.Width; j++ {
			if i == 0 {
				ebitenutil.DrawRect(screen, float64(j), float64(i), 1, 1, BorderColor)
			} else if i == w.Height-1 {
				ebitenutil.DrawRect(screen, float64(j), float64(i), 1, 1, BorderColor)
			} else if j == 0 {
				ebitenutil.DrawRect(screen, float64(j), float64(i), 1, 1, BorderColor)
			} else if j == w.Width-1 {
				ebitenutil.DrawRect(screen, float64(j), float64(i), 1, 1, BorderColor)
			}
		}
	}

	for _, agent := range w.Agents {
		agent.Render(screen)
	}
}

func (w *World) Next() {
	for _, agent := range w.Agents {
		agent.NextStep(w)
	}
	w.LifeTime++
}

func (w *World) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (w *World) drawCircle(screen *ebiten.Image, x, y, radius float64, clr color.Color, fill bool) {
	radius64 := radius
	minAngle := math.Acos(1 - 1/radius64)

	for angle := float64(0); angle <= 360; angle += minAngle {
		xDelta := radius64 * math.Cos(angle)
		yDelta := radius64 * math.Sin(angle)

		x1 := int(math.Round(x + xDelta))
		y1 := int(math.Round(y + yDelta))

		if fill {
			if y1 < int(y) {
				for y2 := y1; y2 <= int(y); y2++ {
					screen.Set(x1, y2, clr)
				}
			} else {
				for y2 := y1; y2 > int(y); y2-- {
					screen.Set(x1, y2, clr)
				}
			}
		}

		screen.Set(x1, y1, clr)
	}
}

func RemoveCircle(circles []*Circle, circle *Circle) []*Circle {
	for i, c := range circles {
		if c == circle {
			return append(circles[:i], circles[i+1:]...)
		}
	}
	return circles
}
