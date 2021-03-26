package terminal

import (
  "fmt"
  //"io"
  "io/ioutil"
  "math"
  "os"
  "os/exec"
  "regexp"
  "strconv"
  "strings"
  "time"

  "golang.org/x/term"
)

// human reaction times are an order of magnitude slower than this
// and auto generated stdin bytes are an order of magnitude faster than this
const WIDTH_POLLING_INTERVAL = 10*time.Millisecond
var EFFICIENT bool = true // false -> redraw upon every action, true -> use the terminal efficiently
const XCLIP = "/usr/bin/xclip"

var debug *os.File = nil

type Handle interface {
  Eval(line string) (out string, hist string) // empty return strings are not stored in history
}

// custom terminal prompt, because the existing discoverable terminal libraries suck
type Terminal struct {
  handler    Handle

  prompt      string   // "> " by default

  history     [][]byte // simply keep everything, it doesn't matter
  historyDir  string   // directory where to store history files
  historyIdx  int      // -1 for last
  historyFile *os.File // open history file, so we can keep appending
  backup      []byte // we can go into a history line, and start editing it

  phraseRe *regexp.Regexp

  reader    *StdinReader

  buffer    []byte // input bytes are accumulated
  status    []byte

  cursor    int // -1 for end of line, TODO: multiline
  onEnd     func()
  width     int
  height    int
  promptRow int // 0-based
}

func NewTerminal(handler Handle) *Terminal {
  t := &Terminal{
    handler: handler,
    prompt: "> ",
    historyDir: "",
    history: make([][]byte, 0),
    historyIdx: -1,
    historyFile: nil,
    phraseRe: regexp.MustCompile(`([0-9a-zA-Z_\-\.]+)`),
    cursor: 0,
    reader: NewStdinReader(),
    buffer: nil,
    backup: nil,
    onEnd: nil,
    width: 0,
    height: 0,
    promptRow: 0,
  }

  return t
}

func (t *Terminal) Width() int {
  return t.width
}

func (t *Terminal) Height() int {
  return t.height
}

func (t *Terminal) Run() error {
  // the terminal needs to be in raw mode, so we can intercept the control sequences
  // (the default canonical mode isn't good enough for repl's)
  if err := t.SetRaw(); err != nil {
    return err
  }

  var err error
  debug, err = os.Create("term_debug.log")
  if err != nil {
    panic(err)
  }

  t.reader.Start()

  t.notifySizeChange()

  t.printPrompt()

  getPos() // get initial prompt position

  // loop forever
  for {
    t.reader.Read()

    bts := <- t.reader.Chan

    t.dispatch(bts)
  }

  return nil
}

func (t *Terminal) SetRaw() error  {
  // we need the term package as a platform independent way of setting the connected terminal emulator to raw mode
  oldState, err := term.MakeRaw(0)
  if err != nil {
    return err
  }

  t.onEnd = func() {
    term.Restore(0, oldState)
  }

  return nil
}

func (t *Terminal) UnsetRaw() {
  t.onEnd()

  t.onEnd = nil
}

func (t *Terminal) notifySizeChange() {
  getSize := func() (int, int) {
    w, h, err := term.GetSize(0)
    if err != nil {
      panic(err)
    }

    return w, h
  }

  t.width, t.height = getSize()

  go func() {
    for {
      <- time.After(WIDTH_POLLING_INTERVAL)

      newW, newH := getSize()

      if t.width != newW || t.height != newH {
        t.changeSize(newW, newH)
      }
    }
  }()
}

