// Copyright 2020 The vine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pg

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	json "github.com/json-iterator/go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lack-io/vine/internal/db"
	"github.com/lack-io/vine/internal/db/selection"
	"github.com/lack-io/vine/internal/db/watch"
	metav1 "github.com/lack-io/vine/internal/meta/v1"
	"github.com/lack-io/vine/internal/runtime"
)

type PostgresSQL struct {
	cfg *Config

	db *gorm.DB

	aspect *APIObjectAspect
}

// New create &PostgresSQL by *Config
func New(cfg *Config) (*PostgresSQL, error) {
	mode := "disable"
	if cfg.SSL {
		mode = "enable"
	}

	dsn := fmt.Sprintf("host=%s port=%d, user=%s, password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		mode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetConnMaxIdleTime(time.Second * time.Duration(cfg.ConnMaxIdleTime))
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(cfg.ConnMaxLifetime))
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	pg := &PostgresSQL{
		cfg:    cfg,
		db:     db,
		aspect: &APIObjectAspect{},
	}

	return pg, nil
}

func (p *PostgresSQL) Aspect() db.Aspect {
	return p.aspect
}

// Get implements interface db.DB. the parameter 'key' should be name of postgresSQL table. The default value
// is 'First' If the Clause field of selection.Predicate does not contain 'First' or 'Last'.
func (p *PostgresSQL) Get(ctx context.Context, key string, pre selection.Predicate, obj runtime.Object, ignoreNotFound bool) (err error) {
	p.aspect.BeforeRequest(ctx, obj)
	defer p.aspect.AfterRequest(ctx, obj)

	tx := p.db.WithContext(ctx).Table(key)

	if pre.Fields != nil {
		tx.Select(pre.Fields)
	}

	buildSelector(tx, pre.Selectors)

	kind, err := buildClause(tx, pre.Clauses, false)
	if err != nil {
		return err
	}

	switch kind {
	case selection.First:
		tx.First(obj)
	case selection.Last:
		tx.Last(obj)
	}

	return checkErr(tx.Error, ignoreNotFound)
}

func (p PostgresSQL) List(ctx context.Context, key string, pre selection.Predicate, listObj runtime.Object) error {
	p.aspect.BeforeRequest(ctx, listObj)
	defer p.aspect.AfterRequest(ctx, listObj)

	tx := p.db.WithContext(ctx).Table(key)

	if pre.Fields != nil {
		tx.Select(pre.Fields)
	}

	buildSelector(tx, pre.Selectors)

	_, err := buildClause(tx, pre.Clauses, true)
	if err != nil {
		return err
	}

	listPtr, err := metav1.GetItemPtr(listObj)
	if err != nil {
		return fmt.Errorf("%w: %v", db.ErrInvalidObj, err)
	}
	v, err := metav1.EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return fmt.Errorf("%w: %v", db.ErrInvalidObj, "need ptr to slice")
	}

	tx.Find(v)

	if err = checkErr(tx.Error, true); err != nil {
		return err
	}

	meta := &db.ResponseMeta{}
	tx.Count(&meta.Count)
	return p.aspect.UpdateList(ctx, listObj, meta)
}

func (p PostgresSQL) Create(ctx context.Context, key string, obj, out runtime.Object) error {
	p.aspect.BeforeRequest(ctx, obj)
	defer p.aspect.AfterRequest(ctx, obj)

	tx := p.db.WithContext(ctx).Table(key)

	err := p.aspect.PrepareObjectForDB(ctx, obj)
	if err != nil {
		return err
	}

	tx.Create(obj)
	if err = checkErr(tx.Error, false); err != nil {
		return err
	}

	out = obj.DeepCopyObject()
	return p.aspect.GetObjectFromDB(ctx, key, out.DeepCopyObject(), watch.Added)
}

func (p PostgresSQL) Update(ctx context.Context, key string, pre selection.Predicate, out runtime.Object, ignoreNotFound bool) error {
	p.aspect.BeforeRequest(ctx, out)
	defer p.aspect.AfterRequest(ctx, out)

	tx := p.db.WithContext(ctx).Table(key)

	meta := &db.ResponseMeta{}
	err := p.aspect.UpdateObject(ctx, out, meta)
	if err != nil {
		return err
	}

	objPtr := out.DeepCopyObject()
	if pre.PK != nil {
		err = p.pkGo(ctx, key, pre.PK, objPtr, ignoreNotFound)
	} else {
		err = p.Get(ctx, key, pre, objPtr, ignoreNotFound)
	}
	if err != nil {
		return err
	}

	buildSelector(tx, pre.Selectors)
	tx.Updates(out)

	if err = checkErr(tx.Error, ignoreNotFound); err != nil {
		return err
	}

	return p.aspect.GetObjectFromDB(ctx, key, out.DeepCopyObject(), watch.Modified)
}

func (p PostgresSQL) Delete(ctx context.Context, key string, pre selection.Predicate, out runtime.Object) error {
	p.aspect.BeforeRequest(ctx, out)
	defer p.aspect.AfterRequest(ctx, out)

	tx := p.db.WithContext(ctx).Table(key)

	if pre.PK != nil {
		tx.Where(spotToJSONBField(pre.PK.Field)+ " = ?", toSelectString(pre.PK.Value))
	} else {
		buildSelector(tx, pre.Selectors)
	}

	tx.Delete(out)

	var err error
	if err = checkErr(tx.Error, true); err != nil {
		return err
	}

	return p.aspect.GetObjectFromDB(ctx, key, out.DeepCopyObject(), watch.Deleted)
}

