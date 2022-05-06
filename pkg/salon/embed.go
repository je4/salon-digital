package salon

import "embed"

//go:embed embed/template/index.gohtml
//go:embed embed/template/grid.gohtml
var TemplateFS embed.FS

//go:embed embed/static/*
var StaticFS embed.FS

//go:embed embed/pfsEmbed/salon-digital.json
var SalonDigitalJSON []byte

//go:embed embed/pfsEmbed/salon-digital.png
var SalonDigitalImage []byte
