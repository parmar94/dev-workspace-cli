// Package assets holds the built-in registry and templates shipped with dws.
package assets

import "embed"

//go:embed registry.yaml
var Registry []byte

//go:embed templates/*.yaml
var Templates embed.FS

// Skills holds agent skills (skills/<name>/SKILL.md plus supporting files)
// installed by `dws skills install`.
//
//go:embed skills
var Skills embed.FS

// DefaultConfig is the starter global config written by `dws config init`.
//
//go:embed config.yaml
var DefaultConfig []byte
