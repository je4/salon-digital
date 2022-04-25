package web

import "embed"

//go:embed static/*
var StaticFS embed.FS

//go:embed inject/*
var InjectFS embed.FS
