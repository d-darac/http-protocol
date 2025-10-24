package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/d-darac/httpfromtcp/internal/headers"
	"github.com/d-darac/httpfromtcp/internal/request"
	"github.com/d-darac/httpfromtcp/internal/response"
	"github.com/d-darac/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		h := headers.NewHeaders()
		h.Override("Content-Type", "text/html")
		html := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		w.WriteStatusLine(response.BadRequest)
		w.WriteHeaders(h)
		w.WriteBody([]byte(html))
	case "/myproblem":
		h := headers.NewHeaders()
		h.Override("Content-Type", "text/html")
		html := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		w.WriteStatusLine(response.InternalServerError)
		w.WriteHeaders(h)
		w.WriteBody([]byte(html))
	default:
		h := headers.NewHeaders()
		h.Override("Content-Type", "text/html")
		html := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
		w.WriteStatusLine(response.OK)
		w.WriteHeaders(h)
		w.WriteBody([]byte(html))
	}
}
