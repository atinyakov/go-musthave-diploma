package config

import (
	"flag"
	"log"
	"os"

	"github.com/gookit/slog"

	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

// AppConfig - глобальная переменная для хранения конфигурации
var AppConfig *Config

// LoadConfig загружает конфигурацию из флагов и переменных окружения
func LoadConfig() *Config {
	// Загружаем .env, если он есть
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные.")
	}

	// Определяем флаги командной строки
	runAddress := flag.String("a", getEnv("RUN_ADDRESS", "localhost:3000"), "Адрес запуска сервиса")
	databaseURI := flag.String("d", getEnv("DATABASE_URI", "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable"), "URI базы данных")
	accrualSystemAddress := flag.String("r", getEnv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8080"), "Адрес системы расчёта начислений")

	// Разбираем флаги
	flag.Parse()

	// Сохраняем настройки в глобальную переменную
	AppConfig = &Config{
		RunAddress:           *runAddress,
		DatabaseURI:          *databaseURI,
		AccrualSystemAddress: *accrualSystemAddress,
	}

	slog.Info("config loaded: %+v\n", AppConfig)

	return AppConfig
}

// getEnv - вспомогательная функция для загрузки переменных окружения с дефолтными значениями
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