func checkErr(err error, ignoreNotFound bool) error {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			if ignoreNotFound {
				err = nil
			} else {
				err = db.ErrNoRecord
			}
		} else {
			err = fmt.Errorf("%w: %v", db.ErrRequest, err)
		}
	}
	return err
}

// pkGo returns database record by primary key
func (p *PostgresSQL) pkGo(ctx context.Context, key string, pk *selection.PK, into runtime.Object, ignoreNotFound bool) error {
	tx := p.db.WithContext(ctx).Table(key)

	tx.Where(spotToJSONBField(pk.Field)+" = ?", toSelectString(pk.Value)).First(into)

	return checkErr(tx.Error, ignoreNotFound)
}

func buildSelector(tx *gorm.DB, selectors []selection.Selector) {
	for _, item := range selectors {
		var expr string
		field := spotToJSONBField(item.Field)
		switch item.Op {
		case selection.Eq:
			expr = field + " = " + toSelectString(item.Value)
		case selection.Ne:
			expr = field + " <> " + toSelectString(item.Value)
		case selection.Gt:
			expr = field + " > " + toSelectString(item.Value)
		case selection.Gte:
			expr = field + " >= " + toSelectString(item.Value)
		case selection.Lt:
			expr = field + " < " + toSelectString(item.Value)
		case selection.Lte:
			expr = field + " <= " + toSelectString(item.Value)
		case selection.In:
			expr = field + " IN " + toSelectString(item.Value)
		case selection.Not:
			expr = field + " NOT " + toSelectString(item.Value)
		case selection.Bw:
			v1, v2 := toSelectorTwoValue(item.Value)
			expr = field + " BEGIN " + v1 + " END " + v2
		case selection.Like:
			expr = field + " LIKE " + "%" + toSelectString(item.Value) + "%"
		}

		switch item.Kind {
		case selection.Where:
			tx.Where(expr)
		case selection.Or:
			tx.Or(expr)
		}
	}
}

func buildClause(tx *gorm.DB, clauses []selection.Clause, paginate bool) (kind selection.ClauseKind, err error) {
	for _, item := range clauses {
		switch {
		case item.Kind == selection.Offset && paginate:
			offset, err := strconv.Atoi(item.Expr)
			if err != nil {
				return 0, fmt.Errorf("%w: %v", db.ErrBadClause, "Expr must be integer")
			}
			tx.Offset(offset)
		case item.Kind == selection.Limit && paginate:
			limit, err := strconv.Atoi(item.Expr)
			if err != nil {
				return 0, fmt.Errorf("%w: %v", db.ErrBadClause, "Expr must be ineger")
			}
			tx.Limit(limit)
		case item.Kind == selection.Order:
			tx.Order(item.Expr)
		case item.Kind == selection.First:
			kind = item.Kind
		case item.Kind == selection.Last:
			kind = item.Kind
		}
	}
	return
}

// spotToJSONBField convert "a.b.c" to "a->'b'->'c'"
func spotToJSONBField(name string) string {
	arr := strings.Split(name, ".")
	if len(arr) > 0 {
		s := []string{}
		for index, item := range arr {
			if index > 0 {
				item = "'" + item + "'"
			}
			s = append(s, item)
		}
		return strings.Join(s, "->")
	}
	return name
}

// toSelectString converts v to specified returned value. The returned value will be used by postgreSQL sql.
func toSelectString(v interface{}) string {
	sb := bytes.NewBuffer([]byte(""))
	switch vv := v.(type) {
	case string:
		sb.WriteString("'" + vv + "'")
	case []string:
		sb.WriteRune('(')
		for index, item := range vv {
			sb.WriteString("'" + item + "'")
			if index < len(vv)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteByte(')')
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		sb.WriteString(fmt.Sprintf("%d", vv))
	case []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		sb.WriteRune('(')
		vvv := reflect.ValueOf(vv)
		for i := 0; i < vvv.Len(); i++ {
			sb.WriteString(fmt.Sprintf("%d", vvv.Index(i).Int()))
			if i < vvv.Len()-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteByte(')')
	case float32, float64:
		sb.WriteString(fmt.Sprintf("%f", vv))
	case []float32, []float64:
		sb.WriteRune('(')
		vvv := reflect.ValueOf(vv)
		for i := 0; i < vvv.Len(); i++ {
			sb.WriteString(fmt.Sprintf("%f", vvv.Index(i).Float()))
			if i < vvv.Len()-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteByte(')')
	default:
		data, err := json.Marshal(vv)
		if err == nil {
			sb.Write(data)
		}
	}

	return sb.String()
}

// toSelectorTwoValue converts v to two string type return value. It only converts the first two
// elements.
// Example:
//   s1, s2 := toSelectorTwoValue([]int{"1", "2"})
//   fmt.Println(s1, s2)
// Out:
//   '1', '2'
func toSelectorTwoValue(v interface{}) (s1 string, s2 string) {
	switch vv := v.(type) {
	case []string:
		if len(vv) > 1 {
			s1, s2 = "'"+vv[0]+"'", "'"+vv[1]+"'"
		}
	case []int, []int16, []int32, []int64, []uint, []int8, []uint16, []uint32, []uint64:
		vvv := reflect.ValueOf(vv)
		if vvv.Len() > 1 {
			s1, s2 = fmt.Sprintf("%d", vvv.Index(0).Int()), fmt.Sprintf("%d", vvv.Index(1).Int())
		}
	case []float32, []float64:
		vvv := reflect.ValueOf(vv)
		if vvv.Len() > 1 {
			s1, s2 = fmt.Sprintf("%f", vvv.Index(0).Float()), fmt.Sprintf("%f", vvv.Index(1).Float())
		}
	default:
		data, err := json.Marshal(vv)
		if err == nil {
			s1 = string(data)
		}
	}

	return
}
