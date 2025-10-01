package main

import (
	"context"
	"embed"
	"log"
	"my-ssg/core"
)

//go:embed all:frontend
//go:embed all:templates
var assets embed.FS

type App struct {
	ctx    context.Context
	engine *core.Engine // Our core logic
}

func NewApp() *App {
	// Initialize our core engine
	engine, err := core.NewEngine()
	if err != nil {
		log.Fatalf("Failed to initialize core engine: %v", err)
	}
	return &App{engine: engine}
}
