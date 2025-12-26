package main

import (
	"fmt"
	"log"
	"user/internal/app/user_service"
	"user/internal/domain/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=user_service sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewUserRepository(db)
	userService := user_service.NewUserService(repo)

	fmt.Println("=== Тест создания пользователя ===")
	user, err := userService.CreateUser(
		"John Doe",
		"john@example.com",
		"password123",
		"user",
	)

	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Printf("Создан пользователь: ID=%d, Name=%s, Email=%s\n",
			user.ID, user.Name, user.Email)
	}

	fmt.Println("\n=== Тест поиска по email ===")
	foundUser, err := userService.GetUserByEmail("john@example.com")
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Printf("Найден: %s (%s)\n", foundUser.Name, foundUser.Email)
	}
}
