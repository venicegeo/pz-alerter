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

const ipRegex = `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`

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
		if err := isValidString(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeStringA:
		if err := isValidStringArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeLong: //int64
		if err := isValidLong(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeLongA: //int64A
		if err := isValidLongArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeInteger: //int32
		if err := isValidInteger(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeIntegerA: //int32A
		if err := isValidIntegerArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeShort: //int16
		if err := isValidShort(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeShortA: //int16A
		if err := isValidShortArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeByte: //int8
		if err := isValidByte(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeByteA: //int8A
		if err := isValidByteArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeDouble: //float64
		if err := isValidDouble(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeDoubleA: //float64A
		if err := isValidDoubleArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeFloat: //float32
		if err := isValidFloat(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeFloatA: //float32A
		if err := isValidFloatArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeBool:
		if err := isValidBool(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeBoolA:
		if err := isValidBoolArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeBinary:
		if err := isValidBinary(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeBinaryA:
		if err := isValidBinaryArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeIp:
		if err := isValidIp(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeIpA:
		if err := isValidIpArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeDate:
		if err := isValidDate(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeDateA:
		if err := isValidDateArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeGeoPoint:
		if err := isValidGeoPoint(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeGeoPointA:
		if err := isValidGeoPointArray(typ, name, value); err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeGeoShape:
		if err := isValidGeoShape(typ, name, value); err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeGeoShapeA:
		if err := isValidGeoShapeArray(typ, name, value); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unknown type %v: %v", typ, name))
	}
	return nil
}
