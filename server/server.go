package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Server struct {
	Serv http.Server
	Db   *pgxpool.Pool
}

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Close() error {
	b.Buffer.Reset()
	return nil
}

//var okHeader = []byte("HTTP/1.1 200 OK\r\n\r\n")

func NewServer(port string, db *pgxpool.Pool) Server {

	return Server{http.Server{
		Addr:         port,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				DbPattern := `^/[0-9]+$`
				TestPattern := `^/[0-9]+/test$`
				HackPattern := `^/hack`
				if match, _ := regexp.Match(DbPattern, []byte(r.URL.String())); match {
					fmt.Println(r.URL, "  ", match)
					Resp(w, r, db)
				} else if match, _ := regexp.Match(TestPattern, []byte(r.URL.String())); match {
					testUrl(w, r, db)
				} else if match, _ := regexp.Match(HackPattern, []byte(r.URL.String())); match {
					ForHack(w, r, db)
				} else {
					handleHTTP(w, r, db)
				}
			}
		}),
	}, db}
}

var Patterns = []string{string('"'), "'"}
var PatternNum = regexp.MustCompile("[0-9]+")

func testUrl(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	//http://127.0.0.1:8080/186/test
	ind := PatternNum.FindAllString(r.URL.String(), -1)[0]
	req := GetReq(ind, db)
	oldUrl := req.URL.RawQuery
	var oldBody io.ReadWriteCloser
	oldBody = &buffer{}
	io.Copy(oldBody, req.Body)
	resp, err := http.DefaultTransport.RoundTrip(&req)
	if err != nil {
		fmt.Println("Inject error", err)
		return
	}
	PureCode := resp.StatusCode
	PureLen := resp.ContentLength
	for _, patern := range Patterns {
		for key, val := range req.URL.Query() {
			old := key + "=" + val[0]
			req.URL.RawQuery = strings.Replace(req.URL.RawQuery, old, old+patern, -1)
			//fmt.Println(req.URL)
			resp, err := http.DefaultTransport.RoundTrip(&req)
			if err != nil {
				fmt.Println("Inject error", err)
				return
			}
			req.URL.RawQuery = oldUrl
			if resp.StatusCode != PureCode || resp.ContentLength != PureLen {
				fmt.Println("CODE ", resp.StatusCode,
					" LEN ", resp.ContentLength, " Param for inj is ", key)
			}
		}
		//buf := &buffer{}
		body, err := ioutil.ReadAll(oldBody)
		if err != nil {
			fmt.Println("Error reading body:", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		for _, val := range strings.Split(string(body), "&") {
			var rwc io.ReadWriteCloser
			rwc = &buffer{}
			rwc.Write([]byte(strings.Replace(string(body), val, val+patern, -1)))
			req.Body = rwc
			resp, err := http.DefaultTransport.RoundTrip(&req)
			if err != nil {
				fmt.Println("Inject error", err)
				return
			}
			rwc.Read([]byte{})
			rwc.Write(body)
			req.Body = rwc
			if resp.StatusCode != PureCode || resp.ContentLength != PureLen {
				fmt.Println("CODE ", resp.StatusCode,
					" LEN ", resp.ContentLength, " Param for inj is ", val)
			}
		}
	}

	w.Write([]byte("Tested"))
	return
}

func Resp(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	ind := r.URL.String()[1:]
	req := GetReq(ind, db)
	http.Redirect(w, &req, req.URL.String(), 301)
	return
}

func handleHTTP(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	var resp *http.Response
	var err error
	err = LogRequest(r, db)
	if err != nil {
		return
	}
	//fmt.Println("URL:", r.URL)
	//fmt.Println("HOST:", r.Host)
	//fmt.Println("URL HOST:", r.URL.Host)
	switch r.Method {
	case "GET":
		resp, err = http.DefaultTransport.RoundTrip(r)
	case "POST":
		resp, err = http.Post(r.URL.String(), r.Header.Get("Content-Type"), r.Body)
	default:
		resp, err = http.Get(r.URL.String())
	}
	if err != nil {
		fmt.Println("after handle", err)
		return
	}
	defer resp.Body.Close()
	for mime, val := range resp.Header {
		if mime == "Proxy-Connection" {
			continue
		}
		w.Header().Set(mime, val[0])
	}
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type")+"; charset=utf8")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	//1)search for the host cert already created
	//2)gen new host cert
	//3)subscribe cert with root cert

	//HostName := r.Host
	//CertPath := "certs/" + HostName + ".crt"
	//out, err := exec.Command("/bin/sh", "gen_host_cert.sh",
	//	HostName, strconv.Itoa(100)).Output()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//err = ioutil.WriteFile(CertPath, out, 0644)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//config, err := newClientConfig(CertPath)
	//if err != nil {
	//	fmt.Print(err)
	//	return
	//}
	//config.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	//	cConfig := new(tls.Config)
	//	cConfig.ServerName = hello.ServerName
	//	_, err := tls.Dial("tcp", r.Host, cConfig)
	//	if err != nil {
	//		log.Println("dial", r.Host, err)
	//		return nil, err
	//	}
	//	fmt.Println("1234")
	//	fmt.Println(config.Certificates)
	//	return &config.Certificates[0], nil
	//}

	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	//dest_conn_TLS := tls.Client(dest_conn, config)
	//err = dest_conn_TLS.Handshake()
	//if err != nil {
	//	fmt.Print(err)
	//	http.Error(w, err.Error(), http.StatusServiceUnavailable)
	//	return
	//}

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
	//client_conn.Write(okHeader)

	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	//var buf io.ReadWriteCloser
	//defer buf.Close()
	//io.Copy(buf, source)
	//fmt.Println(buf)
	//io.Copy(destination, buf)
	io.Copy(destination, source)

}

func newClientConfig(rootCAPath string) (*tls.Config, error) {
	pemBytes, err := ioutil.ReadFile(rootCAPath)
	if err != nil {
		return nil, err
	}
	//rootca := x509.NewCertPool()
	//ok := rootca.AppendCertsFromPEM(pemBytes)
	//if !ok {
	//	return nil, err
	//}
	cert := tls.Certificate{
		Certificate: [][]byte{pemBytes},
	}

	return &tls.Config{
		//RootCAs:            rootca,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}, nil
}

func LogRequest(r *http.Request, db *pgxpool.Pool) error {
	sql := `INSERT INTO requests VALUES(default,$1,$2,$3,$4)`
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	headers := ""
	for key, val := range r.Header {
		headers += key + ": " + val[0] + "\n"
	}
	queryResult, err := db.Exec(context.Background(), sql,
		r.Method, r.URL.String(), headers, string(body))
	affected := queryResult.RowsAffected()
	if (affected != 1) || (err != nil) {
		fmt.Print(err)
		return err
	}
	return nil
}

func GetReq(ind string, db *pgxpool.Pool) http.Request {
	var req = http.Request{}
	id := 0
	URL := ""
	headers := ""
	body := ""
	var rwc io.ReadWriteCloser
	rwc = &buffer{}
	sql := `select * from requests where id = $1`
	queryResult := db.QueryRow(context.Background(), sql, ind)
	err := queryResult.Scan(&id, &req.Method, &URL, &headers, &body)
	if err != nil {
		fmt.Println(err)
		return http.Request{}
	}
	_, err = rwc.Write([]byte(body))
	if err != nil {
		fmt.Println(err)
		return http.Request{}
	}
	hed := make(map[string][]string)
	for _, val := range strings.Split(headers, "\n") {
		if val != "" {
			buf := strings.Split(val, ":")
			hed[buf[0]] = []string{buf[1]}
		}
	}
	req.URL, _ = url.Parse(URL)
	req.Header = hed
	req.Body = rwc
	return req
}

func ForHack(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	//http://127.0.0.1:8080/hack?id=15%20UNION+SELECT+1,+%27a%27,+version(),+%27a:a%27,+version()
	ind := r.URL.Query()["id"][0]
	URL := ""
	sql := "select URL from requests where id = " + ind
	//fmt.Println(sql)
	queryResult := db.QueryRow(context.Background(), sql)
	err := queryResult.Scan(&URL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "", 400)
		return
	}
	w.Write([]byte(URL))
	return
}
