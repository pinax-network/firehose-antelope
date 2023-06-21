package flags

import "golang.org/x/exp/slices"

type Feature string

// available build features

const FeatureUltra Feature = "ultra"

var features = []Feature{}

func GetEnabledFeatures() []Feature {
	return features
}

func EnableFeature(feature Feature) {
	if !slices.Contains(features, feature) {
		features = append(features, feature)
	}
}

func HasEnabledFeatures() bool {
	return len(features) > 0
}

func HasFeatureEnabled(feature Feature) bool {
	return slices.Contains(features, feature)
}
