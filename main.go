package main

import (
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	// Fill the white image with white color
	whiteImage.Fill(color.White)
}

type Ball struct {
	g            *Game
	x            float32
	y            float32
	size         float32 // the ball is always based on a Square so the Height and width are always the same
	velX         float32 // Added for horizontal movement
	velY         float32 // Added for vertical movement
	rect         image.Rectangle
	justCollided bool
}

func (b *Ball) Update() {

	b.x += b.velX
	b.y += b.velY

	b.rect = image.Rect(
		int(b.x),
		int(b.y),
		int(b.x+b.size),
		int(b.y+b.size),
	)

	if b.rect.Min.X <= 0 || b.rect.Max.X >= screenWidth {
		b.velX = -b.velX
	}
	if b.rect.Min.Y <= 0 || b.rect.Max.Y >= screenHeight {
		b.velY = -b.velY
	}

	collided := false
	if b.rect.Overlaps(b.g.player1.rect) {
		p := b.g.player1
		resolveCollisions(b, p)
		collided = true
	}

	if b.rect.Overlaps(b.g.player2.rect) {
		p := b.g.player2
		resolveCollisions(b, p)
		collided = true
	}

	if collided {
		b.velX = maxClamp(b.velX+b.velX*0.05, maxVelocity)
		b.velY = maxClamp(b.velY+b.velY*0.05, maxVelocity)
	}

}

func resolveCollisions(b *Ball, p *Paddle) {
	isXOverlap := b.x+b.size > float32(p.rect.Min.X) && b.x < float32(b.rect.Max.X)
	isYOverlap := b.y+b.size > float32(p.rect.Min.Y) && b.y < float32(b.rect.Max.Y)

	if isXOverlap && isYOverlap {
		xDepth := minAbs(float32(p.rect.Min.X)-(b.x+b.size), float32(p.rect.Max.X)-b.x)
		yDepth := minAbs(float32(p.rect.Min.Y)-(b.y+b.size), float32(p.rect.Max.Y)-b.y)

		if abs(xDepth) < abs(yDepth) {
			// Resolve X collision if it's shallower
			b.x, b.velX = handleCollision(b.x, b.size, b.velX, float32(p.rect.Min.X), float32(p.rect.Max.X))
		} else {
			// Resolve Y collision if it's shallower
			b.y, b.velY = handleCollision(b.y, b.size, b.velY, float32(p.rect.Min.Y), float32(p.rect.Max.Y))
		}

	} else if isXOverlap {
		b.x, b.velX = handleCollision(b.x, b.size, b.velX, float32(p.rect.Min.X), float32(p.rect.Max.X))

	} else if isYOverlap {
		b.y, b.velY = handleCollision(b.y, b.size, b.velY, float32(p.rect.Min.Y), float32(p.rect.Max.Y))

	}

}

func minAbs(a, b float32) float32 {
	if abs(a) < abs(b) {
		return a
	}
	return b
}

func abs(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}

func handleCollision(ballPos, ballSize, ballVel, paddleMin, paddleMax float32) (float32, float32) {
	if ballPos+ballSize > paddleMin && ballPos < paddleMin {
		// Top or Left collision
		ballPos = paddleMin - ballSize - 1 // Makes sure that the ball and the paddle ain't touching again
		ballVel = -ballVel
	} else if ballPos < paddleMax && ballPos+ballSize > paddleMax {
		// Bottom or Right collision
		ballPos = paddleMax + 1 // Makes sure that the ball and the paddle ain't touching again
		ballVel = -ballVel
	}
	return ballPos, ballVel
}

func (b *Ball) Draw(screen *ebiten.Image) {
	path := vector.Path{}
	centerX := b.x + b.size/2
	centerY := b.y + b.size/2

	path.MoveTo(centerX, centerY)
	path.Arc(centerX, centerY, b.size/2, 0, 2*math.Pi, vector.Clockwise)
	path.Close()

	b.g.vertices, b.g.indices = path.AppendVerticesAndIndicesForFilling(b.g.vertices[:0], b.g.indices[:0])
	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	op.FillRule = ebiten.FillRuleNonZero
	screen.DrawTriangles(b.g.vertices, b.g.indices, whiteSubImage, op)

}

type Paddle struct {
	g        *Game
	x        float32
	y        float32
	width    float32
	height   float32
	velocity float32
	rect     image.Rectangle
	keyLeft  ebiten.Key
	keyRight ebiten.Key
}

func (p *Paddle) Update() {

	if ebiten.IsKeyPressed(p.keyLeft) {
		p.x -= p.velocity
		if p.x < 0 {
			p.x = 0
		}
	}

	if ebiten.IsKeyPressed(p.keyRight) {
		p.x += p.velocity
		if p.x > screenWidth-p.width {
			p.x = screenWidth - p.width
		}
	}

	p.rect = image.Rect(int(p.x), int(p.y), int(p.x+p.width), int(p.y+p.height))
}

func (p *Paddle) Draw(screen *ebiten.Image) {
	// Create a path using the vector library
	path := vector.Path{}
	path.MoveTo(p.x, p.y)                  // Top-left corner
	path.LineTo(p.x+p.width, p.y)          // Top-right corner
	path.LineTo(p.x+p.width, p.y+p.height) // Bottom-right corner
	path.LineTo(p.x, p.y+p.height)         // Bottom-left corner
	path.Close()

	p.g.vertices, p.g.indices = path.AppendVerticesAndIndicesForFilling(p.g.vertices[:0], p.g.indices[:0])

	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	op.FillRule = ebiten.FillRuleNonZero
	screen.DrawTriangles(p.g.vertices, p.g.indices, whiteSubImage, op)

}

type Game struct {
	player1  *Paddle
	player2  *Paddle
	ball     *Ball
	vertices []ebiten.Vertex
	indices  []uint16
}

const (
	screenWidth  = 640
	screenHeight = 480
	maxVelocity  = 5.5
)

func maxClamp(value, max float32) float32 {
	if value >= max {
		return max
	} else if value <= -max {
		return -max
	}

	return value

}

func (g *Game) Update() error {
	g.player1.Update()
	g.player2.Update()
	g.ball.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player1.Draw(screen)
	g.player2.Draw(screen)
	g.ball.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{}
	game.player1 = &Paddle{
		g:        game,                    // Pass the game reference
		x:        (screenWidth - 100) / 2, // Start centered horizontally
		y:        screenHeight - 40,       // Start near the bottom
		width:    100,
		height:   8,
		velocity: 10,
		keyLeft:  ebiten.KeyS,
		keyRight: ebiten.KeyD,
	}
	game.player2 = &Paddle{
		g:        game,
		x:        (screenWidth - 100) / 2,
		y:        40,
		width:    100,
		height:   8,
		velocity: 10,
		keyLeft:  ebiten.KeyArrowLeft,
		keyRight: ebiten.KeyArrowRight,
	}
	game.ball = &Ball{
		g:    game,
		size: 20,
		x:    screenHeight / 2,
		y:    screenWidth / 2,
		velY: 2.5,
		velX: 2.5,
	}
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Hello, Pong!")
	if err := ebiten.RunGame(game); err != nil { // Pass the initialized game instance
		log.Fatal(err)
	}
}
