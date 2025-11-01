package db

import "github.com/kacperborowieckb/gen-sql/utils/env"

func DBConfig() (Config, error) {
	return Config{
		Host:     env.GetString("POSTGRES_HOST", "db"),
		Port:     env.GetString("DB_PORT", "5432"),
		User:     env.GetString("POSTGRES_USER", "postgres"),
		Password: env.GetString("POSTGRES_PASSWORD", "postgres"),
		DBName:   env.GetString("POSTGRES_DB", "gensql"),
		SSLMode:  env.GetString("POSTGRES_SSLMODE", "disable"),
	}, nil
}
