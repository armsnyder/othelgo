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
	color color
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
	*c = nil
	for _, p := range oldPaper {
		if p.x >= -width/2 && p.x < width/2 && p.y >= -height/2 && p.y < height/2 {
			*c = append(*c, p)
		}
	}

	// Spawn new paper.
	for i := 0; i < width/30; i++ {
		x := rand.Intn(width) - width/2 //nolint:gosec
		y := -height / 2
		color := func() (fg, bg termbox.Attribute) {
			return confettiColors[rand.Intn(len(confettiColors))], termbox.ColorDefault //nolint:gosec
		}

		*c = append(*c, &paper{x: x, y: y, color: color})
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
		draw(offset(center, p.x, p.y), p.color, shape)
	}
}
