package main

import (
	"fmt"       // For printing to the console
	"math"      // For mathematical operations like sine, cosine, and rounding
	"math/rand" // For generating random numbers (e.g., firework positions, colors)
	"os"        // For command-line arguments, environment variables, and exiting
	"os/exec"   // For executing external commands (like 'stty' for terminal control)
	"os/signal" // For handling system signals (like Ctrl+C)

	// For converting strings to integers
	"syscall" // For system calls (used with os/signal for specific signals)
	"time"    // For time-related operations, especially for animation delays

	"github.com/charmbracelet/x/term"
)

// ANSI escape codes for terminal control.
// These codes are used to manipulate the terminal's display,
// such as clearing the screen, moving the cursor, and hiding/showing it.
const (
	clearScreen = "\033[2J"  // Clears the entire screen
	cursorHome  = "\033[H"   // Moves the cursor to the top-left corner (1,1)
	hideCursor  = "\033[?25l" // Hides the terminal cursor
	showCursor  = "\033[?25h" // Shows the terminal cursor
	resetColor  = "\033[0m"  // Resets all ANSI formatting (color, bold, etc.)
)

// ANSI escape codes for various colors.
// These are used to color the firework elements.
var colors = []string{
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	"\033[37m", // White
}

// Particle represents a single element of the firework explosion.
// It tracks its position, velocity, character, color, and remaining lifetime.
type Particle struct {
	x, y     float64 // Current position (using float for smoother sub-character movement)
	vx, vy   float64 // Velocity in x and y directions
	char     string  // The character to display for this particle (e.g., "*", "+")
	color    string  // ANSI color code for the particle
	lifetime int     // Number of frames the particle will remain visible
}

func main() {
	width, height, err := term.GetSize(0)
	if err != nil {
		fmt.Println("error detecting terminal size")
		panic(err)
	}

	// Seed the random number generator.
	// Using the current UnixNano time ensures different animations each run.
	rand.Seed(time.Now().UnixNano())

	// Set up a channel to listen for interrupt signals (like Ctrl+C).
	// This allows for a graceful exit, restoring terminal settings.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // Listen for Ctrl+C and termination signals

	// Start a goroutine to handle signals.
	// When a signal is received, it restores the terminal and exits.
	go func() {
		<-c // Block until a signal is received on channel 'c'
		// Restore terminal settings before exiting.
		enableInputBuffering() // Re-enable normal input buffering
		fmt.Print(showCursor)  // Show the cursor again
		fmt.Print(clearScreen) // Clear the screen
		fmt.Print(cursorHome)  // Move cursor to home
		os.Exit(0)             // Exit the program cleanly
	}()

	// Disable input buffering and hide the cursor.
	// This prevents user input from appearing on the screen and allows for smooth animation.
	disableInputBuffering()
	fmt.Print(hideCursor)

	// Ensure terminal settings are restored and cursor is shown when the program exits,
	// regardless of how it exits (normal completion or panic).
	defer enableInputBuffering() // This will run when main exits
	defer fmt.Print(showCursor)  // This will run when main exits
	defer fmt.Print(clearScreen) // This will run when main exits
	defer fmt.Print(cursorHome)  // This will run when main exits

	// Clear the screen once at the beginning.
	fmt.Print(clearScreen)

	// Main animation loop: continuously launch and explode fireworks.
	for {
		// Determine the launch position and explosion height for the firework.
		// LaunchX is randomized to avoid edges.
		launchX := rand.Intn(width-4) + 2 // X-coordinate for launch (2 units padding from edges)
		launchY := height - 1            // Y-coordinate for launch (bottom of the terminal)
		// ExplosionY is randomized to be in the upper half of the screen.
		explosionY := rand.Intn(height/2) + height/4 // Explode between 1/4 and 3/4 of height

		// Animate the rocket launching upwards.
		rocketChar := "^" // Character representing the rocket
		rocketColor := colors[rand.Intn(len(colors))] // Random color for the rocket
		for y := launchY; y >= explosionY; y-- {
			// Clear the entire screen and move cursor to home for each frame.
			// This causes a slight flicker but simplifies animation logic.
			fmt.Print(cursorHome)
			fmt.Print(clearScreen)

			// Draw the rocket at its current position.
			moveTo(launchX, y)
			fmt.Printf("%s%s%s", rocketColor, rocketChar, resetColor) // Print colored rocket

			time.Sleep(50 * time.Millisecond) // Pause for a short duration to create animation frames
		}

		// Once the rocket reaches its explosion height, trigger the explosion.
		explode(float64(launchX), float64(explosionY), width, height)

		// Small delay before launching the next firework.
		time.Sleep(time.Second)
	}
}

