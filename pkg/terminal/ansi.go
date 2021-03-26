package terminal

import (
  "fmt"
  "os"
)

const ESC = "\033"

func csi1(n int, char byte) {
  fmt.Fprintf(os.Stdout, "%s[%d%c", ESC, n, char)
}

func csi2(n int, m int, char byte) {
  fmt.Fprintf(os.Stdout, "%s[%d;%d%c", ESC, n, m, char)
}

func esc1(c byte) {
  fmt.Fprintf(os.Stdout, "%s[%c", ESC, c)
}

func control(char byte) {
  fmt.Fprintf(os.Stdout, "%c", char)
}

func moveUpDown(d int) {
  if d < 0 {
    csi1(-d, 'F')
  } else if d > 0 {
    csi1(d, 'E')
  }
}

func moveUp(d int) {
  if d > 0 {
    csi1(d, 'F')
  }
}

func moveLeft() {
  csi1(1, 'D')
}

func moveRight() {
  csi1(1, 'C')
}

func clearScreen() {
  csi1(2, 'J')
}

func moveToRowStart() {
  csi1(1, 'G')
}

func moveToScreenStart() {
  csi2(1, 1, 'H')
}

func moveToRow(r int) {
  csi2(r, 1, 'H')
}

func clearRow() {
  csi1(2, 'K')
}

func clearRowAfterCursor() {
  csi1(0, 'K')
}

func clearRows(d int) {
  for i := 0; i < d; i ++ {
    csi1(2, 'K')

    csi1(1, 'F')
  }
}

func moveDown(d int) {
  if d > 0 {
    csi1(d, 'E')
  }
}

// input: 0-based
// moves to 1-based
func moveToCol(x int) {
  csi1(x + 1, 'G') 
}

func savePos() {
  esc1('s')
}

func restorePos() {
  esc1('u')
}

func getPos() {
  csi1(6, 'n')
}
