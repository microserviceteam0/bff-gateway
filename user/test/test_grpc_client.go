package main

import (
	"context"
	"fmt"
	"log"
	"time"
	userv1 "user/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)

	fmt.Println("=== Testing gRPC API ===")

	fmt.Println("\n1. Checking if user exists...")
	existsResp, err := client.UserExists(context.Background(), &userv1.UserExistsRequest{UserId: 1})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("User 1 exists: %v\n", existsResp.Exists)
	}

	if existsResp.GetExists() {
		fmt.Println("\n2. Getting user by ID...")
		userResp, err := client.GetUser(context.Background(), &userv1.GetUserRequest{UserId: 1})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			user := userResp.GetUser()
			fmt.Printf("User found: %s (%s)\n", user.GetName(), user.GetEmail())
		}
	}

	fmt.Println("\n3. Validating credentials...")
	validResp, err := client.ValidateCredentials(context.Background(),
		&userv1.ValidateCredentialsRequest{
			Email:    "john@example.com",
			Password: "password123",
		})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Credentials valid: %v\n", validResp.GetValid())
		if validResp.GetValid() {
			user := validResp.GetUser()
			fmt.Printf("Authenticated user: %s\n", user.GetName())
		}
	}

	fmt.Println("\n4. Getting multiple users...")
	usersResp, err := client.GetUsers(context.Background(),
		&userv1.GetUsersRequest{UserIds: []int64{1, 2, 3}})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d users\n", len(usersResp.GetUsers()))
		for _, user := range usersResp.GetUsers() {
			fmt.Printf("  - %s (ID: %d)\n", user.GetName(), user.GetId())
		}
	}

	fmt.Println("\n=== gRPC tests completed ===")
}
