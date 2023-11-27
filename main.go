package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

var mutex sync.RWMutex
var keys map[string]bool

func setKeys(values []string) {
	mutex.Lock()
	defer mutex.Unlock()
	clear(keys)
	for _, v := range values {
		keys[v] = true
	}
}

func existKey(key string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return keys[key]
}

func load() error {
	file, err := os.OpenFile(".apikeys", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println("Error al abrir el archivo .apikeys", err)
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		log.Println("Error al leer el archivo .apikeys:", err)
		return err
	}
	setKeys(lines)
	return nil
}

func main() {
	if err := load(); err != nil {
		os.Exit(1)
	}
	e := echo.New()
	e.HideBanner = true
	go func() {
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	e.GET("/", func(c echo.Context) error {
		var apiKey = c.Request().Header.Get("X-Original-Apikey")
		if existKey(apiKey) {
			return c.NoContent(http.StatusOK)
		}
		return c.NoContent(http.StatusUnauthorized)
	})
	e.GET("/refresh", func(c echo.Context) error {
		if err := load(); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusOK, "success!!")
	})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
