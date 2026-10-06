package bossscore

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func post(t *testing.T, srv *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/hit", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func contrib(t *testing.T, srv *Server, player string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/contrib?player="+player, nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("响应不是 JSON：%v", err)
	}
	return out
}

func TestSingleHitCounts(t *testing.T) {
	srv := NewServer()
	post(t, srv, `{"player":"p1","damage":100,"hit_id":"h1","phase":1}`)
	got := contrib(t, srv, "p1")
	if got["total"].(float64) != 100 {
		t.Fatalf("单次 100 伤害应记 100，得到 %+v", got)
	}
}

func TestUnknownPlayerIsZero(t *testing.T) {
	srv := NewServer()
	got := contrib(t, srv, "nobody")
	if got["total"].(float64) != 0 {
		t.Fatalf("没上报过的玩家应记 0，得到 %+v", got)
	}
}

func TestBadBodyIsRejected(t *testing.T) {
	srv := NewServer()
	rec := post(t, srv, `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("坏请求体应回 400，得到 %d", rec.Code)
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("未知路径应回 404，得到 %d", rec.Code)
	}
}

func TestTierThresholds(t *testing.T) {
	if tierOf(100) != "bronze" || tierOf(300) != "silver" || tierOf(600) != "gold" {
		t.Fatal("档位门槛不对")
	}
}
