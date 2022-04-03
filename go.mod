module github.com/je4/salon-digital/v2

go 1.17

replace github.com/je4/salon-digital/v2 => ./

replace github.com/je4/PictureFS/v2 => ../PictureFS/

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/dlclark/regexp2 v1.4.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/je4/PictureFS/v2 v2.0.0-20220317073613-af6129bfc191
	github.com/je4/utils/v2 v2.0.6
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/errors v0.9.1
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
)
