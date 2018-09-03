package inasnap

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ryanjyoder/couchdb"
)

type Configs struct {
	CouchURL      string
	CouchUser     string
	CouchPassword string
	CouchDBName   string
	ListenPort    string
}

type APIServer interface {
	Run() error
}

type server struct {
	couchdb couchdb.DatabaseService
	port    int64
}

func NewAPIServer(configs Configs) (APIServer, error) {
	s := server{}
	p, err := strconv.ParseInt(configs.ListenPort, 10, 64)
	if err != nil {
		return nil, err
	}
	s.port = p

	u, err := url.Parse(configs.CouchURL)
	if err != nil {
		return nil, err
	}
	client, err := couchdb.NewAuthClient(configs.CouchUser, configs.CouchPassword, u)
	if err != nil {
		return nil, err
	}

	s.couchdb = client.Use(configs.CouchDBName)

	return &s, nil

}

func (s *server) Run() error {

	http.HandleFunc("/snap/", s.snapHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *server) snapHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		s.createHandler(w, req)
	case "GET":
		s.statusHandler(w, req)
	default:
		io.WriteString(w, "bad method\n")
	}
}

type createRequest struct {
	Snap   string `json:"snap"`
	Port   int64  `json:"port"`
	Domain string `json:"domain"`
}

func (r createRequest) GetID() string {
	return ""
}
func (r createRequest) GetRev() string {
	return ""
}

func (s *server) createHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		io.WriteString(w, "could read request:"+err.Error()+"\n")
		return
	}

	reqBody := createRequest{}
	json.Unmarshal(body, &reqBody)

	resp, err := s.couchdb.Post(reqBody)
	if err != nil {
		io.WriteString(w, "could not begin deployment"+err.Error()+"\n")
		return
	}
	io.WriteString(w, fmt.Sprintf(`{"appkey":"%s"}`, resp.ID)+"\n")
}

type appJob struct {
	ID     string `json:"_id"`
	Rev    string `json:"_rev"`
	Snap   string `json:"snap"`
	Port   int64  `json:"port"`
	Domain string `json:"domain"`
	Status string `json:"status"`
}

func (a *appJob) GetID() string {
	return a.ID
}
func (a *appJob) GetRev() string {
	return a.Rev
}

func (s *server) statusHandler(w http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) < 10 {
		io.WriteString(w, "appkey is invalid")
		return
	}
	appkey := req.URL.Path[6:]
	job := appJob{}
	err := s.couchdb.Get(&job, appkey)
	if err != nil {
		io.WriteString(w, "could not find app: "+err.Error())
	}
	jsonBytes, _ := json.Marshal(job)
	io.WriteString(w, string(jsonBytes))
}
