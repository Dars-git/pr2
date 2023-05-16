package main

import (
	"context"
	"flag"
	"log"

	"google.golang.org/grpc"

	"github.com/openai/gpt-3.5-turbo/examples/token-management/token"
)

func main() {
	// Command-line flags
	createFlag := flag.Bool("create", false, "Create a new token")
	dropFlag := flag.Bool("drop", false, "Drop an existing token")
	writeFlag := flag.Bool("write", false, "Write properties to a token")
	readFlag := flag.Bool("read", false, "Read the final value of a token")
	idFlag := flag.String("id", "", "Token ID")
	nameFlag := flag.String("name", "", "Token name")
	lowFlag := flag.Uint64("low", 0, "Low value")
	midFlag := flag.Uint64("mid", 0, "Mid value")
	highFlag := flag.Uint64("high", 0, "High value")
	hostFlag := flag.String("host", "localhost", "Server host")
	portFlag := flag.String("port", "50051", "Server port")

	flag.Parse()

	// Check command-line flags
	if !*createFlag && !*dropFlag && !*writeFlag && !*readFlag {
		log.Fatal("Please specify an operation: -create, -drop, -write, or -read")
	}
	if *createFlag && (*idFlag == "" || *nameFlag == "" || *lowFlag == 0 || *midFlag == 0 || *highFlag == 0) {
		log.Fatal("Please provide all the required parameters for -create")
	}
	if *dropFlag && *idFlag == "" {
		log.Fatal("Please provide the token ID for -drop")
	}
	if *writeFlag && (*idFlag == "" || *nameFlag == "" || *lowFlag == 0 || *midFlag == 0 || *highFlag == 0) {
		log.Fatal("Please provide all the required parameters for -write")
	}
	if *readFlag && *idFlag == "" {
		log.Fatal("Please provide the token ID for -read")
	}

	// Set up gRPC connection
	conn, err := grpc.Dial(*hostFlag+":"+*portFlag, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create a new token client
	client := token.NewTokenServiceClient(conn)

	// Perform the requested operation
	switch {
	case *createFlag:
		createToken(client, *idFlag, *nameFlag, *lowFlag, *midFlag, *highFlag)
	case *dropFlag:
		dropToken(client, *idFlag)
	case *writeFlag:
		writeToken(client, *idFlag, *nameFlag, *lowFlag, *midFlag, *highFlag)
	case *readFlag:
		readToken(client, *idFlag)
	}
}

// createToken creates a new token
func createToken(client token.TokenServiceClient, id, name string, low, mid, high uint64) {
	// Create a new token message
	req := &token.Token{
		Id:   id,
		Name: name,
		Low:  low,
		Mid:  mid,
		High: high,
	}

	// Send the create token request
	res, err := client.CreateToken(context.Background(), req)
	if err != nil {
		log.Fatalf("CreateToken failed: %v", err)
	}

	// Print the response message
	log.Println(res.Message)

	// Print the list of tokens
	printTokens(res.Tokens)
}

// dropToken drops an existing token
func dropToken(client token.TokenServiceClient, id string) {
	// Create a new token message
	req := &token.Token{
		Id: id,
	}

	// Send the drop token request
	res, err := client.DropToken(context.Background(), req)
	if err != nil {
		log.Fatalf("DropToken failed: %v", err)
	}

	// Print the response message
	log.Println(res.Message)

	// Print the list of tokens
	printTokens(res.Tokens)
}

// writeToken writes properties to a token
func writeToken(client token.TokenServiceClient, id, name string, low, mid, high uint64) {
	// Create a new token message
	req := &token.Token{
		Id:   id,
		Name: name,
		Low:  low,
		Mid:  mid,
		High: high,
	}

	// Send the write token request
	res, err := client.WriteToken(context.Background(), req)
	if err != nil {
		log.Fatalf("WriteToken failed: %v", err)
	}

	// Print the response message
	log.Println(res.Message)

	// Print the list of tokens
	printTokens(res.Tokens)
}

// readToken reads the final value of a token
func readToken(client token.TokenServiceClient, id string) {
	// Create a new token message
	req := &token.Token{
		Id: id,
	}

	// Send the read token request
	res, err := client.ReadToken(context.Background(), req)
	if err != nil {
		log.Fatalf("ReadToken failed: %v", err)
	}

	// Print the response message
	log.Println(res.Message)

	// Print the list of tokens
	printTokens(res.Tokens)
}

// printTokens prints the list of tokens
func printTokens(tokens []*token.Token) {
	log.Println("Tokens:")
	for _, t := range tokens {
		log.Printf("- ID: %s, Name: %s, Partial Value: %d, Final Value: %d", t.Id, t.Name, t.PartialValue, t.FinalValue)
	}
}
