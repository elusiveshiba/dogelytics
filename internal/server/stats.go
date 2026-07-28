package server

import (
	"log"
	"net/http"
)

// HandleUsageStats returns usage statistics for a given timeframe.
func (s *Server) HandleUsageStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hours, filterType, filterValues := s.usageQuery(r)
	stats, err := s.authStore.GetUsageStats(r.Context(), hours, filterType, filterValues)
	if err != nil {
		log.Printf("[Dogelytics] Error getting usage stats: %v", err)
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	s.sendJSON(w, http.StatusOK, stats)
}

// HandleUsageTimeSeries returns time-series data for charts.
func (s *Server) HandleUsageTimeSeries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hours, filterType, filterValues := s.usageQuery(r)
	series, err := s.authStore.GetUsageTimeSeries(r.Context(), hours, filterType, filterValues)
	if err != nil {
		log.Printf("[Dogelytics] Error getting usage time series: %v", err)
		http.Error(w, "failed to get time series", http.StatusInternalServerError)
		return
	}

	s.sendJSON(w, http.StatusOK, series)
}

func (s *Server) usageQuery(r *http.Request) (int, string, []string) {
	hours := hoursForTimeframe(r.URL.Query().Get("timeframe"))
	filterType := r.URL.Query().Get("filter")

	var filterValues []string
	if filterType == "keys" {
		if user, ok := s.getUserFromRequest(r); ok {
			keys, err := s.authStore.GetAPIKeysByUserID(r.Context(), user.ID)
			if err == nil {
				for _, key := range keys {
					filterValues = append(filterValues, key.KID)
				}
			}
		}
	}

	return hours, filterType, filterValues
}

func hoursForTimeframe(timeframe string) int {
	switch timeframe {
	case "hour":
		return 1
	case "day":
		return 24
	case "week":
		return 168
	case "month":
		return 720
	case "year":
		return 8760
	default:
		return 24
	}
}
