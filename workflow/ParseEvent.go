// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"errors"
	"fmt"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

func (db *EventDB) valueIsValidType(typi interface{}, nami interface{}, value interface{}) error {
	typ := fmt.Sprint(typi)
	name := fmt.Sprint(nami)
	if !elasticsearch.IsValidMappingType(typ) {
		return errors.New(fmt.Sprintf("Variable %v is not a valid mapping type", typ))
	}
	if elasticsearch.IsValidArrayTypeMapping(typ) {
		if !piazza.ValueIsValidArray(value) {
			return errors.New(fmt.Sprintf("Value %v was passed into array field %v", value, name))
		}
	} else {
		if piazza.ValueIsValidArray(value) {
			return errors.New(fmt.Sprintf("Value %v was passed into non-array field %v", value, value))
		}
	}

	switch elasticsearch.MappingElementTypeName(typ) {

	case elasticsearch.MappingElementTypeString:
		if err := elasticsearch.IsValidString(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeStringA:
		if err := elasticsearch.IsValidStringArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeLong: //int64
		if err := elasticsearch.IsValidLong(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeLongA: //int64A
		if err := elasticsearch.IsValidLongArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeInteger: //int32
		if err := elasticsearch.IsValidInteger(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeIntegerA: //int32A
		if err := elasticsearch.IsValidIntegerArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeShort: //int16
		if err := elasticsearch.IsValidShort(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeShortA: //int16A
		if err := elasticsearch.IsValidShortArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeByte: //int8
		if err := elasticsearch.IsValidByte(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeByteA: //int8A
		if err := elasticsearch.IsValidByteArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeDouble: //float64
		if err := elasticsearch.IsValidDouble(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeDoubleA: //float64A
		if err := elasticsearch.IsValidDoubleArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeFloat: //float32
		if err := elasticsearch.IsValidFloat(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeFloatA: //float32A
		if err := elasticsearch.IsValidFloatArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeBool:
		if err := elasticsearch.IsValidBool(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeBoolA:
		if err := elasticsearch.IsValidBoolArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeBinary:
		if err := elasticsearch.IsValidBinary(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeBinaryA:
		if err := elasticsearch.IsValidBinaryArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeIp:
		if err := elasticsearch.IsValidIp(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeIpA:
		if err := elasticsearch.IsValidIpArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeDate:
		if err := elasticsearch.IsValidDate(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeDateA:
		if err := elasticsearch.IsValidDateArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeGeoPoint:
		if err := elasticsearch.IsValidGeoPoint(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeGeoPointA:
		if err := elasticsearch.IsValidGeoPointArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeGeoShape:
		if err := elasticsearch.IsValidGeoShape(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeGeoShapeA:
		if err := elasticsearch.IsValidGeoShapeArray(typ, name, value); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unknown type %v: %v", typ, name))
	}
	return nil
}
