package main

import (
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
)

// mdDefaultSettings returns encoded default classic monopoly_deal settings for
// standing up the parallel classic game (Scenario 4 cross-game isolation).
func mdDefaultSettings() []byte {
	s := monopoly_deal.DefaultSettings()
	buf, err := s.Encode()
	must(err, "encode classic settings")
	return buf
}