// turn stdin bytes into something useful
func (t *Terminal) dispatch(b []byte) {
  n := len(b)

  fmt.Fprintln(debug, "keypress: ", b)

  if n == 1 {
    switch b[0] {
    case 0: // NULL, or CTRL-2
      return 
    case 1: // CTRL-A
      t.moveToLineStart()
    case 2: // CTRL-B
      t.moveLeftOneChar()
    case 3: // CTRL-C
      t.ignoreLine()
    case 4: // CTRL-D
      t.quit()
    case 5: // CTRL-E
      t.moveToLineEnd()
    case 6: // CTRL-F
      t.moveRightOneChar()
    case 8: // CTRL-H
      t.backspace()
    case 9: // TAB
      t.tab()
    case 10:
      t.addToLine([]byte{'\n', '\r'})
    case 11:
      t.clearRestOfLine()
    case 12: // CTRL-L
      t.clearScreen()
    case 13: // RETURN
      t.newLine()
    case 14: // CTRL-N
      t.historyForward()
    case 16: // CTRL-P
      t.historyBack()
    case 17:
      t.clearOnePhraseRight()
    case 18: // CTRL-R
      t.startReverseSearch()
    case 21: // CTRL-U
      t.clearToStart()
    case 22: // CTRL-V : paste clipboard
      t.pasteClipboard()
    case 25: // CTRL-Y : yank (copy) selection
      // so text should be selectable
      return


    case 23: // CTRL-W
      t.clearOnePhraseLeft()
    case 27: // ESC
      t.ignoreLine()
    case 127: // BACKSPACE
      t.backspace()
    default:
      if b[0] >= 32 {
        t.addToLine(b)
      }
    }
  } else if n == 2 && b[0] == 195 {
    // ALT + KEY
  } else if n > 2 && b[0] == 27 && b[1] == 79 { // [ESCAPE, O, ...]
    switch b[2] {
    case 80: // F1
    case 81: // F2
    // ...
    default:
      // function keys not yet supported
    }
  } else if n > 2 && b[0] == 27 && b[1] == 91 { // [ESCAPE, OPEN_BRACKET, ...]
    if n == 3 {
      switch b[2] {
      case 65:
        t.historyBack()
      case 66:
        t.historyForward()
      case 67: // ArrowRight
        t.moveRightOneChar()
      case 68: // ArrowLeft
        t.moveLeftOneChar()
      case 72:
        t.moveToLineStart()
      case 70:
        t.moveToLineEnd()
      }
    } else if n == 4 {
      if b[2] == 51 && b[3] == 126 {
        t.deleteChar()
      }
    } else if n == 6 && b[2] == 49 && b[3] == 59 {
      if b[4] == 53 && b[5] == 68 { // CTRL-ArrowLeft
        t.moveLeftOnePhrase()
      } else if b[4] == 53 && b[5] == 67 {
        t.moveRightOnePhrase()
      }
    } else if len(b) > 5 && b[n-1] == 82 {
      pp, err := strconv.Atoi(strings.Split(string(b[2:n-1]), ";")[0])
      if err == nil {
        t.updatePromptRow(pp - 1)
      }
    }
  } else {
    //t.cleanAndAddToLine(b)
  }

  return
}

func (t *Terminal) printPrompt() {
  moveToRowStart()
  //csi1(1, 'G')
  fmt.Print(t.prompt)
}

func (t *Terminal) resetLine() {
  t.cursor = 0 
  t.buffer = make([]byte, 0)
  t.printPrompt()
}

func (t *Terminal) addToLine(b []byte) {
  _, y0 := t.bufferPosCoord(t.cursor)
  _, y1 := t.bufferPosCoord(t.cursor + len(b))
  if t.cursor == len(t.buffer) && y0 == y1 && EFFICIENT {
    t.buffer = append(t.buffer, b...)
    t.cursor = len(t.buffer)
    fmt.Print(string(b))
  } else {
    oldN := len(t.buffer)
    extraN := len(b)
    newN := oldN + extraN

    //aft := t.buffer[t.cursor:len(t.buffer)]
    newBuffer := make([]byte, len(t.buffer) + len(b))
    newCursor := t.cursor + extraN 

    for i := 0; i < newN; i++ {
      if i < t.cursor {
        newBuffer[i] = t.buffer[i]
      } else if i < newCursor {
        newBuffer[i] = b[i - t.cursor]
      } else {
        newBuffer[i] = t.buffer[i - extraN]
      }
    }

    t.forceLine(newBuffer, newCursor)
  }
}

