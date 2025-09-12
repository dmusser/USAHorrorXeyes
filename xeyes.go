package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	windowWidth  = 480  // 20% smaller (was 600)
	windowHeight = 320  // 20% smaller (was 400)
	eyeRadius    = 64   // 20% smaller (was 80)
	pupilRadius  = 24   // 20% smaller (was 30)
	creepyPupilRadius = 30 // 20% smaller (was 38)
	eyeSpacing   = 96   // 20% smaller (was 120)
)

type Game struct {
	dragging       bool
	dragStartX     int
	dragStartY     int
	windowStartX   int
	windowStartY   int
	frozen         bool
	lastMouseX     int
	lastMouseY     int
	idleTime       int
	randomTargetX  float32
	randomTargetY  float32
	currentX       float32
	currentY       float32
	
	// Creepy mode
	creepyMode     bool
	bloodDrops     []BloodDrop
	lastBloodTime  int
	pupilShake     float32
	shakeTimer     int
	bloodTrails    []BloodTrail
	
	// Double click detection
	lastClickTime  int64
	clickCount     int
}

type BloodDrop struct {
	x, y   float32
	speed  float32
	length float32
	alpha  uint8
	width  float32
}

type BloodTrail struct {
	x, y      float32
	length    float32
	thickness float32
	alpha     uint8
}

func (g *Game) Update() error {
	// Exit on ESC key
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// Toggle freeze with F key
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.frozen = !g.frozen
	}

	// Handle right-click for freeze (separate from dragging)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.frozen = !g.frozen
		return nil // Don't process other mouse events this frame
	}

	// Handle double-click to switch modes
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		currentTime := time.Now().UnixMilli()
		
		// Check if this is part of a double-click (within 500ms)
		if currentTime-g.lastClickTime < 500 {
			g.clickCount++
			if g.clickCount >= 2 {
				g.creepyMode = !g.creepyMode
				g.clickCount = 0
				// Clear blood effects when switching to normal mode
				if !g.creepyMode {
					g.bloodDrops = g.bloodDrops[:0]
					g.bloodTrails = g.bloodTrails[:0]
				}
				return nil // Don't start dragging on mode switch
			}
		} else {
			g.clickCount = 1
		}
		g.lastClickTime = currentTime
		
		// Start dragging (but only if not a double-click)
		time.AfterFunc(200*time.Millisecond, func() {
			if g.clickCount == 1 && !g.dragging {
				g.dragging = true
				g.dragStartX, g.dragStartY = ebiten.CursorPosition()
				g.windowStartX, g.windowStartY = ebiten.WindowPosition()
			}
		})
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.dragging = false
	}

	if g.dragging {
		currentX, currentY := ebiten.CursorPosition()
		deltaX := currentX - g.dragStartX
		deltaY := currentY - g.dragStartY
		newX := g.windowStartX + deltaX
		newY := g.windowStartY + deltaY
		ebiten.SetWindowPosition(newX, newY)
	}

	// Normal mode: Check for mouse movement and random eye movement
	if !g.creepyMode {
		currentMouseX, currentMouseY := ebiten.CursorPosition()
		if currentMouseX != g.lastMouseX || currentMouseY != g.lastMouseY {
			g.idleTime = 0
			g.lastMouseX = currentMouseX
			g.lastMouseY = currentMouseY
		} else {
			g.idleTime++
			// After 120 frames (~2 seconds) of no mouse movement, generate new random target
			if g.idleTime == 120 || (g.idleTime > 120 && (g.idleTime-120)%(180+rand.Intn(240)) == 0) {
				g.randomTargetX = rand.Float32() * windowWidth
				g.randomTargetY = rand.Float32() * windowHeight
			}
		}
	}

	// Creepy mode: Horror effects updates
	if g.creepyMode {
		g.updateHorrorEffects()
	}

	return nil
}

