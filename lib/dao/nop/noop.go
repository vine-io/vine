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

package nop

import (
	"github.com/vine-io/vine/lib/dao"
	"github.com/vine-io/vine/lib/dao/clause"
	"github.com/vine-io/vine/lib/dao/schema"
)

type noopDialect struct {
	opts dao.Options
}

func (d noopDialect) Init(opts ...dao.Option) error {
	d.opts = dao.Options{}
	for _, opt := range opts {
		opt(&d.opts)
	}
	return nil
}

func (d noopDialect) Options() dao.Options {
	return d.opts
}

func (d noopDialect) NewTx() *dao.DB {
	return &dao.DB{}
}

func (noopDialect) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

func (noopDialect) Migrator() dao.Migrator {
	return &Migrator{}
}

func (noopDialect) BindVarTo(writer clause.Writer, _ *dao.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (noopDialect) QuoteTo(writer clause.Writer, str string) {
	writer.WriteByte('`')
	writer.WriteString(str)
	writer.WriteByte('`')
}

func (noopDialect) Explain(string, ...interface{}) string {
	return ""
}

func (noopDialect) DataTypeOf(*schema.Field) string {
	return ""
}

func (d noopDialect) JSONDataType() string {
	return "JSON"
}

func (d noopDialect) JSONBuild(string) dao.JSONQuery {
	return jsonQueryExpression{}
}

func (noopDialect) String() string {
	return "noop"
}

type jsonQueryExpression struct {
}

func (j jsonQueryExpression) Tx(*dao.DB) dao.JSONQuery {
	return j
}

func (j jsonQueryExpression) Op(dao.JSONOp, interface{}, ...string) dao.JSONQuery {
	return j
}

func (j jsonQueryExpression) Contains(op dao.JSONOp, values interface{}, keys ...string) dao.JSONQuery {
	return j
}

func (j jsonQueryExpression) Build(clause.Builder) {
	return
}

func NewDialect(opts ...dao.Option) dao.Dialect {
	options := dao.NewOptions(opts...)
	return &noopDialect{opts: options}
}
