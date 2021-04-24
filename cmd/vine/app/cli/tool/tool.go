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

package tool

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Package Package `json:"package" toml:"package"`
	Mod     *Mods   `json:"mod" toml:"mod"`
	Pkg     *Mod    `json:"pkg" toml:"pkg"`
	Proto   Protos  `json:"proto" toml:"proto"`
}

type Package struct {
	Kind      string `json:"kind" toml:"kind"`
	Namespace string `json:"namespace" toml:"namespace"`
}

type Mods []Mod

type Mod struct {
	Name    string   `json:"name" toml:"name"`
	Alias   string   `json:"alias" toml:"alias"`
	Type    string   `json:"type" toml:"type"`
	Version string   `json:"version" toml:"version"`
	Dir     string   `json:"dir" toml:"dir"`
	Output  string   `json:"output" toml:"output"`
	Flags   []string `json:"flags" toml:"flags"`
}

type Protos []Proto

type Proto struct {
	Name    string   `json:"name" toml:"name"`
	Pb      string   `json:"pb" toml:"pb"`
	Type    string   `json:"type" toml:"type"`
	Plugins []string `json:"plugins" toml:"plugins"`
}

func New(f string) (*Config, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var c Config
	_, err = toml.Decode(string(b), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
