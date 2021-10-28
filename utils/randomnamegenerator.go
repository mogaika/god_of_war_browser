package utils

import (
	"math/rand"

	"github.com/Pallinder/go-randomdata"
)

type RandomNameGenerator map[string]struct{}

func (rng *RandomNameGenerator) RandomName() string {
	if *rng == nil {
		*rng = make(map[string]struct{})
		randomdata.CustomRand(rand.New(rand.NewSource(0)))
	}
	for {
		name := randomdata.SillyName()
		// avoid duplicate names
		if _, exists := (*rng)[name]; !exists {
			(*rng)[name] = struct{}{}
			return name
		}
	}
}
