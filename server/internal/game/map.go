package game

import "net/http"

func (s *GameService) GetMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	http.ServeFile(w, r, "../scripts/game_init/world_data/world.txt")
}
