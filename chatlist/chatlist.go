// source: https://github.com/gizak/termui/blob/master/widgets/list.go
// modifying standard list to not highlight the selected text
// pretty cursed, but no other way to prevent highlighted chat messages
package chatlist

import (
	"image"

	rw "github.com/mattn/go-runewidth"

	. "github.com/gizak/termui/v3"
)

type List struct {
	Block
	Rows        []string
	WrapText    bool
	TextStyle   Style
	SelectedRow int
	topRow      int
}

func NewList() *List {
	return &List{
		Block:     *NewBlock(),
		TextStyle: Theme.List.Text,
	}
}

func (self *List) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	point := self.Inner.Min

	self.topRow = self.SelectedRow

	// draw rows
	for row := self.topRow; row < len(self.Rows) && point.Y < self.Inner.Max.Y; row++ {
		cells := ParseStyles(self.Rows[row], self.TextStyle)
		if self.WrapText {
			cells = WrapCells(cells, uint(self.Inner.Dx()))
		}
		for j := 0; j < len(cells) && point.Y < self.Inner.Max.Y; j++ {
			style := cells[j].Style
			if cells[j].Rune == '\n' {
				point = image.Pt(self.Inner.Min.X, point.Y+1)
			} else {
				if point.X+1 == self.Inner.Max.X+1 && len(cells) > self.Inner.Dx() {
					buf.SetCell(NewCell(ELLIPSES, style), point.Add(image.Pt(-1, 0)))
					break
				} else {
					buf.SetCell(NewCell(cells[j].Rune, style), point)
					point = point.Add(image.Pt(rw.RuneWidth(cells[j].Rune), 0))
				}
			}
		}
		point = image.Pt(self.Inner.Min.X, point.Y+1)
	}

	// draw UP_ARROW if needed
	if self.topRow > 0 {
		buf.SetCell(
			NewCell(UP_ARROW, NewStyle(ColorWhite)),
			image.Pt(self.Inner.Max.X-1, self.Inner.Min.Y),
		)
	}

	// draw DOWN_ARROW if needed
	if len(self.Rows) > int(self.topRow)+self.realDy() {
		buf.SetCell(
			NewCell(DOWN_ARROW, NewStyle(ColorWhite)),
			image.Pt(self.Inner.Max.X-1, self.Inner.Max.Y-1),
		)
	}
}

// ScrollAmount scrolls by amount given. If amount is < 0, then scroll up.
// There is no need to set self.topRow, as this will be set automatically when drawn,
// since if the selected item is off screen then the topRow variable will change accordingly.
func (self *List) ScrollAmount(amount int) {
	if len(self.Rows)-int(self.SelectedRow) <= amount {
		self.SelectedRow = len(self.Rows) - 1
	} else if int(self.SelectedRow)+amount < 0 {
		self.SelectedRow = 0
	} else {
		self.SelectedRow += amount
	}
}

func (self *List) ScrollUp() {
	self.ScrollAmount(-1)
}

func (self *List) ScrollDown() {
	self.ScrollAmount(1)
}

func (self *List) ScrollPageUp() {
	// If an item is selected below top row, then go to the top row.
	if self.SelectedRow > self.topRow {
		self.SelectedRow = self.topRow
	} else {
		self.ScrollAmount(-self.Inner.Dy())
	}
}

func (self *List) ScrollPageDown() {
	self.ScrollAmount(self.Inner.Dy())
}

func (self *List) ScrollHalfPageUp() {
	self.ScrollAmount(-int(FloorFloat64(float64(self.Inner.Dy()) / 2)))
}

func (self *List) ScrollHalfPageDown() {
	self.ScrollAmount(int(FloorFloat64(float64(self.Inner.Dy()) / 2)))
}

func (self *List) ScrollTop() {
	self.SelectedRow = 0
}

func (self *List) realDy() int {
	real := self.Inner.Dy()

	point := self.Inner.Min
	for row := self.topRow; row < len(self.Rows) && point.Y < self.Inner.Max.Y; row++ {
		cells := ParseStyles(self.Rows[row], self.TextStyle)
		cells = WrapCells(cells, uint(self.Inner.Dx()))
		for j := 0; j < len(cells) && point.Y < self.Inner.Max.Y; j++ {
			if cells[j].Rune == '\n' {
				real--
				point = image.Pt(self.Inner.Min.X, point.Y+1)
			} else {
				if point.X+1 == self.Inner.Max.X+1 && len(cells) > self.Inner.Dx() {
					break
				} else {
					point = point.Add(image.Pt(rw.RuneWidth(cells[j].Rune), 0))
				}
			}
		}
		point = image.Pt(self.Inner.Min.X, point.Y+1)
	}

	return real
}

func (self *List) ScrollBottom() {
	self.SelectedRow = len(self.Rows) - self.realDy()
	if self.SelectedRow < 0 {
		self.SelectedRow = 0
	}
}
