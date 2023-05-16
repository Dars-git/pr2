package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"pr2/token"
	"syscall"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// TokenServer represents the gRPC server
type TokenServer struct {
	tokens map[string]*token.Token
}

// CreateToken creates a new token
func (s *TokenServer) CreateToken(ctx context.Context, req *token.Token) (*token.TokenResponse, error) {
	if _, exists := s.tokens[req.Id]; exists {
		return &token.TokenResponse{
			Message: fmt.Sprintf("Token with ID %s already exists", req.Id),
			Tokens:  s.getTokens(),
		}, nil
	}

	s.tokens[req.Id] = req

	return &token.TokenResponse{
		Message: fmt.Sprintf("Token with ID %s created successfully", req.Id),
		Tokens:  s.getTokens(),
	}, nil
}

// DropToken drops a token
func (s *TokenServer) DropToken(ctx context.Context, req *token.Token) (*token.TokenResponse, error) {
	if _, exists := s.tokens[req.Id]; !exists {
		return &token.TokenResponse{
			Message: fmt.Sprintf("Token with ID %s does not exist", req.Id),
			Tokens:  s.getTokens(),
		}, nil
	}

	delete(s.tokens, req.Id)

	return &token.TokenResponse{
		Message: fmt.Sprintf("Token with ID %s dropped successfully", req.Id),
		Tokens:  s.getTokens(),
	}, nil
}

// WriteToken writes properties to a token
func (s *TokenServer) WriteToken(ctx context.Context, req *token.Token) (*token.TokenResponse, error) {
	t, exists := s.tokens[req.Id]
	if !exists {
		return &token.TokenResponse{
			Message: fmt.Sprintf("Token with ID %s does not exist", req.Id),
			Tokens:  s.getTokens(),
		}, nil
	}

	t.Name = req.Name
	t.Low = req.Low
	t.Mid = req.Mid
	t.High = req.High

	t.PartialValue = computePartialValue(t.Name, t.Low, t.Mid)
	t.FinalValue = 0

	return &token.TokenResponse{
		Message: fmt.Sprintf("Properties updated for token with ID %s", req.Id),
		Tokens:  s.getTokens(),
	}, nil
}

// ReadToken reads the final value of a token
func (s *TokenServer) ReadToken(ctx context.Context, req *token.Token) (*token.TokenResponse, error) {
	t, exists := s.tokens[req.Id]
	if !exists {
		return &token.TokenResponse{
			Message: fmt.Sprintf("Token with ID %s does not exist", req.Id),
			Tokens:  s.getTokens(),
		}, nil
	}

	finalValue := computeFinalValue(t.Name, t.Mid, t.High)
	t.FinalValue = min(finalValue, t.PartialValue)

	return &token.TokenResponse{
		Message: fmt.Sprintf("Final value read for token with ID %s", req.Id),
		Tokens:  s.getTokens(),
	}, nil
}

// getTokens returns a list of all tokens
func (s *TokenServer) getTokens() []*token.Token {
	tokens := make([]*token.Token, 0, len(s.tokens))
	for _, t := range s.tokens {
		tokens = append(tokens, t)
	}
	return tokens
}

// computePartialValue computes the partial value of a token
func computePartialValue(name string, low, mid uint64) uint64 {
	var partialValue uint64
	minHash := ^uint64(0)

	for x := low; x < mid; x++ {
		hash := Hash(name, x)
		if hash < minHash {
			minHash = hash
			partialValue = x
		}
	}

	return partialValue
}

// computeFinalValue computes the final value of a token
func computeFinalValue(name string, mid, high uint64) uint64 {
	var finalValue uint64
	minHash := ^uint64(0)

	for x := mid; x < high; x++ {
		hash := Hash(name, x)
		if hash < minHash {
			minHash = hash
			finalValue = x
		}
	}

	return finalValue
}

// Hash concatenates a message and a nonce and generates a hash value.
func Hash(name string, nonce uint64) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s %d", name, nonce)))
	return binary.BigEndian.Uint64(hasher.Sum(nil))
}

// min returns the minimum of two uint64 values
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Create a gRPC server
	server := grpc.NewServer()

	// Initialize the token server
	tokenServer := &TokenServer{
		tokens: make(map[string]*token.Token),
	}

	// Register the token service with the server
	token.RegisterTokenServiceServer(server, tokenServer)

	// Start the server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Server listening on port %s", port)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for termination signal (CTRL-C)
	waitForTerminationSignal()

	// Stop the server gracefully
	server.GracefulStop()

	log.Println("Server stopped")
}

// waitForTerminationSignal waits for a termination signal (CTRL-C)
func waitForTerminationSignal() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	<-signalCh
}