// 0 based indexing!
func (t *Terminal) bufferPosCoord(cursor int, args ...int) (int, int) {
  w := t.Width()
  if len(args) == 1 {
    w = args[0]
  } else if len(args) != 0 {
    panic("expected 0 or 1 args")
  }

  y := int(math.Floor(float64(len(t.prompt) + cursor)/float64(w)))
  x := len(t.prompt) + cursor - y*w

  return x, y
}

func (t *Terminal) cursorCoord() (int, int, int) {
  x, y := t.bufferPosCoord(t.cursor)
  
  _, yEnd := t.bufferPosCoord(len(t.buffer) - 1) // end cursor is on last char, not one past last char!

  return x, y, yEnd - y
}

func (t *Terminal) clearLine() {
  _, y, dy := t.cursorCoord()

  moveUpDown(dy)

  clearRows(y + dy)

  clearRow()

  t.resetLine()
}

func copyBytes(b []byte) []byte {
  l := make([]byte, len(b))

  for i, c := range b {
    l[i] = c
  }

  return l
}

func (t *Terminal) changeSize(newW int, newH int) {
  // backup of the current line
  cursor := t.cursor
  line := copyBytes(t.buffer)

  // clear the line using the old width

  t.clearLine()

  // change the width
  t.width = newW
  t.height = newH


  t.forceLine(line, cursor)

  _, y, _ := t.cursorCoord()
  if (t.promptRow >= t.height - y) {
    t.updatePromptRow(t.height - y - 1)
  }
}

// this works for a single line
func (t *Terminal) forceLine(line []byte, cursor int) {
  t.clearLine()

  t.buffer = line
  fmt.Print(string(line))

  if cursor >= len(line) {
    cursor = len(line)
  }

  t.cursor = cursor

  fmt.Fprintf(debug, "cursor: %d, bufferLen: %d, width: %d\n", t.cursor, len(t.buffer), t.Width())

  x, _, dy := t.cursorCoord()
  moveUpDown(-dy)

  fmt.Fprintf(debug, "moved y by %d, move to x %d\n", -dy, x+1)

  _, yp := t.bufferPosCoord(len(t.buffer))
  if t.promptRow + yp >= t.Height() {
    t.updatePromptRow(t.Height() - yp)
  }

  moveToCol(x)
}

func (t *Terminal) newLine() {
  line := t.buffer
  fmt.Print("\n\r")

  out, history_ := t.handler.Eval(string(line))
  if len(out) > 0 {
    fmt.Print(out)
    fmt.Print("\n\r")
  }

  history := []byte(history_)

  if len(history) != 0 {
    t.appendToHistory(history)
    t.historyIdx = -1
  }

  t.backup = nil

  t.resetLine()
}

func (t *Terminal) moveToLineEnd() {
  line := copyBytes(t.buffer)
  t.forceLine(line, len(line))
}

func (t *Terminal) moveToLineStart() {
  line := copyBytes(t.buffer)
  t.forceLine(line, 0)
}

func (t *Terminal) moveLeftOneChar() {
  if len(t.buffer) > 0 && t.cursor > 0 {
    x, _, _ := t.cursorCoord()

    if x > 0 && EFFICIENT {
      t.cursor -= 1
      moveLeft()
      //csi1(1, 'D')
    } else {
      line := copyBytes(t.buffer)
      t.forceLine(line, t.cursor - 1)
    }
  }
}

func (t *Terminal) moveRightOneChar() {
  if t.cursor < len(t.buffer) {
    x, _, _ := t.cursorCoord()

    if x < t.Width() - 1 && EFFICIENT {
      t.cursor += 1
      moveRight()
      //csi1(1, 'C')
    } else {
      line := copyBytes(t.buffer)
      t.forceLine(line, t.cursor + 1)
    }
  }
}

func (t *Terminal) moveLeftOnePhrase() {
  newCursor, ok := t.prevPhrasePos()
  if ok {
    _, y0 := t.bufferPosCoord(t.cursor)
    x1, y1 := t.bufferPosCoord(newCursor)

    if y0 == y1 && EFFICIENT && x1 > 0 {
      t.cursor = newCursor
      moveToCol(x1)
    } else {
      line := copyBytes(t.buffer)
      t.forceLine(line, newCursor)
    }
  }
}

