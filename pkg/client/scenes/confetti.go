package scenes

import (
	"math/rand"

	"github.com/nsf/termbox-go"
)

var confettiColors = []termbox.Attribute{
	termbox.ColorBlack,
	termbox.ColorRed,
	termbox.ColorGreen,
	termbox.ColorYellow,
	termbox.ColorBlue,
	termbox.ColorMagenta,
	termbox.ColorCyan,
	termbox.ColorWhite,
}

var confettiShapes = []rune{'▪', '▮', '▰', '▴', '▸', '▾', '◂', '◆'}

type paper struct {
	x, y  int
	color termbox.Attribute
}

type confetti []*paper

func (c *confetti) tick() {
	if c == nil {
		return
	}

	// Fall.
	for _, p := range *c {
		p.y++
		p.x = p.x - 1 + rand.Intn(3) //nolint:gosec
	}

	width, height := termbox.Size()

	// Destroy off-screen paper.
	oldPaper := *c
	*c = make([]*paper, 0)
	for _, p := range oldPaper {
		if p.y < height {
			*c = append(*c, p)
		}
	}

	// Spawn new paper.
	for i := 0; i < width/30; i++ {
		x := rand.Intn(width)                                   //nolint:gosec
		color := confettiColors[rand.Intn(len(confettiColors))] //nolint:gosec
		*c = append(*c, &paper{x: x, color: color})
	}
}

func (c *confetti) draw() {
	if c == nil {
		return
	}

	if len(*c) == 0 {
		return
	}

	for _, p := range *c {
		shape := confettiShapes[rand.Intn(len(confettiShapes))] //nolint:gosec
		termbox.SetCell(p.x, p.y, shape, p.color, termbox.ColorDefault)
	}

	// Clear the last color.
	width, height := termbox.Size()
	termbox.SetCell(width-1, height-1, ' ', termbox.ColorDefault, termbox.ColorDefault)
}
