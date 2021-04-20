package wwwserver

import (
  "net/http"
)

type Resource interface {
	Serve(resp *ResponseWriter, req *http.Request) error
  ServeFrozen(resp *ResponseWriter, req *http.Request) error
  Freeze() error
}
