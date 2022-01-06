package web

import "embed"

//go:embed template/index.gohtml
var TemplateFS embed.FS

//go:embed static/*
var StaticFS embed.FS

//go:embed inject/*
var InjectFS embed.FS
