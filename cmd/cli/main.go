package main

import (
	"fmt"

	"github.com/newbpydev/tusk/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Println("DB_URL:", cfg.DBURL)
	fmt.Println("PORT:", cfg.Port)
	fmt.Println("APP_ENV:", cfg.AppEnv)
}
