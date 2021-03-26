package terminal

import (
  "bufio"
  "os"
  "time"
  "sync"
)

const MACHINE_INTERVAL = time.Millisecond

// InputReader collects inputs 
type StdinReader struct {
  reader       *bufio.Reader
  lastTime     time.Time
  buffer       []byte
  lock         *sync.Mutex

  Chan         chan []byte
}

func NewStdinReader() *StdinReader {
  return &StdinReader{
    reader:      nil,
    lastTime:    time.Time{},
    buffer:      make([]byte, 0),
    lock:        &sync.Mutex{},

    Chan:        make(chan []byte),
  }
}

func (r *StdinReader) Start() {
  go func() {
    for {
      <- time.After(MACHINE_INTERVAL)

      r.lock.Lock()

      if len(r.buffer) > 0 {
        if time.Now().After(r.lastTime.Add(MACHINE_INTERVAL)) {
          msg := r.buffer

          r.buffer = make([]byte, 0)

          r.Chan <- msg
        }
      }

      r.lock.Unlock()
    }
  }()
}

func (r *StdinReader) Read() {
  if r.reader != nil {
    return 
  }

  r.reader = bufio.NewReader(os.Stdin)
  r.lastTime = time.Now()

  go func() {
    for {
      b, err := r.reader.ReadByte()
      if err != nil {
        panic(err)
      }

      stopNow := false
      if b == 13 && time.Now().After(r.lastTime.Add(MACHINE_INTERVAL)) {
        // it is unlikely that a carriage return followed by some text is pasted into the terminal, so we can use this as a queu to quit
        stopNow = true
      }

      r.lastTime = time.Now()

      r.lock.Lock()

      r.buffer = append(r.buffer, b)

      r.lock.Unlock()

      if stopNow {
        r.reader = nil
        return
      }
    }
  }()
}
