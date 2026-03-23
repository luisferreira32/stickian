package game

import (
	"errors"
	"net/http"
)

const worldSize = 256

type MapChunkResponse struct {
	Biome [][]int `json:"biome"`
}

type MapChunkRequest struct {
	MinQ int `json:"minQ"`
	MaxQ int `json:"maxQ"`
	MinR int `json:"minR"`
	MaxR int `json:"maxR"`
}

func validateMapChunkRequest(req *MapChunkRequest) error {
	if req.MinQ < 0 || req.MaxQ > worldSize || req.MinR < 0 || req.MaxR > worldSize {
		return errors.New("invalid map chunk request")
	}
	return nil
}

func (s *GameService) GetMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	w.Write(MapChunk())
}
