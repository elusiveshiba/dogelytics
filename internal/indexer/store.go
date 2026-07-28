package indexer

import "github.com/dogeorg/dogelytics/internal/config"

// Balance represents address balance information.
type Balance struct {
	Incoming  config.DogeAmount `json:"incoming"`
	Available config.DogeAmount `json:"available"`
	Outgoing  config.DogeAmount `json:"outgoing"`
	Current   config.DogeAmount `json:"current"`
}
