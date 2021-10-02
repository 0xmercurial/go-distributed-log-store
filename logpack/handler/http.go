package handler

import (
	"encoding/json"
	logr "logpack/internal/log"
	"net/http"

	"github.com/gorilla/mux"
)

func NewHTTPLogServer(addr string) *http.Server {
	logsrv := newLogServer()
	router := mux.NewRouter()
	router.HandleFunc("/", logsrv.handleAppend).Methods("POST")
	router.HandleFunc("/", logsrv.handleRead).Methods("GET")
	router.HandleFunc("/all", logsrv.handleReadAll).Methods("GET")
	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}

type LogServer struct {
	Log *logr.Log
}

func newLogServer() *LogServer {
	return &LogServer{
		Log: logr.NewLog(),
	}
}

//Record being appended
type AppendRequest struct {
	Record logr.Record `json"record"`
}

//Offset of appended Record
type AppendResponse struct {
	Offset uint64 `json"offset"`
}

//Offset of requested Record
type ReadRequest struct {
	Offset uint64 `json"offset"`
}

//Record at requested Offset
type ReadResponse struct {
	Record logr.Record `json"record"`
}

//All available records
type ReadAllResponse struct {
	Records []logr.Record `json"records"`
}

func (l *LogServer) handleAppend(w http.ResponseWriter, r *http.Request) {
	var req AppendRequest
	err := json.NewDecoder(r.Body).Decode(&req) // Bodies must contain base64 encoded values
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	offset, err := l.Log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := AppendResponse{Offset: offset}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *LogServer) handleRead(w http.ResponseWriter, r *http.Request) {
	var req ReadRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record, err := l.Log.Read(req.Offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ReadResponse{Record: record}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (l *LogServer) handleReadAll(w http.ResponseWriter, r *http.Request) {
	records := l.Log.ReadAll()
	res := ReadAllResponse{Records: records}
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
