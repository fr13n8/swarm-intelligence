package main

import (
	"fmt"
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"golang.org/x/image/math/f64"

	"github.com/fr13n8/swarm-intelligence/camera"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var numWorkers = runtime.NumCPU()

type World struct {
	Agents           []*Agent
	Width            int
	Height           int
	camera           camera.Camera
	LifeTime         int
	Paused           bool
	Screen           *ebiten.Image
	signalWorkerPool sync.WaitGroup
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
	return int(worldX), int(worldY)
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
					resourcesCircles = append(resourcesCircles, NewCircle(float64(x), float64(y), 20,
						color.RGBA{R: 223, G: 250, B: 90, A: 255}))
				}
				return nil
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		w.Paused = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		w.Paused = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		w.camera.Position[0] -= 10
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		w.camera.Position[0] += 10
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		w.camera.Position[1] -= 10
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		w.camera.Position[1] += 10
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

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		w.camera.Reset()
	}

	return nil
}

func (w *World) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	background := ebiten.NewImage(ScreenWidth, ScreenHeight)
	background.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

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
			"Next step (N)\n", ebiten.ActualTPS()),
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

func (w *World) Init() {
	w.Agents = make([]*Agent, agentsCount)
	w.InitAgents()
}

func (w *World) InitAgents() {
	var wg sync.WaitGroup
	batchSize := agentsCount / numWorkers

	for b := 0; b < numWorkers; b++ {
		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()

			for i := startIdx; i < endIdx; i++ {
				if i >= agentsCount {
					break
				}

				x := GetNextRandom(0, float64(w.Width))
				y := GetNextRandom(0, float64(w.Height))
				angle := GetNextRandom(0, Tau)

				w.Agents[i] = NewAgent(x, y, angle)
			}
		}(b*batchSize, (b+1)*batchSize)
	}

	if remCount := agentsCount % numWorkers; remCount > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startIdx := numWorkers * batchSize

			for i := startIdx; i < agentsCount; i++ {
				x := GetNextRandom(0, float64(w.Width))
				y := GetNextRandom(0, float64(w.Height))
				angle := GetNextRandom(0, Tau)

				w.Agents[i] = NewAgent(x, y, angle)
			}
		}()
	}

	wg.Wait()
}

func (w *World) Print(screen *ebiten.Image) {
	for _, circle := range resourcesCircles {
		w.drawCircle(screen, circle.x, circle.y, circle.r, circle.c, true)
	}
	for _, circle := range pedestalCircles {
		w.drawCircle(screen, circle.x, circle.y, circle.r, circle.c, true)
	}

	borderWidth := float32(1.0)

	// Top border
	vector.StrokeLine(screen, 0, 1, float32(w.Width), 0, borderWidth, BorderColor, false)
	// Bottom border
	vector.StrokeLine(screen, 0, float32(w.Height-1), float32(w.Width), float32(w.Height-1), borderWidth, BorderColor, false)
	// Left border
	vector.StrokeLine(screen, 1, 0, 0, float32(w.Height), borderWidth, BorderColor, false)
	// Right border
	vector.StrokeLine(screen, float32(w.Width-1), 0, float32(w.Width-1), float32(w.Height), borderWidth, BorderColor, false)

	var wg sync.WaitGroup
	batchSize := len(w.Agents) / numWorkers

	for b := 0; b < numWorkers; b++ {
		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()

			for i := startIdx; i < endIdx; i++ {
				if i >= len(w.Agents) {
					break
				}
				w.Agents[i].Render(screen)
			}
		}(b*batchSize, (b+1)*batchSize)
	}

	if remCount := len(w.Agents) % numWorkers; remCount > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startIdx := numWorkers * batchSize

			for i := startIdx; i < len(w.Agents); i++ {
				w.Agents[i].Render(screen)
			}
		}()
	}

	wg.Wait()
}

func (w *World) Next() {
	w.signalWorkerPool.Wait()

	var wg sync.WaitGroup
	batchSize := len(w.Agents) / numWorkers

	for b := 0; b < numWorkers; b++ {
		wg.Add(1)
		go func(startIdx, endIdx int) {
			defer wg.Done()

			for i := startIdx; i < endIdx; i++ {
				if i >= len(w.Agents) {
					break
				}
				w.Agents[i].NextStep(w)
			}
		}(b*batchSize, (b+1)*batchSize)
	}

	if remCount := len(w.Agents) % numWorkers; remCount > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startIdx := numWorkers * batchSize

			for i := startIdx; i < len(w.Agents); i++ {
				w.Agents[i].NextStep(w)
			}
		}()
	}

	wg.Wait()
	w.LifeTime++
}

func (w *World) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (w *World) drawCircle(screen *ebiten.Image, x, y, radius float64, clr color.Color, fill bool) {
	if fill {
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(radius), clr, true)
		return
	}

	radius64 := radius
	minAngle := math.Acos(1 - 0.5/radius64)

	for angle := float64(0); angle <= Tau; angle += minAngle {
		sin, cos := math.Sincos(angle)
		xDelta := radius64 * cos
		yDelta := radius64 * sin

		x1 := int(math.Round(x + xDelta))
		y1 := int(math.Round(y + yDelta))

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
