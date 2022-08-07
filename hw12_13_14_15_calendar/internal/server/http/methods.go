package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type Muxer struct {
	app Application
	mux *http.ServeMux
}

type eventRequest struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Start          string `json:"start"`
	End            string `json:"end"`
	OwnerID        int    `json:"ownerID"` //nolint:tagliatelle
	RemindBefore   string `json:"remindBefore"`
	RemindSent     bool   `json:"remindSent"`
	RemindReceived bool   `json:"remindReceived"`
	tStart         time.Time
	tEnd           time.Time
	remind         time.Duration
}

type searchRequest struct {
	ID    string `json:"id"`
	Day   string `json:"day"`
	Week  string `json:"week"`
	Month string `json:"month"`
}

var (
	errInvalidJSON           = errors.New("failed to parse request body")
	errInvalidDateTimeFormat = errors.New("failed to parse date time: expecting \"YYYY-MM-DD HH:MM\" format")
	errInvalidDateFormat     = errors.New("failed to parse date: expecting \"YYYY-MM-DD\" format")
	errInvalidDuration       = errors.New("failed to parse duration: expecting \"_h_m_s\" format")
	errMarshalFailed         = errors.New("failed to marshal event for response")
	errReaderFail            = errors.New("failed to read request contents")
	errDataMissing           = errors.New("some required fields are not filled")
)

func NewMux(app Application) *Muxer {
	muxer := Muxer{
		app: app,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", muxer.hello)
	mux.HandleFunc("/event", muxer.event)
	mux.HandleFunc("/events", muxer.list)

	muxer.mux = mux
	return &muxer
}

func (m *Muxer) hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}

func (m *Muxer) list(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	interval := r.URL.Query().Get("interval")
	if date == "" || interval == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(errDataMissing.Error()))
	}
	tDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(errInvalidDateFormat.Error()))
	}
	events, err := m.app.GetList(tDate, interval)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	response, err := json.Marshal(events)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMarshalFailed.Error()))
	}
	w.Write(response)
}

func (m *Muxer) event(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		m.getByID(w, r)
	case http.MethodPost:
		m.create(w, r)
	case http.MethodPut:
		m.update(w, r)
	case http.MethodDelete:
		m.delete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *Muxer) create(w http.ResponseWriter, r *http.Request) {
	request, err := parseEventRequest(r, false)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	id, err := m.app.CreateEvent(
		context.Background(),
		request.Title,
		request.Description,
		request.tStart,
		request.tEnd,
		request.OwnerID,
		request.remind,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(id))
	}
}

func (m *Muxer) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("param \"id\" required"))
	}

	event, err := m.app.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		eventResponse, err := json.Marshal(event)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errMarshalFailed.Error()))
		}
		w.Write(eventResponse)
	}
}

func parseEventRequest(r *http.Request, isUpdate bool) (eventRequest, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return eventRequest{}, errReaderFail
	}
	request := eventRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		return eventRequest{}, errInvalidJSON
	}
	if !isEventRequestValid(request, isUpdate) {
		return eventRequest{}, errDataMissing
	}
	start, err := time.Parse("2006-01-02 15:04", request.Start)
	if err != nil {
		return eventRequest{}, errInvalidDateTimeFormat
	}
	end, err := time.Parse("2006-01-02 15:04", request.End)
	if err != nil {
		return eventRequest{}, errInvalidDateTimeFormat
	}
	remind, err := time.ParseDuration(request.RemindBefore)
	if err != nil {
		return eventRequest{}, errInvalidDuration
	}
	request.tStart = start
	request.tEnd = end
	request.remind = remind
	return request, nil
}

func isEventRequestValid(request eventRequest, isUpdate bool) bool {
	isValid := request.Title != "" && request.Start != "" && request.End != ""
	if isUpdate {
		// для апдейта обязателен ID
		isValid = isValid && request.ID != ""
	} else {
		// для создания обязателен хозяин
		isValid = isValid && request.OwnerID != 0
	}
	return isValid
}

func (m *Muxer) delete(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	request := searchRequest{}
	err = json.Unmarshal(data, &request)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(errInvalidJSON.Error()))
		return
	}

	if request.ID == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("param \"id\" required"))
	}

	err = m.app.DeleteEvent(request.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusGone)
	}
}

func (m *Muxer) update(w http.ResponseWriter, r *http.Request) {
	request, err := parseEventRequest(r, true)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}
	err = m.app.UpdateEvent(
		context.Background(),
		request.ID,
		request.Title,
		request.Description,
		request.tStart,
		request.tEnd,
		request.remind,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
