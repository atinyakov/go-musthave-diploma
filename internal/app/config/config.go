package config

import (
	"cmp"
	"flag"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

// LoadConfig загружает конфигурацию из флагов и переменных окружения
func LoadConfig() *Config {
	// Загружаем .env, если он есть
	if err := godotenv.Load(); err != nil {
		slog.Info("Файл .env не найден, используются системные переменные.")
	}

	// Определяем флаги командной строки
	runAddress := flag.String("a", cmp.Or(os.Getenv("RUN_ADDRESS"), "localhost:3000"), "Адрес запуска сервиса")
	databaseURI := flag.String("d", cmp.Or(os.Getenv("DATABASE_URI"), "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable"), "URI базы данных")
	accrualSystemAddress := flag.String("r", cmp.Or(os.Getenv("ACCRUAL_SYSTEM_ADDRESS"), "http://localhost:8080"), "Адрес системы расчёта начислений")

	// Разбираем флаги
	flag.Parse()

	// Сохраняем настройки в глобальную переменную
	AppConfig := &Config{
		RunAddress:           *runAddress,
		DatabaseURI:          *databaseURI,
		AccrualSystemAddress: *accrualSystemAddress,
	}

	slog.Info("config loaded: %+v\n", slog.Any("config", AppConfig))

	return AppConfig
}