func (t *Terminal) moveRightOnePhrase() {
  newCursor, ok := t.nextPhrasePos()
  if ok {
    _, y0 := t.bufferPosCoord(t.cursor)
    x1, y1 := t.bufferPosCoord(newCursor)

    if y0 == y1 && EFFICIENT && x1 < t.Width() - 1 {
      t.cursor = newCursor
      moveToCol(x1)
    } else {
      line := copyBytes(t.buffer)
      t.forceLine(line, newCursor)
    }
  }
}

// dont append if the same as the previous
func (t *Terminal) appendToHistory(line []byte) {
  n := len(t.history)

  if n == 0 {
    t.history = append(t.history, line)
  } else if string(t.history[n-1]) != string(line) {
    t.history = append(t.history, line)
  }
}

func (t *Terminal) historyForward() {
  if t.historyIdx != -1 {
    if t.historyIdx < len(t.history) - 1 {
      t.historyIdx += 1

      line := copyBytes(t.history[t.historyIdx])

      t.forceLine(line, len(line))
    } else {
      t.forceLine(copyBytes(t.backup), len(t.backup))

      t.backup = nil

      t.historyIdx = -1
    }
  } 
}

func (t *Terminal) historyBack() {
  if t.historyIdx == -1 {
    if len(t.history) > 0 {
      t.historyIdx = len(t.history) - 1

      t.backup = t.buffer

      line := copyBytes(t.history[t.historyIdx])
      t.forceLine(line, len(line))
    }
  } else if t.historyIdx > 0 {
    t.historyIdx -= 1

    line := copyBytes(t.history[t.historyIdx])
    t.forceLine(line, len(line))
  }
}

func (t *Terminal) startReverseSearch() {
  getPos()

  //t.writeStatus()
}

func (t *Terminal) tab() {
}

func (t *Terminal) quit() {
  fmt.Print("\n\r")

  moveToRowStart()

  //csi1(1, 'G')

  t.UnsetRaw()

  os.Exit(0)
}

func (t *Terminal) ignoreLine() {
  _, y, dy := t.cursorCoord()

  moveDown(dy)

  // clear whole line
  fmt.Print("\n\r")

  t.updatePromptRow(t.promptRow + y + dy + 1)

  t.resetLine()
}

func (t *Terminal) clearScreen() {
  clearScreen()

  moveToScreenStart()

  t.updatePromptRow(0)

  t.resetLine()
}

func (t *Terminal) backspace() {
  n := len(t.buffer)

  if n > 0 {
    if t.cursor > 0 {
      newCursor := t.cursor - 1
      newBuffer := append(t.buffer[0:newCursor], t.buffer[newCursor+1:len(t.buffer)]...)

      _, y0 := t.bufferPosCoord(t.cursor)
      x1, y1 := t.bufferPosCoord(newCursor)

      if y0 == y1 && t.cursor == len(t.buffer) && EFFICIENT {
        moveToCol(x1)
        clearRowAfterCursor()
        //csi1(0, 'K')
        t.buffer = newBuffer
        t.cursor = newCursor
      } else {
        t.forceLine(newBuffer, newCursor)
      }
    }
  }
}

func (t *Terminal) deleteChar() {
  if t.cursor != len(t.buffer) {
    var newBuffer []byte
    if t.cursor == len(t.buffer) - 1 {
      newBuffer = t.buffer[0:t.cursor] 
    } else {
      newBuffer = append(t.buffer[0:t.cursor], t.buffer[t.cursor+1:len(t.buffer)]...)
    }

    t.forceLine(newBuffer, t.cursor)
  }
}

func (t *Terminal) clearRestOfLine() {
  if t.cursor != len(t.buffer) {
    newBuffer := t.buffer[0:t.cursor]

    t.forceLine(newBuffer, t.cursor)
  }
}

func (t *Terminal) clearToStart() {
  if t.cursor > 0 {
    newBuffer := t.buffer[t.cursor:len(t.buffer)]

    t.forceLine(newBuffer, 0)
  }
}

