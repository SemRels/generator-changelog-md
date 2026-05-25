// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-semrel Authors

// Package grpc provides the gRPC transport adapter for the changelog generator plugin.
package grpc

import (
	"context"

	semrelv1 "github.com/SemRels/generator-changelog-md/internal/gen/v1"
	"github.com/SemRels/generator-changelog-md/internal/plugin"
)

// ChangelogServer implements semrelv1.ChangelogGeneratorPluginServer by delegating
// to the pure-Go generator in internal/plugin.
type ChangelogServer struct {
	semrelv1.UnimplementedChangelogGeneratorPluginServer
}

// NewChangelogServer constructs a ChangelogServer ready to be registered with a gRPC server.
func NewChangelogServer() *ChangelogServer {
	return &ChangelogServer{}
}

// GenerateNotes renders a Markdown changelog fragment for the release described by req.Ctx.
func (s *ChangelogServer) GenerateNotes(
	_ context.Context,
	req *semrelv1.GenerateNotesRequest,
) (*semrelv1.GenerateNotesResponse, error) {
	notes := plugin.FormatMarkdown(req.GetCtx())
	return &semrelv1.GenerateNotesResponse{Notes: notes}, nil
}
