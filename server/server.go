package server

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

//func NewRouter() http.Handler {
//	r := mux.NewRouter()
//	r.HandleFunc("/last", Resp)
//	//r.HandleFunc("/+", handleHTTP).Methods("GET", "POST")
//	return r
//}

func NewServer(port string) *http.Server {

	return &http.Server{
		Addr:         port,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		//Handler:      NewRouter(),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				//rout := NewRouter()
				//http.Handle("/", rout)
				if r.URL.String() == "/last" {
					Resp(w, r)
				} else {
					handleHTTP(w, r)
				}
			}
		}),
	}
}

func Resp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!?!!!!"))
	return
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error

	switch r.Method {
	case "GET":
		resp, err = http.Get(r.URL.String())
	case "POST":
		resp, err = http.Post(r.URL.String(), r.Header.Get("Content-Type"), r.Body)
	default:
		resp, err = http.Get(r.URL.String())
	}

	if err != nil {
		return
	}
	defer resp.Body.Close()
	for mime, val := range resp.Header {
		w.Header().Set(mime, val[0])
	}
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type")+"; charset=utf8")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
