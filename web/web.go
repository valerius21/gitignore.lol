package web

import "embed"

//go:embed landing-page/dist/*
var LandingPageFiles embed.FS
