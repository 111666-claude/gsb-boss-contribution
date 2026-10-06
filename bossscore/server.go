// Package bossscore 统计世界 Boss 的玩家贡献分。
package bossscore

import (
	"encoding/json"
	"net/http"
	"sync"
)

// DamageCap 是单次伤害计入的上限。
const DamageCap = 500

// Hit 是一次伤害上报。
type Hit struct {
	Player string `json:"player"`
	Damage int    `json:"damage"`
	HitID  string `json:"hit_id"`
	Phase  int    `json:"phase"`
}

type tierDef struct {
	Min  int
	Name string
}

var tiers = []tierDef{{100, "bronze"}, {300, "silver"}, {600, "gold"}}

// Server 是贡献分的 HTTP 处理器。
type Server struct {
	mu     sync.Mutex
	seen   map[string]struct{}
	totals map[string]int
}

// NewServer 建一个空的处理器。
func NewServer() *Server {
	return &Server{
		seen:   make(map[string]struct{}),
		totals: make(map[string]int),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/hit":
		s.handleHit(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/contrib":
		s.handleContrib(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleHit(w http.ResponseWriter, r *http.Request) {
	var hit Hit
	if err := json.NewDecoder(r.Body).Decode(&hit); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	if hit.Damage > 0 {
		if _, dup := s.seen[hit.HitID]; !dup {
			s.seen[hit.HitID] = struct{}{}
			damage := hit.Damage
			if damage > DamageCap {
				damage = DamageCap
			}
			s.totals[hit.Player] += damage
		}
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func tierOf(total int) string {
	name := ""
	for _, t := range tiers {
		if total >= t.Min {
			name = t.Name
		}
	}
	return name
}

func (s *Server) handleContrib(w http.ResponseWriter, r *http.Request) {
	player := r.URL.Query().Get("player")
	s.mu.Lock()
	total := s.totals[player]
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"player":  player,
		"total":   total,
		"tier":    tierOf(total),
		"scanned": 1,
	})
}
