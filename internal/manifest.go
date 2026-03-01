package sish

import (
	_ "embed"
	"encoding/json"
)

//go:embed package.json
var manifestData []byte

type Manifest struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	ShortDescription string `json:"shortDescription"`
}

func GetManifest() (*Manifest, error) {
	var manifest Manifest

	err := json.Unmarshal(manifestData, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}
