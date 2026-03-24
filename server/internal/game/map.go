package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

const worldSize = 256

type MapChunkResponse struct {
	Biome [][]int `json:"biome"`
}

type MapChunkRequest struct {
	MinQ int `json:"MinQ"`
	MaxQ int `json:"MaxQ"`
	MinR int `json:"MinR"`
	MaxR int `json:"MaxR"`
}

func validateMapChunkRequest(req *MapChunkRequest) error {
	if req.MinQ < 0 || req.MaxQ > worldSize || req.MinR < 0 || req.MaxR > worldSize {
		return errors.New("invalid map chunk request")
	}
	return nil
}

func (s *GameService) GetMap(w http.ResponseWriter, r *http.Request) {
	req := MapChunkRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}
	defer r.Body.Close()

	if err := validateMapChunkRequest(&req); err != nil {
		utils.WithError(w, err)
		return
	}

	tiles, err := s.Database.GetMap(req.MinQ, req.MaxQ, req.MinR, req.MaxR)
	if err != nil {
		utils.WithError(w, fmt.Errorf("failed to fetch map: %w", err))
		return
	}

	// Transform tiles into the 2D biome array expected by MapChunkResponse
	// We need to know the dimensions to create the array
	width := req.MaxQ - req.MinQ + 1
	height := req.MaxR - req.MinR + 1
	biome := make([][]int, width)
	for i := range biome {
		biome[i] = make([]int, height)
	}

	for _, t := range tiles {
		qIdx := t.Q - req.MinQ
		rIdx := t.R - req.MinR
		if qIdx >= 0 && qIdx < width && rIdx >= 0 && rIdx < height {
			biome[qIdx][rIdx] = t.Biome
		}
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(MapChunkResponse{Biome: biome}); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode map: %w", err))
		return
	}
}
