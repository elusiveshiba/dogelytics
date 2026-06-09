package indexer

import (
	"github.com/dogeorg/doge/koinu"
)

// Balance represents address balance information.
type Balance struct {
	Incoming  koinu.Koinu
	Available koinu.Koinu
	Outgoing  koinu.Koinu
	Current   koinu.Koinu
}
