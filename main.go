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
	rand.Seed(time.Now().UnixNano())
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

	fireworks := []firework{}
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

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
					fireworks = append(fireworks, newFirework(float64(x), float64(y)))
				}
			}
		}
	}()

loop:
	for {
		select {
		case <-quit:
			break loop
		case <-ticker.C:
			screen.Clear()
			active := []firework{}
			for i := range fireworks {
				fireworks[i].update()
				for _, p := range fireworks[i].particles {
					sx := int(math.Round(p.x))
					sy := int(math.Round(p.y))
					if sx >= 0 && sy >= 0 {
						st := tcell.StyleDefault.Foreground(p.color)
						screen.SetContent(sx, sy, p.char, nil, st)
					}
				}
				if fireworks[i].frames < 60 && len(fireworks[i].particles) > 0 {
					active = append(active, fireworks[i])
				}
			}
			fireworks = active
			screen.Show()
		}
	}
}