func (t *Terminal) phraseStartPositions() []int {
  if len(t.buffer) == 0 {
    return []int{0}
  }

  re := t.phraseRe

  indices := re.FindAllIndex(t.buffer, -1)

  res := make([]int, 0)

  for i, match := range indices {
    start := match[0]
    stop := match[1]
    if i == 0 && start != 0 {
      res = append(res, 0)
    }

    res = append(res, start, stop)

    if i == len(indices) - 1 && stop != len(t.buffer) {
      res = append(res, len(t.buffer))
    }
  }

  if len(res) == 0 || res[len(res)-1] != len(t.buffer) {
    res = append(res, len(t.buffer))
  }

  return res
}

func (t *Terminal) nextPhrasePos() (int, bool) {
  var res int
  if (t.cursor == len(t.buffer)) {
    res = len(t.buffer)
  } else {
    indices := t.phraseStartPositions()

    for _, idx := range indices {
      if idx > t.cursor {
        res = idx
        break
      }
    }
  }

  return res, res != t.cursor
}

func (t *Terminal) prevPhrasePos() (int, bool) {
  var res int
  if (t.cursor == 0) {
    res = 0
  } else {
    indices := t.phraseStartPositions()

    for i := len(indices) - 1; i >= 0; i-- {
      idx := indices[i]
      if idx < t.cursor {
        res = idx
        break
      }
    }
  }

  return res, res != t.cursor
}

func (t *Terminal) clearOnePhraseLeft() {
  idx, ok := t.prevPhrasePos()
  if ok {
    newCursor := idx
    newBuffer := append(t.buffer[0:idx], t.buffer[t.cursor:len(t.buffer)]...)

    _, y0 := t.bufferPosCoord(t.cursor)
    x1, y1 := t.bufferPosCoord(newCursor)

    if t.cursor == len(t.buffer) && y0 == y1 && EFFICIENT && x1 > 0 {
      t.cursor = newCursor
      t.buffer = newBuffer
      moveToCol(x1)
      clearRowAfterCursor()
    } else {
      t.forceLine(newBuffer, newCursor)
    }
  }
}

func (t *Terminal) clearOnePhraseRight() {
  idx, ok := t.nextPhrasePos()
  if ok {
    newCursor := t.cursor
    newBuffer := append(t.buffer[0:t.cursor], t.buffer[idx:len(t.buffer)]...)

    t.forceLine(newBuffer, newCursor)
  }
}

func haveXClip() bool {
  if info, err := os.Stat(XCLIP); os.IsNotExist(err) {
    return false
  } else if info.Mode() & 0111 == 0 {
    return false
  } else {
    return true
  }
}

func (t *Terminal) pasteClipboard() {
  // only works if xclip is installed
  if !haveXClip() {
    return
  }

  cmd := exec.Command(XCLIP, "-o")

  cmdOut, err := cmd.StdoutPipe()
  if err != nil {
    panic(err)
  }

  if err := cmd.Start(); err != nil {
    return
  }

  outMsg, _ := ioutil.ReadAll(cmdOut)

  t.cleanAndAddToLine(outMsg)
}

func (t *Terminal) cleanAndAddToLine(msg []byte) {
  // remove bad chars
  // what about unicode?
  filtered := make([]byte, 0)

  for _, c := range msg {
    if c == '\n' || c == '\t' {
      filtered = append(filtered, ' ')
    } else if c >= 32 && c < 127 {
      filtered = append(filtered, c)
    }
  }

  t.addToLine(filtered)
}

func (t *Terminal) updatePromptRow(pr int) {
  if pr >= t.height {
    pr = t.height - 1
  }

  t.promptRow = pr

  fmt.Fprintf(debug, "prompt row %d/%d\n", t.promptRow, t.Height())
}

// when typing ctrl-R, we should start typing in the bottom left buffer
func (t *Terminal) writeStatus() {
  savePos()

  moveToRow(t.height)

  fmt.Print(string(t.status))

  restorePos()
}


