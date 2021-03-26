package terminal

import (
  "fmt"
  "os"
)

const ESC = "\033"

// structure the terminal window as a rectangle of characters
type Canvas struct {
  statusHeight int // usually 1
  statusBuffer []byte

  col int // 1-based
  row int // 1-based
  width int
  innerHeight int
}

func NewCanvas() *Canvas {
  return &Canvas{1, 0, 0, 1, 1}
}

func (c *Canvas) UpdateSize(w int, h int) {
  c.ClearStatus()

  c.width = w
  c.innerHeight = h - c.statusHeight

  if c.col > c.width {
    c.col = c.width
  } 

  if c.row > c.innerHeight {
    c.row = c.innerHeight
  }

  c.PrintStatus()

  c.SyncCursor()
}

func (c *Canvas) UpdatePos(col int, row int) {
  c.col = col
  c.row = row

  c.SyncCursor()
}

func (c *Canvas) Pos() (int, int) {
  return c.col, c.row
}

func (c *Canvas) Width() int
  return c.width
}

// dont return height, because that should be completely managed by this struct

func (c *Canvas) SyncCursor() {
  fmt.Fprintf(os.Stdout, "%s[%d;%dH", c.row, c.col)
}

func (c *Canvas) ClearScreen() {
  fmt.Fprintf(os.Stdout, "%s[2J", ESC)
  c.col = 1
  c.row = 1
  c.SyncCursor()
}

func (c *Canvas) ClearLine() {
  fmt.Fprintf(os.Stdout, "%s[2K" ESC)
  c.col = 1
  c.SyncCursor()
}

func (c *Canvas) ClearStatus() {
  for i := 1; i <= c.statusHeight; i++ {
    fmt.Fprintf(os.Stdout, "%s[%d;%dH%s[2K", ESC, c.innerHeight + i, c.col, ESC)
  }

  c.SyncCursor()
}

func (c *Canvas) CursorUp() {
  if c.row > 1 {
    c.row -= 1
    fmt.Fprintf(os.Stdout, "%s[1A", ESC)
  }
}

func (c *Canvas) CursorLeft() {
  if c.col > 1 {
    c.col -= 1
    fmt.Fprintf(os.Stdout, "%s[1D", ESC)
  } 
}

func (c *Canvas) CursorRight() {
  if c.col < c.width {
    c.col += 1
    fmt.Fprintf(os.Stdout, "%s[1C", ESC)
  } 
}

func (c *Canvas) CursorDown() {
  if c.row < c.innerHeight {
    c.row += 1
    fmt.Fprintf(os.Stdout, "%s[1B", ESC)
  } 
}

func (c *Canvas) WriteByte(b byte) {
  if b == '\n' {
    c.row += 1

    if c.row > c.innerHeight {
      c.ScrollUp() // automatically reprints status buffer
      c.row = c.innerHeight
    }

    c.SyncCursor()
  } else {
    // should be a printable character
    fmt.Fprintf(os.Stdout, "%c", b)
    
    c.col += 1
    if c.col > c.width {
      c.col = 1
      c.row += 1

      if c.row > c.innerHeight {
        c.row = c.innerHeight
      }
    }
  }
}

// for the selection state
func (c *Canvas) HighlightBytes(bs []byte) {
  fmt.Fprintf(os.Stdout, "%s[47m%s[30m", ESC, ESC)

  for _, b := range bs {
    c.WriteByte(b)
  }

  fmt.Fprintf(os.Stdout, "%s[0m", ESC)
}
