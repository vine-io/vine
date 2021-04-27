// MIT License
//
// Copyright (c) 2021 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package template

var (
	TOML = `[package]
kind = "{{.Toml.Package.Kind}}"
namespace = "{{.Toml.Package.Namespace}}"
root = "{{.Toml.Package.Root}}"
{{if .Toml.Mod}}{{range .Toml.Mod}}
[[mod]]
name = "{{.Name}}"
alias = "{{.Alias}}"
type = "{{.Type}}"
version = "{{.Version}}"
main = "{{.Main}}"
dir = "{{.Dir}}"
output = "{{.Output}}"
flags = [
	{{range .Flags}}"{{quota .}}", {{end}}
]{{end}}{{end}}
{{if .Toml.Pkg}}[pkg]
name = "{{.Toml.Pkg.Name}}"
alias = "{{.Toml.Pkg.Alias}}"
type = "{{.Toml.Pkg.Type}}"
version = "{{.Toml.Pkg.Version}}"
main = "{{.Toml.Pkg.Main}}"
dir = "{{.Toml.Pkg.Dir}}"
output = "{{.Toml.Pkg.Output}}"
flags = [
	{{range .Toml.Pkg.Flags}}"{{quota .}}", {{end}}
]{{end}}
{{range .Toml.Proto}}
[[proto]]
name = "{{.Name}}"
pb = "{{.Pb}}"
type = "{{.Type}}"
version = "{{.Version}}"
plugins = ["gogo"{{range .Plugins}}{{if ne . "gogo"}}, "{{.}}"{{end}}{{end}}]
{{end}}
`
)
