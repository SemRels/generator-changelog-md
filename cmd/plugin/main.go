// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-semrel Authors

// Binary plugin starts the markdown changelog generator as a gRPC server.
//
// The server address is controlled by the SEMREL_PLUGIN_ADDR environment
// variable (default ":50051").  All log output is written to stderr; stdout
// is reserved for the go-plugin handshake per ADR-001.
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	semrelv1 "github.com/SemRels/generator-changelog-md/internal/gen/v1"
	grpcserver "github.com/SemRels/generator-changelog-md/internal/grpc"
)

func main() {
	logger := log.New(os.Stderr, "[generator-changelog-md] ", log.LstdFlags)

	addr := os.Getenv("SEMREL_PLUGIN_ADDR")
	if addr == "" {
		addr = ":50051"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("failed to listen on %s: %v", addr, err)
	}

	srv := grpc.NewServer()
	semrelv1.RegisterChangelogGeneratorPluginServer(srv, grpcserver.NewChangelogServer())

	logger.Printf("gRPC server listening on %s", lis.Addr())

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Println("shutting down gRPC server…")
		srv.GracefulStop()
	}()

	if err := srv.Serve(lis); err != nil {
		logger.Fatalf("gRPC server exited with error: %v", err)
	}
}
