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

type playerScore struct {
	total     int
	validHits int
}

// Server 是贡献分的 HTTP 处理器。
type Server struct {
	mu     sync.Mutex
	seen   map[string]bool
	scores map[string]*playerScore
}

// NewServer 建一个空的处理器。
func NewServer() *Server {
	return &Server{
		seen:   make(map[string]bool),
		scores: make(map[string]*playerScore),
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
	// 同一个 hit_id 只计一次：重复上报（无论内容）不改变分数。
	if !s.seen[hit.HitID] {
		s.seen[hit.HitID] = true
		if hit.Damage > 0 {
			sc := s.scores[hit.Player]
			if sc == nil {
				sc = &playerScore{}
				s.scores[hit.Player] = sc
			}
			damage := hit.Damage
			if damage > DamageCap {
				damage = DamageCap
			}
			// 阶段累计：分数跨阶段累加，不按阶段清零。
			sc.total += damage
			sc.validHits++
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
	// 查询只需读一条玩家聚合记录，scanned 与上报总条数无关。
	scanned := 1
	total := 0
	if sc := s.scores[player]; sc != nil {
		total = sc.total
		if total > sc.validHits*DamageCap {
			total = sc.validHits * DamageCap
		}
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"player":  player,
		"total":   total,
		"tier":    tierOf(total),
		"scanned": scanned,
	})
}
