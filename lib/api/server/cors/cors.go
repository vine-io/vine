// MIT License
//
// Copyright (c) 2020 Lack
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

package cors

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vine-io/vine/lib/api/handler"
)

// CombinedCORSHandler wraps a server and provides CORS headers
func CombinedCORSHandler() handler.Handler {
	return &corsHandler{}
}

type corsHandler struct {
}

func (c *corsHandler) Handle(ctx *fiber.Ctx) error {
	SetHeaders(ctx)

	if ctx.Method() == "OPTIONS" {
		return nil
	}

	return ctx.Next()
}

func (c corsHandler) String() string {
	return "cors"
}

// SetHeaders sets the CORS headers
func SetHeaders(ctx *fiber.Ctx) error {
	set := func(ctx *fiber.Ctx, k, v string) {
		if v := ctx.Get(k); len(v) > 0 {
			return
		}
		ctx.Set(k, v)
	}

	if origin := ctx.Get("Origin", ""); len(origin) > 0 {
		set(ctx, "Access-Control-Allow-Origin", origin)
	} else {
		set(ctx, "Access-Control-Allow-Origin", "*")
	}

	set(ctx, "Access-Control-Allow-Credentials", "true")
	set(ctx, "Access-Control-Allow-Methods", "POST, PATCH, GET, OPTIONS, PUT, DELETE")
	set(ctx, "Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return nil
}
