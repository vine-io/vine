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

package dao

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vine-io/vine/lib/dao/schema"
)

func prepareValues(values []interface{}, db *DB, columnTypes []*sql.ColumnType, columns []string) {
	if db.Statement.Schema != nil {
		for idx, name := range columns {
			if field := db.Statement.Schema.LookUpField(name); field != nil {
				values[idx] = reflect.New(reflect.PtrTo(field.FieldType)).Interface()
				continue
			}
			values[idx] = new(interface{})
		}
	} else if len(columnTypes) > 0 {
		for idx, columnType := range columnTypes {
			if columnType.ScanType() != nil {
				values[idx] = reflect.New(reflect.PtrTo(columnType.ScanType())).Interface()
			} else {
				values[idx] = new(interface{})
			}
		}
	} else {
		for idx := range columns {
			values[idx] = new(interface{})
		}
	}
}

func scanIntoMap(mapValue map[string]interface{}, values []interface{}, columns []string) {
	for idx, column := range columns {
		if reflectValue := reflect.Indirect(reflect.Indirect(reflect.ValueOf(values[idx]))); reflectValue.IsValid() {
			mapValue[column] = reflectValue.Interface()
			if valuer, ok := mapValue[column].(driver.Valuer); ok {
				mapValue[column], _ = valuer.Value()
			} else if b, ok := mapValue[column].(sql.RawBytes); ok {
				mapValue[column] = string(b)
			}
		} else {
			mapValue[column] = nil
		}
	}
}

func Scan(rows *sql.Rows, db *DB, initialized bool) {
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	db.RowsAffected = 0

	switch dest := db.Statement.Dest.(type) {
	case map[string]interface{}, *map[string]interface{}:
		if initialized || rows.Next() {
			columnTypes, _ := rows.ColumnTypes()
			prepareValues(values, db, columnTypes, columns)

			db.RowsAffected++
			db.AddError(rows.Scan(values...))

			mapValue, ok := dest.(map[string]interface{})
			if !ok {
				if v, ok := dest.(*map[string]interface{}); ok {
					mapValue = *v
				}
			}
			scanIntoMap(mapValue, values, columns)
		}
	case *[]map[string]interface{}:
		columnTypes, _ := rows.ColumnTypes()
		for initialized || rows.Next() {
			prepareValues(values, db, columnTypes, columns)

			initialized = false
			db.RowsAffected++
			db.AddError(rows.Scan(values...))

			mapValue := map[string]interface{}{}
			scanIntoMap(mapValue, values, columns)
			*dest = append(*dest, mapValue)
		}
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
		*float32, *float64,
		*bool, *string, *time.Time,
		*sql.NullInt32, *sql.NullInt64, *sql.NullFloat64,
		*sql.NullBool, *sql.NullString, *sql.NullTime:
		for initialized || rows.Next() {
			initialized = false
			db.RowsAffected++
			db.AddError(rows.Scan(dest))
		}
	default:
		Schema := db.Statement.Schema

		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			var (
				reflectValueType = db.Statement.ReflectValue.Type().Elem()
				isPtr            = reflectValueType.Kind() == reflect.Ptr
				fields           = make([]*schema.Field, len(columns))
				joinFields       [][2]*schema.Field
			)

			if isPtr {
				reflectValueType = reflectValueType.Elem()
			}

			db.Statement.ReflectValue.Set(reflect.MakeSlice(db.Statement.ReflectValue.Type(), 0, 20))

			if Schema != nil {
				if reflectValueType != Schema.ModelType && reflectValueType.Kind() == reflect.Struct {
					Schema, _ = schema.Parse(db.Statement.Dest, db.cacheStore, db.NamingStrategy)
				}

				for idx, column := range columns {
					if field := Schema.LookUpField(column); field != nil && field.Readable {
						fields[idx] = field
					} else if names := strings.Split(column, "__"); len(names) > 1 {
						if rel, ok := Schema.Relationships.Relations[names[0]]; ok {
							if field := rel.FieldSchema.LookUpField(strings.Join(names[1:], "__")); field != nil && field.Readable {
								fields[idx] = field

								if len(joinFields) == 0 {
									joinFields = make([][2]*schema.Field, len(columns))
								}
								joinFields[idx] = [2]*schema.Field{rel.Field, field}
								continue
							}
						}
						values[idx] = &sql.RawBytes{}
					} else {
						values[idx] = &sql.RawBytes{}
					}
				}
			}

			// pluck values into slice of data
			isPluck := false
			if len(fields) == 1 {
				if _, ok := reflect.New(reflectValueType).Interface().(sql.Scanner); ok || // is scanner
					reflectValueType.Kind() != reflect.Struct || // is not struct
					Schema.ModelType.ConvertibleTo(schema.TimeReflectType) { // is time
					isPluck = true
				}
			}

			for initialized || rows.Next() {
				initialized = false
				db.RowsAffected++

				elem := reflect.New(reflectValueType)
				if isPluck {
					db.AddError(rows.Scan(elem.Interface()))
				} else {
					for idx, field := range fields {
						if field != nil {
							values[idx] = reflect.New(reflect.PtrTo(field.IndirectFieldType)).Interface()
						}
					}

					db.AddError(rows.Scan(values...))

					for idx, field := range fields {
						if len(joinFields) != 0 && joinFields[idx][0] != nil {
							value := reflect.ValueOf(values[idx]).Elem()
							relValue := joinFields[idx][0].ReflectValueOf(elem)

							if relValue.Kind() == reflect.Ptr && relValue.IsNil() {
								if value.IsNil() {
									continue
								}
								relValue.Set(reflect.New(relValue.Type().Elem()))
							}

							field.Set(relValue, values[idx])
						} else if field != nil {
							field.Set(elem, values[idx])
						}
					}
				}

				if isPtr {
					db.Statement.ReflectValue.Set(reflect.Append(db.Statement.ReflectValue, elem))
				} else {
					db.Statement.ReflectValue.Set(reflect.Append(db.Statement.ReflectValue, elem.Elem()))
				}
			}
		case reflect.Struct, reflect.Ptr:
			if db.Statement.ReflectValue.Type() != Schema.ModelType {
				Schema, _ = schema.Parse(db.Statement.Dest, db.cacheStore, db.NamingStrategy)
			}

			if initialized || rows.Next() {
				for idx, column := range columns {
					if field := Schema.LookUpField(column); field != nil && field.Readable {
						values[idx] = reflect.New(reflect.PtrTo(field.IndirectFieldType)).Interface()
					} else if names := strings.Split(column, "__"); len(names) > 1 {
						if rel, ok := Schema.Relationships.Relations[names[0]]; ok {
							if field := rel.FieldSchema.LookUpField(strings.Join(names[1:], "__")); field != nil && field.Readable {
								values[idx] = reflect.New(reflect.PtrTo(field.IndirectFieldType)).Interface()
								continue
							}
						}
						values[idx] = &sql.RawBytes{}
					} else {
						values[idx] = &sql.RawBytes{}
					}
				}

				db.RowsAffected++
				db.AddError(rows.Scan(values...))

				for idx, column := range columns {
					if field := Schema.LookUpField(column); field != nil && field.Readable {
						field.Set(db.Statement.ReflectValue, values[idx])
					} else if names := strings.Split(column, "__"); len(names) > 1 {
						if rel, ok := Schema.Relationships.Relations[names[0]]; ok {
							if field := rel.FieldSchema.LookUpField(strings.Join(names[1:], "__")); field != nil && field.Readable {
								relValue := rel.Field.ReflectValueOf(db.Statement.ReflectValue)
								value := reflect.ValueOf(values[idx]).Elem()

								if relValue.Kind() == reflect.Ptr && relValue.IsNil() {
									if value.IsNil() {
										continue
									}
									relValue.Set(reflect.New(relValue.Type().Elem()))
								}

								field.Set(relValue, values[idx])
							}
						}
					}
				}
			}
		}
	}

	if db.RowsAffected == 0 && db.Statement.RaiseErrorOnNotFound {
		db.AddError(ErrRecordNotFound)
	}
}

