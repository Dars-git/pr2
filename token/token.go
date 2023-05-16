package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/openai/gpt-3.5-turbo/examples/token-management/token"
)

type tokenServer struct {
	mu     sync.Mutex
	tokens map[string]*token.Token
}

// CreateToken creates a new token with the given ID.
func (s *tokenServer) CreateToken(ctx context.Context, req *token.Token) (*token.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tokens[req.Id]; exists {
		return &token.Response{
			Message: "Token with the same ID already exists",
			Tokens:  s.getTokensList(),
		}, nil
	}

	s.tokens[req.Id] = &token.Token{
		Id:           req.Id,
		Name:         req.Name,
		Domain:       req.Domain,
		State:        &token.TokenState{},
		PartialValue: 0,
		FinalValue:   0,
	}

	return &token.Response{
		Message: "Token created successfully",
		Tokens:  s.getTokensList(),
	}, nil
}

// DropToken drops the token with the given ID.
func (s *tokenServer) DropToken(ctx context.Context, req *token.Token) (*token.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tokens[req.Id]; !exists {
		return &token.Response{
			Message: "Token not found",
			Tokens:  s.getTokensList(),
		}, nil
	}

	delete(s.tokens, req.Id)

	return &token.Response{
		Message: "Token dropped successfully",
		Tokens:  s.getTokensList(),
	}, nil
}

// WriteToken writes properties to the token with the given ID.
func (s *tokenServer) WriteToken(ctx context.Context, req *token.Token) (*token.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tokens[req.Id]
	if !exists {
		return &token.Response{
			Message: "Token not found",
			Tokens:  s.getTokensList(),
		}, nil
	}

	t.Name = req.Name
	t.Domain = req.Domain
	t.State = &token.TokenState{}
	t.PartialValue = s.computePartialValue(t.Name, t.Domain.Low, t.Domain.Mid)

	return &token.Response{
		Message: "Token properties updated successfully",
		Tokens:  s.getTokensList(),
	}, nil
}

// ReadToken reads the final value of the token with the given ID.
func (s *tokenServer) ReadToken(ctx context.Context, req *token.Token) (*token.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tokens[req.Id]
	if !exists {
		return &token.Response{
			Message: "Token not found",
			Tokens:  s.getTokensList(),
		}, nil
	}

	finalValue := s.computeFinalValue(t.Name, t.Domain.Mid, t.Domain.High)
	t.FinalValue = finalValue

	return &token.Response{
		Message:    "Final value retrieved successfully",
		Tokens:     s.getTokensList(),
		FinalValue: finalValue,
	}, nil
}

// computePartialValue computes the partial value of the token using the given name and range.
func (s *tokenServer) computePartialValue(name string, low, high uint64) uint64 {
	var minHash uint64 = ^uint64(0) // maximum possible value

	for i := low; i < high; i++ {
		hash := Hash(name, i)
		if hash < minHash {
			minHash = hash
		}
	}

	return minHash
}

// computeFinalValue computes the final value of the token using the given name and range.
func (s *tokenServer) computeFinalValue(name string, low, high uint64) uint64 {
	var minHash uint64 = ^uint64(0) // maximum possible value

	for i := low; i < high; i++ {
		hash := Hash(name, i)
		if hash < minHash {
			minHash = hash
		}
	}

	return minHash
}

// getTokensList returns the list of tokens.
func (s *tokenServer) getTokensList() []*token.Token {
	tokens := make([]*token.Token, 0, len(s.tokens))

	for _, t := range s.tokens {
		tokens = append(tokens, t)
	}

	return tokens
}

func main() {
	// Parse command line arguments
	port := flag.Int("port", 50051, "the server port")
	flag.Parse()

	// Create a new token server
	server := &tokenServer{
		tokens: make(map[string]*token.Token),
	}

	// Create a gRPC server
	grpcServer := grpc.NewServer()

	// Register the token server with the gRPC server
	token.RegisterTokenServiceServer(grpcServer, server)

	// Start the gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Server listening on port %d", *port)

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
