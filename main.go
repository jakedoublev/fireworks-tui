package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

type particle struct {
	x, y     float64
	vx, vy   float64
	char     rune
	color    tcell.Color
	lifetime int
}

type firework struct {
	particles []particle
	frames    int
}

type rocket struct {
	x, y       int
	targetY    int
	color      tcell.Color
	trailTimer int
}

func newFirework(x, y float64) firework {
	chars := []rune{'*', '+', 'o', 'x', '.'}
	colors := []tcell.Color{
		tcell.ColorRed,
		tcell.ColorGreen,
		tcell.ColorYellow,
		tcell.ColorBlue,
		tcell.ColorDarkMagenta,
		tcell.ColorDarkCyan,
		tcell.ColorWhite,
	}

	n := rand.Intn(30) + 20
	p := make([]particle, n)
	color := colors[rand.Intn(len(colors))]

	for i := 0; i < n; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float64()*2 + 1
		p[i] = particle{
			x:        x,
			y:        y,
			vx:       speed * math.Cos(angle),
			vy:       speed * math.Sin(angle) * 0.5,
			char:     chars[rand.Intn(len(chars))],
			color:    color,
			lifetime: rand.Intn(20) + 10,
		}
	}

	return firework{
		particles: p,
	}
}

func (fw *firework) update() {
	alive := []particle{}
	for i := range fw.particles {
		p := &fw.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.1 // gravity
		p.lifetime--
		if p.lifetime > 0 {
			alive = append(alive, *p)
		}
	}
	fw.particles = alive
	fw.frames++
}

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	defer screen.Fini()
	screen.Clear()
	screen.EnableMouse()
	screen.HideCursor()

	width, height := screen.Size()

	fireworks := []firework{}
	rockets := []rocket{}

	// Timers
	frameTicker := time.NewTicker(80 * time.Millisecond)
	launchTicker := time.NewTicker(1200 * time.Millisecond)
	defer frameTicker.Stop()
	defer launchTicker.Stop()

	quit := make(chan struct{})

	go func() {
		for {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC {
					close(quit)
					return
				}
			case *tcell.EventMouse:
				x, y := ev.Position()
				if ev.Buttons()&tcell.Button1 != 0 {
					rockets = append(rockets, newRocket(x, y, height))
				}
			case *tcell.EventResize:
				width, height = screen.Size()
			}
		}
	}()

loop:
	for {
		select {
		case <-quit:
			break loop

		case <-launchTicker.C:
			x := rand.Intn(width-4) + 2
			y := rand.Intn(height/3) + height/4
			rockets = append(rockets, newRocket(x, y, height))

		case <-frameTicker.C:
			screen.Clear()

			// === Animate rockets ===
			activeRockets := []rocket{}
			for i := range rockets {
				r := &rockets[i]
				if r.y > r.targetY {
					style := tcell.StyleDefault.Foreground(r.color)
					screen.SetContent(r.x, r.y, '^', nil, style)
					r.y--
					if r.trailTimer%2 == 0 {
						screen.SetContent(r.x, r.y+1, '|', nil, style)
					}
					r.trailTimer++
					activeRockets = append(activeRockets, *r)
				} else {
					fireworks = append(fireworks, newFirework(float64(r.x), float64(r.y)))
				}
			}
			rockets = activeRockets

			// === Animate fireworks ===
			activeFireworks := []firework{}
			for i := range fireworks {
				fireworks[i].update()
				for _, p := range fireworks[i].particles {
					sx := int(math.Round(p.x))
					sy := int(math.Round(p.y))
					if sx >= 0 && sy >= 0 && sx < width && sy < height {
						style := tcell.StyleDefault.Foreground(p.color)
						screen.SetContent(sx, sy, p.char, nil, style)
					}
				}
				if fireworks[i].frames < 60 && len(fireworks[i].particles) > 0 {
					activeFireworks = append(activeFireworks, fireworks[i])
				}
			}
			fireworks = activeFireworks

			screen.Show()
		}
	}
}

// === Rocket ===

func newRocket(x, targetY, height int) rocket {
	colors := []tcell.Color{
		tcell.ColorRed,
		tcell.ColorGreen,
		tcell.ColorYellow,
		tcell.ColorBlue,
		tcell.ColorDarkMagenta,
		tcell.ColorDarkCyan,
		tcell.ColorWhite,
	}
	return rocket{
		x:       x,
		y:       height - 2,
		targetY: targetY,
		color:   colors[rand.Intn(len(colors))],
	}
}
