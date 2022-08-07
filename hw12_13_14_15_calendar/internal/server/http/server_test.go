package internalhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	memorystorage "github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestHttpServerHelloWorld(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	httpHandler := newHandler()
	httpHandler.mux.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, "Hello world", string(body))
	resp.Body.Close()
}

func TestHttpServerEventsOperations(t *testing.T) {
	body := bytes.NewBufferString(`{
        "title":"test",
        "description":"test",
        "start":"2021-08-21 21:30",
        "end":"2021-08-21 21:35",
        "ownerID":3,
        "remindBefore":"24h"
	}`)
	req := httptest.NewRequest("POST", "/event", body)
	w := httptest.NewRecorder()

	httpHandler := newHandler()
	httpHandler.mux.ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	require.Equal(t, 36, len(respBody))
	uid := string(respBody)
	resp.Body.Close()

	body = bytes.NewBufferString(`{
		"id":"` + uid + `",
		"title":"test upd",
        "description":"test upd",
        "start":"2022-08-21 21:30",
        "end":"2022-08-21 21:35",
        "remindBefore":"12h"
	}`)
	req = httptest.NewRequest("PUT", "/event", body)
	w = httptest.NewRecorder()

	httpHandler.mux.ServeHTTP(w, req)

	resp = w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	body = bytes.NewBufferString(`{
        "ID":"` + uid + `"
	}`)
	req = httptest.NewRequest("DELETE", "/event", body)
	w = httptest.NewRecorder()

	httpHandler.mux.ServeHTTP(w, req)

	resp = w.Result()
	require.Equal(t, http.StatusGone, resp.StatusCode)
	resp.Body.Close()
}

func newHandler() *Muxer {
	logg := logger.NewLogger("debug", "STDERR")
	storage := memorystorage.New(*logg)
	calendar := app.New(logg, storage)
	return NewMux(calendar)
}
