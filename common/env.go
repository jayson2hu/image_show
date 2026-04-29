package common

import "github.com/joho/godotenv"

func LoadEnv() {
	_ = godotenv.Load()
}