// moveTo moves the terminal cursor to the specified (x, y) coordinates.
// Terminal coordinates are 1-based, so (1,1) is the top-left corner.
// Our internal coordinates are 0-based, so we add 1.
func moveTo(x, y int) {
	fmt.Printf("\033[%d;%dH", y+1, x+1)
}

// explode creates and animates the firework explosion.
// It generates multiple particles that spread outwards and fade away.
func explode(centerX, centerY float64, width, height int) {
	numParticles := rand.Intn(30) + 20 // Generate between 20 and 50 particles
	particles := make([]Particle, numParticles)
	explosionColor := colors[rand.Intn(len(colors))] // All particles in this explosion share a color

	// Characters to use for the explosion particles.
	particleChars := []string{"*", "+", "o", "x", "."}

	// Initialize each particle with a random direction, speed, and lifetime.
	for i := 0; i < numParticles; i++ {
		angle := rand.Float64() * 2 * math.Pi // Random angle for radial spread (0 to 2*PI radians)
		speed := rand.Float64()*2 + 1.0       // Random initial speed (1.0 to 3.0)
		particles[i] = Particle{
			x:        centerX,
			y:        centerY,
			vx:       speed * math.Cos(angle),     // X velocity component
			vy:       speed * math.Sin(angle) * 0.5, // Y velocity component (vertical spread is less pronounced)
			char:     particleChars[rand.Intn(len(particleChars))], // Random character for the particle
			color:    explosionColor,
			lifetime: rand.Intn(20) + 10, // Particle visible for 10-30 frames
		}
	}

	// Animate the explosion over several frames.
	for frame := 0; frame < 60; frame++ { // Max 60 frames for the explosion animation
		fmt.Print(cursorHome)  // Move cursor to home
		fmt.Print(clearScreen) // Clear the screen for redrawing

		aliveParticles := []Particle{} // Slice to hold particles that are still active
		for i := range particles {
			p := &particles[i] // Get a pointer to the current particle

			if p.lifetime > 0 { // Only process particles that are still alive
				// Update particle position based on velocity.
				p.x += p.vx
				p.y += p.vy
				p.vy += 0.1 // Apply a small "gravity" effect, pulling particles downwards

				// Convert float coordinates to integer coordinates for drawing on the terminal.
				drawX := int(math.Round(p.x))
				drawY := int(math.Round(p.y))

				// Check if the particle is within the terminal bounds.
				if drawX >= 0 && drawX < width && drawY >= 0 && drawY < height {
					moveTo(drawX, drawY) // Move cursor to particle's position
					fmt.Printf("%s%s%s", p.color, p.char, resetColor) // Print colored particle
					p.lifetime--                                     // Decrease particle's remaining lifetime
					aliveParticles = append(aliveParticles, *p)      // Add to the list of still active particles
				}
			}
		}
		particles = aliveParticles // Update the main particles slice with only the alive ones

		// If all particles have faded or moved off-screen, and some initial frames have passed, break early.
		if len(particles) == 0 && frame > 10 {
			break
		}
		time.Sleep(80 * time.Millisecond) // Pause for animation speed
	}
}

// disableInputBuffering attempts to put the terminal into "raw" mode.
// This prevents user input from being echoed to the screen and allows direct cursor control.
// WARNING: This function uses the 'stty' command, which is specific to Unix-like systems (Linux, macOS).
// It will not work on Windows. For cross-platform terminal control in Go, consider
// using a library like 'golang.org/x/term'.
func disableInputBuffering() {
	// 'cbreak' makes input available character by character without waiting for newline.
	// 'min 1' ensures read operations return after at least one character.
	cmd := exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		// fmt.Println("Warning: Could not disable input buffering:", err)
	}

	// '-echo' disables echoing of input characters to the terminal.
	cmd = exec.Command("stty", "-F", "/dev/tty", "-echo")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		// fmt.Println("Warning: Could not disable echo:", err)
	}
}

// enableInputBuffering attempts to restore the terminal to its normal "cooked" mode.
// This re-enables input buffering and echoing.
// WARNING: Like disableInputBuffering, this uses 'stty' and is Unix-specific.
func enableInputBuffering() {
	// 'cooked' restores normal line-buffered input.
	// 'echo' re-enables echoing of input characters.
	cmd := exec.Command("stty", "-F", "/dev/tty", "cooked", "echo")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		// fmt.Println("Warning: Could not re-enable input buffering:", err)
	}
}