func (g *Game) updateHorrorEffects() {
	// Subtle pupil shaking/twitching
	g.shakeTimer++
	if g.shakeTimer > 180 && rand.Float32() < 0.03 {
		g.pupilShake = 1.5 + rand.Float32()*2.0
		g.shakeTimer = 0
	}
	if g.pupilShake > 0 {
		g.pupilShake *= 0.9
	}

	// More frequent, larger blood drops - starting from BOTTOM of eyes
	g.lastBloodTime++
	if g.lastBloodTime > 80 && rand.Float32() < 0.12 {
		leftEyeX := float32(windowWidth/2 - eyeSpacing/2)
		rightEyeX := float32(windowWidth/2 + eyeSpacing/2)
		eyeY := float32(windowHeight/2)
		
		var eyeX float32
		if rand.Float32() < 0.5 {
			eyeX = leftEyeX
		} else {
			eyeX = rightEyeX
		}
		
		// Blood starts from BOTTOM of eye (eyeY + eyeRadius)
		g.bloodDrops = append(g.bloodDrops, BloodDrop{
			x:      eyeX + (rand.Float32()-0.5)*48, // Scaled down 20%
			y:      eyeY + eyeRadius,
			speed:  0.3 + rand.Float32()*1.2,
			length: 20 + rand.Float32()*32, // Scaled down 20%
			alpha:  180 + uint8(rand.Float32()*70),
			width:  3.2 + rand.Float32()*3.2, // Scaled down 20%
		})
		g.lastBloodTime = 0
	}

	// Add blood trails from bottom of eyes
	if rand.Float32() < 0.015 {
		leftEyeX := float32(windowWidth/2 - eyeSpacing/2)
		rightEyeX := float32(windowWidth/2 + eyeSpacing/2)
		eyeY := float32(windowHeight/2)
		
		var startX, startY float32
		if rand.Float32() < 0.5 {
			startX = leftEyeX + (rand.Float32()-0.5)*eyeRadius*0.8
			startY = eyeY + eyeRadius - 8    // Scaled down 20%
		} else {
			startX = rightEyeX + (rand.Float32()-0.5)*eyeRadius*0.8
			startY = eyeY + eyeRadius - 8    // Scaled down 20%
		}
		
		g.bloodTrails = append(g.bloodTrails, BloodTrail{
			x:         startX,
			y:         startY,
			length:    16 + rand.Float32()*28, // Scaled down 20%
			thickness: 1.6 + rand.Float32()*2.4, // Scaled down 20%
			alpha:     130 + uint8(rand.Float32()*90),
		})
	}

	// Update blood drops
	for i := len(g.bloodDrops) - 1; i >= 0; i-- {
		g.bloodDrops[i].y += g.bloodDrops[i].speed
		g.bloodDrops[i].alpha = uint8(float32(g.bloodDrops[i].alpha) * 0.996)
		if g.bloodDrops[i].y > windowHeight+50 || g.bloodDrops[i].alpha < 10 {
			g.bloodDrops = append(g.bloodDrops[:i], g.bloodDrops[i+1:]...)
		}
	}

	// Update blood trails
	for i := len(g.bloodTrails) - 1; i >= 0; i-- {
		g.bloodTrails[i].alpha = uint8(float32(g.bloodTrails[i].alpha) * 0.985)
		if g.bloodTrails[i].alpha < 20 {
			g.bloodTrails = append(g.bloodTrails[:i], g.bloodTrails[i+1:]...)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear with transparent background
	screen.Fill(color.RGBA{0, 0, 0, 0})

	if g.creepyMode {
		g.drawCreepyEyes(screen)
	} else {
		g.drawNormalEyes(screen)
	}
}

func (g *Game) drawNormalEyes(screen *ebiten.Image) {
	// Get target position - mouse or smoothly interpolated random
	var targetX, targetY float32
	if !g.dragging && !g.frozen {
		if g.idleTime < 120 {
			// Use mouse position
			targetX = float32(g.lastMouseX)
			targetY = float32(g.lastMouseY)
			g.currentX = targetX
			g.currentY = targetY
		} else {
			// Smoothly interpolate toward random target (slower movement)
			lerpSpeed := float32(0.015) // Slower interpolation speed
			g.currentX += (g.randomTargetX - g.currentX) * lerpSpeed
			g.currentY += (g.randomTargetY - g.currentY) * lerpSpeed
			targetX = g.currentX
			targetY = g.currentY
		}
	} else {
		// Keep pupils centered when dragging or frozen
		targetX = windowWidth / 2
		targetY = windowHeight / 2
		g.currentX = targetX
		g.currentY = targetY
	}

	// Eye centers
	leftEyeCenterX := float32(windowWidth/2 - eyeSpacing/2)
	leftEyeCenterY := float32(windowHeight / 2)
	rightEyeCenterX := float32(windowWidth/2 + eyeSpacing/2)
	rightEyeCenterY := float32(windowHeight / 2)

	// Calculate pupil positions
	leftPupilX, leftPupilY := updatePupil(leftEyeCenterX, leftEyeCenterY, targetX, targetY, pupilRadius)
	rightPupilX, rightPupilY := updatePupil(rightEyeCenterX, rightEyeCenterY, targetX, targetY, pupilRadius)

	// Draw eye shadows
	vector.DrawFilledCircle(screen, leftEyeCenterX+2.4, leftEyeCenterY+2.4, eyeRadius, color.RGBA{100, 100, 100, 120}, false)
	vector.DrawFilledCircle(screen, rightEyeCenterX+2.4, rightEyeCenterY+2.4, eyeRadius, color.RGBA{100, 100, 100, 120}, false)

	// Draw eye balls (white)
	vector.DrawFilledCircle(screen, leftEyeCenterX, leftEyeCenterY, eyeRadius, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeCenterX, rightEyeCenterY, eyeRadius, color.RGBA{255, 255, 255, 255}, false)

	// Draw eye borders
	vector.StrokeCircle(screen, leftEyeCenterX, leftEyeCenterY, eyeRadius, 1.6, color.RGBA{200, 200, 200, 255}, false)
	vector.StrokeCircle(screen, rightEyeCenterX, rightEyeCenterY, eyeRadius, 1.6, color.RGBA{200, 200, 200, 255}, false)

	// Draw pupils (black)
	vector.DrawFilledCircle(screen, leftPupilX, leftPupilY, pupilRadius, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightPupilX, rightPupilY, pupilRadius, color.RGBA{0, 0, 0, 255}, false)

	// Draw highlights on pupils
	highlightSize := float32(pupilRadius / 4)
	vector.DrawFilledCircle(screen, leftPupilX-pupilRadius/3, leftPupilY-pupilRadius/3, highlightSize, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, rightPupilX-pupilRadius/3, rightPupilY-pupilRadius/3, highlightSize, color.RGBA{255, 255, 255, 255}, false)
}

func (g *Game) drawCreepyEyes(screen *ebiten.Image) {
	// Get mouse position
	var cursorX, cursorY int
	if !g.dragging && !g.frozen {
		cursorX, cursorY = ebiten.CursorPosition()
	} else {
		cursorX = windowWidth / 2
		cursorY = windowHeight / 2
	}

	// Eye centers
	leftEyeCenterX := float32(windowWidth/2 - eyeSpacing/2)
	leftEyeCenterY := float32(windowHeight / 2)
	rightEyeCenterX := float32(windowWidth/2 + eyeSpacing/2)
	rightEyeCenterY := float32(windowHeight / 2)

	// Calculate pupil positions with subtle shake
	leftPupilX, leftPupilY := updatePupil(leftEyeCenterX, leftEyeCenterY, float32(cursorX), float32(cursorY), creepyPupilRadius)
	rightPupilX, rightPupilY := updatePupil(rightEyeCenterX, rightEyeCenterY, float32(cursorX), float32(cursorY), creepyPupilRadius)

	// Apply subtle shake to pupils
	shakeX := (rand.Float32() - 0.5) * g.pupilShake
	shakeY := (rand.Float32() - 0.5) * g.pupilShake
	leftPupilX += shakeX
	leftPupilY += shakeY
	rightPupilX += shakeX
	rightPupilY += shakeY

	// Draw subtle shadows
	vector.DrawFilledCircle(screen, leftEyeCenterX+2.4, leftEyeCenterY+2.4, eyeRadius, color.RGBA{40, 10, 10, 120}, false)
	vector.DrawFilledCircle(screen, rightEyeCenterX+2.4, rightEyeCenterY+2.4, eyeRadius, color.RGBA{40, 10, 10, 120}, false)

	// Draw bloodshot eye whites (static redness)
	vector.DrawFilledCircle(screen, leftEyeCenterX, leftEyeCenterY, eyeRadius, color.RGBA{240, 180, 180, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeCenterX, rightEyeCenterY, eyeRadius, color.RGBA{240, 180, 180, 255}, false)

	// Draw COMPLETELY STATIC bloodshot veins (same pattern every time)
	g.drawCompletelyStaticVeins(screen, leftEyeCenterX, leftEyeCenterY)
	g.drawCompletelyStaticVeins(screen, rightEyeCenterX, rightEyeCenterY)

	// Draw dark eye borders
	vector.StrokeCircle(screen, leftEyeCenterX, leftEyeCenterY, eyeRadius, 2.4, color.RGBA{100, 20, 20, 255}, false)
	vector.StrokeCircle(screen, rightEyeCenterX, rightEyeCenterY, eyeRadius, 2.4, color.RGBA{100, 20, 20, 255}, false)

	// Draw large, dark pupils
	vector.DrawFilledCircle(screen, leftPupilX, leftPupilY, creepyPupilRadius, color.RGBA{20, 5, 5, 255}, false)
	vector.DrawFilledCircle(screen, rightPupilX, rightPupilY, creepyPupilRadius, color.RGBA{20, 5, 5, 255}, false)
	vector.DrawFilledCircle(screen, leftPupilX, leftPupilY, creepyPupilRadius-2.4, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightPupilX, rightPupilY, creepyPupilRadius-2.4, color.RGBA{0, 0, 0, 255}, false)

	// Static red highlights
	highlightSize := float32(creepyPupilRadius / 6)
	vector.DrawFilledCircle(screen, leftPupilX-creepyPupilRadius/3, leftPupilY-creepyPupilRadius/3, highlightSize, color.RGBA{200, 50, 50, 180}, false)
	vector.DrawFilledCircle(screen, rightPupilX-creepyPupilRadius/3, rightPupilY-creepyPupilRadius/3, highlightSize, color.RGBA{200, 50, 50, 180}, false)

	// Draw blood trails
	for _, trail := range g.bloodTrails {
		vector.StrokeLine(screen, trail.x, trail.y, trail.x, trail.y+trail.length, trail.thickness, color.RGBA{130, 10, 10, trail.alpha}, false)
	}

	// Draw blood drops
	for _, drop := range g.bloodDrops {
		vector.StrokeLine(screen, drop.x, drop.y, drop.x, drop.y+drop.length, drop.width, color.RGBA{120, 15, 15, drop.alpha}, false)
		vector.DrawFilledCircle(screen, drop.x, drop.y+drop.length, drop.width/2+0.8, color.RGBA{100, 10, 10, drop.alpha}, false)
	}
}

func (g *Game) drawCompletelyStaticVeins(screen *ebiten.Image, centerX, centerY float32) {
	// Completely static vein pattern - no randomness, same every frame (scaled down 20%)
	
	// Main radial veins
	angles := []float64{0, 0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0, 5.5}
	
	for _, angle := range angles {
		// Inner vein (scaled down 20%)
		startX := centerX + float32(math.Cos(angle))*20
		startY := centerY + float32(math.Sin(angle))*20
		endX := centerX + float32(math.Cos(angle))*48
		endY := centerY + float32(math.Sin(angle))*48
		vector.StrokeLine(screen, startX, startY, endX, endY, 0.96, color.RGBA{180, 30, 30, 150}, false)
		
		// Outer vein (scaled down 20%)
		startX2 := centerX + float32(math.Cos(angle))*28
		startY2 := centerY + float32(math.Sin(angle))*28
		endX2 := centerX + float32(math.Cos(angle))*56
		endY2 := centerY + float32(math.Sin(angle))*56
		vector.StrokeLine(screen, startX2, startY2, endX2, endY2, 0.64, color.RGBA{160, 40, 40, 120}, false)
	}
	
	// Additional cross veins (static pattern, scaled down 20%)
	crossVeins := [][4]float32{
		{centerX - 32, centerY - 16, centerX + 24, centerY - 12},
		{centerX - 24, centerY + 20, centerX + 32, centerY + 16},
		{centerX - 16, centerY - 32, centerX - 8, centerY + 28},
		{centerX + 12, centerY - 28, centerX + 20, centerY + 24},
		{centerX - 40, centerY, centerX + 36, centerY + 4},
		{centerX - 20, centerY - 24, centerX + 16, centerY + 28},
	}
	
	for _, vein := range crossVeins {
		vector.StrokeLine(screen, vein[0], vein[1], vein[2], vein[3], 0.72, color.RGBA{170, 35, 35, 130}, false)
	}
}

func updatePupil(eyeCenterX, eyeCenterY, cursorX, cursorY float32, pupilSize float32) (float32, float32) {
	dx := cursorX - eyeCenterX
	dy := cursorY - eyeCenterY
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	maxDist := float32(eyeRadius) - pupilSize

	if dist > maxDist {
		pupilX := eyeCenterX + dx/dist*maxDist
		pupilY := eyeCenterY + dy/dist*maxDist
		return pupilX, pupilY
	}
	return cursorX, cursorY
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return windowWidth, windowHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Switchable Eyes")
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowFloating(true)  // Always on top

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