func patch(obj reflect.Value, outs map[string]interface{}, key string) {
	vo := reflectDirect(obj)
	switch vo.Kind() {
	case reflect.String:
		vv := vo.String()
		if vv != "" {
			outs[key] = vv
		}
	case reflect.Float32, reflect.Float64:
		vv := vo.Float()
		if vv != 0 {
			outs[key] = vv
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vo.Int()
		if vv != 0 {
			outs[key] = vv
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vo.Uint()
		if vv != 0 {
			outs[key] = vv
		}
	case reflect.Map:
		for _, mkv := range vo.MapKeys() {
			var kk string
			switch mkv.Type().Kind() {
			case reflect.String:
				kk = key + "." + mkv.String()
			case reflect.Float32, reflect.Float64:
				kk = key + "." + fmt.Sprintf("%f", mkv.Float())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				kk = key + "." + fmt.Sprintf("%d", mkv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				kk = key + "." + fmt.Sprintf("%d", mkv.Uint())
			default:
				continue
			}
			mvv := vo.MapIndex(mkv)
			switch mvv.Type().Kind() {
			case reflect.String:
				outs[kk] = mvv.String()
			case reflect.Float32, reflect.Float64:
				outs[kk] = mvv.Float()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				outs[kk] = mvv.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				outs[kk] = mvv.Uint()
			case reflect.Map:
				patch(mvv, outs, kk)
			case reflect.Ptr:
				mvv = reflectDirect(mvv)
				fallthrough
			case reflect.Struct:
				if !mvv.IsValid() {
					continue
				}
				if mvv.IsZero() {
					outs[kk] = nil
					continue
				}

				for i := 0; i < mvv.NumField(); i++ {
					mvvd := mvv.Type().Field(i)
					mvfdv := mvv.Field(i)
					jsonName := strings.Split(mvvd.Tag.Get("json"), ",")[0]
					patch(mvfdv, outs, kk+"."+jsonName)
				}
			}

		}
	case reflect.Struct:
		if !vo.IsValid() {
			return
		}
		if vo.IsZero() {
			outs[key] = nil
			return
		}
		for i := 0; i < vo.Type().NumField(); i++ {
			kk := ""
			fd := vo.Type().Field(i)
			fdv := vo.Field(i)

			if !fdv.IsValid() {
				continue
			}

			if fdv.IsZero() {
				continue
			}

			jsonName := strings.Split(fd.Tag.Get("json"), ",")[0]
			if key == "" {
				kk = jsonName
			} else {
				kk = key + "." + jsonName
			}

			patch(fdv, outs, kk)
		}
	}
}

func reflectDirect(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		for v.Type().Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}
	return v
}

func FieldPatch(v interface{}) map[string]interface{} {
	outs := make(map[string]interface{})
	patch(reflect.ValueOf(v), outs, "")
	return outs
}
