package gormrepository

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

type updateBuilder interface{
	UpdateUUID(key string, value uuid.UUID) updateBuilder
	SetPtrUUID(key string, value *uuid.UUID) updateBuilder
	UpdateField(key string, value string) updateBuilder
	UpdateTime(key string, value time.Time) updateBuilder
	UpdateValue(key string, value interface{}) updateBuilder
	Build() map[string]interface{}
}

type partialUpdateBuilder struct {
	updateList map[string]interface{}
}

func PartialUpdateBuilder() updateBuilder {
	return &partialUpdateBuilder{updateList: make(map[string]interface{})}
}

func (partial *partialUpdateBuilder) UpdateUUID(key string, value uuid.UUID) updateBuilder {
	if value != uuid.Nil {
		partial.updateList[key] = value
	}
	return partial
}

func (partial *partialUpdateBuilder) SetPtrUUID(key string, value *uuid.UUID) updateBuilder {
	partial.updateList[key] = value
	return partial
}

func (partial *partialUpdateBuilder) UpdateField(key string, value string) updateBuilder {
	if value != "" {
		partial.updateList[key] = value
	}
	return partial
}

func (partial *partialUpdateBuilder) UpdateTime(key string, value time.Time) updateBuilder {
	if !value.IsZero() {
		partial.updateList[key] = value
	}
	return partial
}

func (partial *partialUpdateBuilder) UpdateValue(key string, value interface{}) updateBuilder {
 	if value == nil {
 		return partial
 	}
 	v := reflect.ValueOf(value)
 	t := v.Type()
 	if t.Kind() == reflect.Ptr && v.IsNil() {
 		return partial
 	}
 	if t.Kind() == reflect.String {
 		if v.Len() == 0 {
 			return partial
 		}
	}
	partial.updateList[key] = value
 	return partial
}

func (partial *partialUpdateBuilder) Build() map[string]interface{} {
	return partial.updateList
}
