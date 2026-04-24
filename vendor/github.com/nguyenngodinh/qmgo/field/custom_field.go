/*
 Copyright 2020 The Qmgo Authors.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package field

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomFields defines struct of supported custom fields
type CustomFields struct {
	createAt    string
	createBy    string
	createByKey string
	updateAt    string
	updateBy    string
	updateByKey string
	id          string
}

// CustomFieldsHook defines the interface, CustomFields return custom field user want to change
type CustomFieldsHook interface {
	CustomFields() CustomFieldsBuilder
}

// CustomFieldsBuilder defines the interface which user use to set custom fields
type CustomFieldsBuilder interface {
	SetUpdateAt(fieldName string) CustomFieldsBuilder
	SetUpdateBy(fieldName string, key string) CustomFieldsBuilder
	SetCreateAt(fieldName string) CustomFieldsBuilder
	SetCreateBy(fieldName string, key string) CustomFieldsBuilder
	SetId(fieldName string) CustomFieldsBuilder
}

// NewCustom creates new Builder which is used to set the custom fields
func NewCustom() CustomFieldsBuilder {
	return &CustomFields{}
}

// SetUpdateAt set the custom UpdateAt field
func (c *CustomFields) SetUpdateAt(fieldName string) CustomFieldsBuilder {
	c.updateAt = fieldName
	return c
}

// SetUpdateBy set the custom UpdateAt field
func (c *CustomFields) SetUpdateBy(fieldName string, key string) CustomFieldsBuilder {
	c.updateBy = fieldName
	c.updateByKey = key
	return c
}

// SetCreateAt set the custom CreateAt field
func (c *CustomFields) SetCreateAt(fieldName string) CustomFieldsBuilder {
	c.createAt = fieldName
	return c
}

// SetCreateBy set the custom CreateAt field
func (c *CustomFields) SetCreateBy(fieldName string, key string) CustomFieldsBuilder {
	c.createBy = fieldName
	c.createByKey = key
	return c
}

// SetId set the custom Id field
func (c *CustomFields) SetId(fieldName string) CustomFieldsBuilder {
	c.id = fieldName
	return c
}

// CustomCreateTime changes the custom create time
func (c CustomFields) CustomCreateTime(ctx context.Context, doc interface{}) {
	if c.createAt == "" {
		return
	}
	fieldName := c.createAt
	setTime(ctx, doc, fieldName, false)
	return
}

// CustomCreateBy changes the custom create by
func (c CustomFields) CustomCreateBy(ctx context.Context, doc interface{}) {
	if c.createBy == "" {
		return
	}
	fieldName := c.createBy
	ctxkey := c.createByKey
	setBy(ctx, doc, fieldName, ctxkey)
	return
}

// CustomUpdateTime changes the custom update time
func (c CustomFields) CustomUpdateTime(ctx context.Context, doc interface{}) {
	if c.updateAt == "" {
		return
	}
	fieldName := c.updateAt
	setUpdatedTime(ctx, doc, fieldName)
	return
}

// CustomUpdateBy changes the custom update by
func (c CustomFields) CustomUpdateBy(ctx context.Context, doc interface{}) {
	if c.createBy == "" {
		return
	}
	fieldName := c.updateBy
	ctxkey := c.updateByKey
	setBy(ctx, doc, fieldName, ctxkey)
	return
}

// CustomUpdateTime changes the custom update time
func (c CustomFields) CustomId(ctx context.Context, doc interface{}) {
	if c.id == "" {
		return
	}
	fieldName := c.id
	setId(ctx, doc, fieldName)
	return
}

// setTime changes the custom time fields
// The overWrite defines if change value when the filed has valid value
func setUpdatedTime(ctx context.Context, doc interface{}, fieldName string) {
	if reflect.Ptr != reflect.TypeOf(doc).Kind() {
		fmt.Println("not a point type")
		return
	}
	e := reflect.ValueOf(doc).Elem()
	ca := e.FieldByName(fieldName)
	fmt.Printf("Set time e:%v", e)
	fmt.Printf("Set time ca:%v ---> can set %v", ca, ca.CanSet())
	if ca.CanSet() {
		tt := time.Now()
		atype := ca.Interface()
		fmt.Printf("ca type %v ", atype)
		switch a := ca.Interface().(type) {
		case time.Time:
			ca.Set(reflect.ValueOf(tt.UnixMilli()))
		case int64:
			ca.SetInt(tt.Unix() * 1000)
		case uint64:
			t := (uint64)(tt.Unix())
			ca.SetUint(t * 1000)
		default:
			fmt.Println("unsupported type to setTime", a)
		}
	}
}

func setBy(ctx context.Context, doc interface{}, fieldName string, ctxKey string) {
	if reflect.Ptr != reflect.TypeOf(doc).Kind() {
		fmt.Println("not a point type")
		return
	}
	e := reflect.ValueOf(doc).Elem()
	ca := e.FieldByName(fieldName)
	fmt.Printf("Set time e:%v\n", e)
	fmt.Printf("Set time ca:%v ---> can set %v\n", ca, ca.CanSet())
	if ca.CanSet() {
		fmt.Printf("ca type %v ", ca.Interface())
		switch a := ca.Interface().(type) {
		case string:
			by := ctx.Value(ctxKey)
			fmt.Printf("ctx value %v --> %v", ctxKey, by.(string))
			ca.SetString(by.(string))
		default:
			fmt.Println("unsupported type to SetBy", a)
		}
	}
}

// setTime changes the custom time fields
// The overWrite defines if change value when the filed has valid value
func setTime(ctx context.Context, doc interface{}, fieldName string, overWrite bool) {
	if reflect.Ptr != reflect.TypeOf(doc).Kind() {
		fmt.Println("not a point type")
		return
	}
	e := reflect.ValueOf(doc).Elem()
	ca := e.FieldByName(fieldName)
	fmt.Printf("Set time e:%v\n", e)
	fmt.Printf("Set time ca:%v ---> can set %v\n", ca, ca.CanSet())
	if ca.CanSet() {
		tt := time.Now()
		atype := ca.Interface()
		fmt.Printf("ca type %v ", atype)
		switch a := ca.Interface().(type) {
		case time.Time:
			if ca.Interface().(time.Time).IsZero() {
				ca.Set(reflect.ValueOf(tt.UnixMilli()))
			} else if overWrite {
				ca.Set(reflect.ValueOf(tt))
			}
		case int64:
			if ca.Interface().(int64) == 0 {
				ca.SetInt(tt.Unix() * 1000)
			} else if overWrite {
				ca.SetInt(tt.Unix() * 1000)
			}
		case uint64:
			if ca.Interface().(uint64) == 0 {
				t := (uint64)(tt.Unix() * 1000)
				ca.SetUint(t)
			} else if overWrite {
				t := (uint64)(tt.Unix() * 1000)
				ca.SetUint(t)
			}
		default:
			fmt.Println("unsupported type to setTime", a)
		}
	}
}

// setId changes the custom Id fields
func setId(ctx context.Context, doc interface{}, fieldName string) {
	if reflect.Ptr != reflect.TypeOf(doc).Kind() {
		fmt.Println("not a point type")
		return
	}
	e := reflect.ValueOf(doc).Elem()
	ca := e.FieldByName(fieldName)
	if ca.CanSet() {
		switch a := ca.Interface().(type) {
		case primitive.ObjectID:
			if ca.Interface().(primitive.ObjectID).IsZero() {
				ca.Set(reflect.ValueOf(primitive.NewObjectID()))
			}
		case string:
			if ca.String() == "" {
				// ca.SetString(primitive.NewObjectID().Hex())
				ca.SetString(uuid.NewString())
			}
		default:
			fmt.Println("unsupported type to setId", a)
		}
	}
}
