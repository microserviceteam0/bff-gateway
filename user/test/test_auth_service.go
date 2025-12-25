package main

import (
	"fmt"
	"log"
	"time"
	"user/internal/app/auth_service"

	"user/internal/app/user_service"
	"user/internal/domain/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	dsn := "host=localhost port=5433 user=postgres password=postgres dbname=user_service sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	fmt.Println("✅ Подключение к БД успешно")

	userRepo := repository.NewUserRepository(db)
	userService := user_service.NewUserService(userRepo)

	fmt.Println("\n=== Создание тестового пользователя ===")
	_, err = userService.CreateUser(
		"Test User",
		"test@example.com",
		"password123",
		"admin",
	)

	if err != nil {

		fmt.Println("ℹ️  Пользователь уже существует или ошибка:", err)
	} else {
		fmt.Println("✅ Тестовый пользователь создан")
	}

	fmt.Println("\n=== Инициализация Auth Service ===")
	userAdapter := NewUserAdapter(userService)
	jwtSecret := "test-secret-key"
	tokenDuration := 1 * time.Hour
	authService := auth_service.NewAuthService(userAdapter, jwtSecret, tokenDuration)

	fmt.Println("\n=== Тест 1: Логин с правильными данными ===")
	token, err := authService.Login("test@example.com", "password123")
	if err != nil {
		fmt.Printf("❌ Ошибка логина: %v\n", err)
	} else {
		fmt.Printf("✅ Токен получен: %s...\n", token[:50])

		claims, err := authService.ValidateToken(token)
		if err != nil {
			fmt.Printf("❌ Ошибка валидации токена: %v\n", err)
		} else {
			fmt.Printf("✅ Токен валиден:\n")
			fmt.Printf("   UserID: %d\n", claims.UserID)
			fmt.Printf("   Email: %s\n", claims.Email)
			fmt.Printf("   Role: %s\n", claims.Role)
		}
	}

	fmt.Println("\n=== Тест 2: Логин с неправильным паролем ===")
	_, err = authService.Login("test@example.com", "wrongpassword")
	if err != nil {
		fmt.Printf("✅ Ожидаемая ошибка: %v\n", err)
	} else {
		fmt.Println("❌ Ошибка: токен не должен был быть сгенерирован")
	}

	fmt.Println("\n=== Тест 3: Логин с несуществующим email ===")
	_, err = authService.Login("nonexistent@example.com", "password123")
	if err != nil {
		fmt.Printf("✅ Ожидаемая ошибка: %v\n", err)
	} else {
		fmt.Println("❌ Ошибка: токен не должен был быть сгенерирован")
	}

	fmt.Println("\n=== Тест 4: Refresh токена ===")
	newToken, err := authService.RefreshToken(token)
	if err != nil {
		fmt.Printf("❌ Ошибка refresh: %v\n", err)
	} else {
		fmt.Printf("✅ Новый токен получен: %s...\n", newToken[:50])

		claims, err := authService.ValidateToken(newToken)
		if err != nil {
			fmt.Printf("❌ Ошибка валидации нового токена: %v\n", err)
		} else {
			fmt.Printf("✅ Новый токен валиден для пользователя: %s\n", claims.Email)
		}
	}

	fmt.Println("\n=== Тест 5: Валидация с истекшим токеном ===")

	shortAuthService := auth_service.NewAuthService(userAdapter, jwtSecret, 1*time.Nanosecond)
	shortToken, _ := shortAuthService.Login("test@example.com", "password123")
	time.Sleep(2 * time.Nanosecond)

	_, err = authService.ValidateToken(shortToken)
	if err != nil {
		fmt.Printf("✅ Ожидаемая ошибка для истекшего токена: %v\n", err)
	} else {
		fmt.Println("❌ Ошибка: истекший токен должен быть невалидным")
	}

	fmt.Println("\n=== Все тесты завершены ===")
}

type LocalUserAdapter struct {
	userService user_service.UserService
}

func NewUserAdapter(userService user_service.UserService) auth_service.UserProvider {
	return &LocalUserAdapter{userService: userService}
}

func (a *LocalUserAdapter) ValidateCredentials(email, password string) (*struct {
	ID    int64
	Email string
	Role  string
}, error) {
	user, err := a.userService.ValidateCredentials(email, password)
	if err != nil {
		return nil, err
	}

	return &struct {
		ID    int64
		Email string
		Role  string
	}{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}
