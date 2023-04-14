package third_party

import (
	"embed"
	"io/fs"
)

//go:embed OpenAPI
var FS embed.FS

func GetStatic() fs.FS {
	sub, err := fs.Sub(FS, "OpenAPI")
	if err != nil {
		panic(err)
	}
	return sub
}

func GetSwagger() []byte {
	b, err := FS.ReadFile("OpenAPI/swagger/index.html")
	if err != nil {
		panic(err)
	}

	return b
}

func GetRedoc() []byte {
	b, err := FS.ReadFile("OpenAPI/redoc/index.html")
	if err != nil {
		panic(err)
	}

	return b
}
