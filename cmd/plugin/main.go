// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The SemRels Authors

package main

import (
	generatorplugin "github.com/SemRels/generator-changelog-md/internal/plugin"
	semrelapi "github.com/SemRels/semrel-api/plugin"
	plugin "github.com/hashicorp/go-plugin"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: semrelapi.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"generator": &semrelapi.ChangelogGeneratorGRPCPlugin{
				Impl: generatorplugin.New(),
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
