package httpapi

import "net/http"

func (s *server) systemStatus(w http.ResponseWriter, r *http.Request) {
	if s.config.SystemStatus == nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "system_metrics_unavailable", "message": "System metrics are unavailable."})
		return
	}
	snapshot, err := s.config.SystemStatus(r)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, envelope{"code": "system_metrics_unavailable", "message": "System metrics are unavailable."})
		return
	}
	writeJSON(w, http.StatusOK, envelope{
		"metrics":           snapshot,
		"traffic_semantics": "server_counter_delta",
	})
}
