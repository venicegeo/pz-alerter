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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"time"

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
		if fmt.Sprint(reflect.TypeOf(value)) != "string" {
			return errors.New(fmt.Sprintf("Value of %v is not a valid String", name))
		}
	case elasticsearch.MappingElementTypeStringA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid String array", name))
		}
		for _, v := range arr {
			if fmt.Sprint(reflect.TypeOf(v)) != "string" {
				return errors.New(fmt.Sprintf("String array %v contains non-valid String: %v", name, v))
			}
		}

	case elasticsearch.MappingElementTypeLong: //int64
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Long", name))
		}
		/*intNum*/ _, err := num.Int64()
		if err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeLongA: //int64A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Long array", name))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Long array %v contains a non-valid Long: %v", name, v))
			}
			/*intNum*/ _, err := num.Int64()
			if err != nil {
				return err
			}
		}

	case elasticsearch.MappingElementTypeInteger: //int32
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Integer", name))
		}
		intNum, err := num.Int64()
		if err != nil {
			return err
		}
		if int64(int32(intNum)) != intNum {
			return errors.New(fmt.Sprintf("Value of %v is outside the range of Integer", name))
		}
	case elasticsearch.MappingElementTypeIntegerA: //int32A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Integer array", name))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Integer array %v contains a non-valid Integer: %v", name, v))
			}
			intNum, err := num.Int64()
			if err != nil {
				return err
			}
			if int64(int32(intNum)) != intNum {
				return errors.New(fmt.Sprintf("Integer array %v contains value %d - outside the range of Integer", name, intNum))
			}
		}

	case elasticsearch.MappingElementTypeShort: //int16
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Short", name))
		}
		intNum, err := num.Int64()
		if err != nil {
			return err
		}
		if int64(int16(intNum)) != intNum {
			return errors.New(fmt.Sprintf("Value of %v is outside the range of Short", name))
		}
	case elasticsearch.MappingElementTypeShortA: //int16A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Short array", name))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Short array %v contains a non-valid Short: %v", name, v))
			}
			intNum, err := num.Int64()
			if err != nil {
				return err
			}
			if int64(int16(intNum)) != intNum {
				return errors.New(fmt.Sprintf("Short array %v contains value %d - outside the range of Short", name, intNum))
			}
		}

	case elasticsearch.MappingElementTypeByte: //int8
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Byte", name))
		}
		intNum, err := num.Int64()
		if err != nil {
			return err
		}
		if int64(int8(intNum)) != intNum {
			return errors.New(fmt.Sprintf("Value of %v is outside the range of Byte", name))
		}
	case elasticsearch.MappingElementTypeByteA: //int8A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Byte array", name))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Byte array %v contains a non-valid Byte: %v", name, v))
			}
			intNum, err := num.Int64()
			if err != nil {
				return err
			}
			if int64(int8(intNum)) != intNum {
				return errors.New(fmt.Sprintf("Byte array %v contains value %d - outside the range of Byte", name, intNum))
			}
		}

	case elasticsearch.MappingElementTypeDouble: //float64
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Double", name))
		}
		/*floatNum*/ _, err := num.Float64()
		if err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeDoubleA: //float64A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Double array", name))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Double array %v contains a non-valid Double: %v", name, v))
			}
			/*floatNum*/ _, err := num.Float64()
			if err != nil {
				return err
			}
		}

	case elasticsearch.MappingElementTypeFloat: //float32
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Float", name))
		}
		floatNum, err := num.Float64()
		if err != nil {
			return err
		}
		if floatNum > 3.4*math.Pow10(38) || floatNum < -3.4*math.Pow10(38) {
			return errors.New(fmt.Sprintf("Value of %v is outside the range of Float", name))
		}
	case elasticsearch.MappingElementTypeFloatA: //float32A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Float array", name))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Float array %v contains a non-valid Float: %v", name, v))
			}
			floatNum, err := num.Float64()
			if err != nil {
				return err
			}
			if floatNum > 3.4*math.Pow10(38) || floatNum < -3.4*math.Pow10(38) {
				return errors.New(fmt.Sprintf("Float array %v contains %f - outside the range of Float", name, floatNum))
			}
		}

	case elasticsearch.MappingElementTypeBool:
		if reflect.TypeOf(value).Kind() != reflect.Bool {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Boolean", name))
		}
	case elasticsearch.MappingElementTypeBoolA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Boolean array", name))
		}
		for _, v := range arr {
			if reflect.TypeOf(v).Kind() != reflect.Bool {
				return errors.New(fmt.Sprintf("Boolean array %v contains a non-valid Boolean: %v", name, v))
			}
		}

	case elasticsearch.MappingElementTypeBinary:
		v, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Binary", name))
		}
		/*binary*/ _, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeBinaryA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Binary array", name))
		}
		for _, vA := range arr {
			v, ok := vA.(string)
			if !ok {
				return errors.New(fmt.Sprintf("Binary array %v contains a non-valid Binary: %v", name, vA))
			}
			/*binary*/ _, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}
		}

	case elasticsearch.MappingElementTypeIp:
		ip, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid IP", name))
		}
		re, err := regexp.Compile(ipRegex)
		if err != nil {
			return err
		}
		if !re.MatchString(ip) {
			return errors.New(fmt.Sprintf("Value of %v is not a valid IP", name))
		}
	case elasticsearch.MappingElementTypeIpA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid IP array", name))
		}
		for _, v := range arr {
			ip, ok := v.(string)
			if !ok {
				return errors.New(fmt.Sprintf("IP array %v contains a non-valid IP: %v", name, v))
			}
			re, err := regexp.Compile(ipRegex)
			if err != nil {
				return err
			}
			if !re.MatchString(ip) {
				return errors.New(fmt.Sprintf("IP array %v contains a non-valid IP: %v", name, ip))
			}
		}

	case elasticsearch.MappingElementTypeDate:
		stringDate, okString := value.(string)
		milliDate, okNumber := value.(json.Number)
		if !okString && !okNumber {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Date", name))
		}
		if okString {
			_, err1 := time.Parse("2006-01-02T15:04:05Z07:00", stringDate)
			_, err2 := time.Parse("2006-01-02", stringDate)
			if err1 != nil && err2 != nil {
				return errors.New(fmt.Sprintf("Value of %v is not a valid Date", name))
			}
		} else {
			num, err := milliDate.Int64()
			if err != nil {
				return err
			}
			if num <= 0 {
				return errors.New(fmt.Sprintf("Value of %v is not a valid Date", name))
			}
		}
	case elasticsearch.MappingElementTypeDateA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid Date array", name))
		}
		for _, v := range arr {
			stringDate, okString := v.(string)
			milliDate, okNumber := v.(json.Number)
			if !okString && !okNumber {
				return errors.New(fmt.Sprintf("Date array %v contains a non-valid Date: %v", name, v))
			}
			if okString {
				_, err1 := time.Parse("2006-01-02T15:04:05Z07:00", stringDate)
				_, err2 := time.Parse("2006-01-02", stringDate)
				if err1 != nil && err2 != nil {
					return errors.New(fmt.Sprintf("Date array %v contains a non-valid Date: %v", name, stringDate))
				}
			} else {
				num, err := milliDate.Int64()
				if err != nil {
					return err
				}
				if num <= 0 {
					return errors.New(fmt.Sprintf("Date array %v contains a non-valid Date: %d", name, num))
				}
			}
		}

	case elasticsearch.MappingElementTypeGeoPoint:
		sPoint, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point", name))
		}
		var point Geo_Point
		err := json.Unmarshal([]byte(sPoint), &point)
		if err != nil {
			return err
		}
		if !point.valid() {
			return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point", name))
		}
	case elasticsearch.MappingElementTypeGeoPointA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point array", name))
		}
		for _, v := range arr {
			mPoint, ok := v.(map[string]interface{})
			if !ok {
				return errors.New(fmt.Sprintf("geo_point array %v contains a non-valid geo_point: %v", name, v))
			}
			sPoint, err := piazza.StructInterfaceToString(mPoint)
			if err != nil {
				return err
			}
			var point Geo_Point
			err = json.Unmarshal([]byte(sPoint), &point)
			if err != nil {
				return err
			}
			if !point.valid() {
				return errors.New(fmt.Sprintf("geo_point array %v contains a non-valid geo_point: %v", name, point))
			}
		}

	case elasticsearch.MappingElementTypeGeoShape:
		sShape, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape", name))
		}
		shape := Geo_Shape{}
		err := json.Unmarshal([]byte(sShape), &shape)
		if err != nil {
			return err
		}
		if ok, err := shape.valid(); !ok {
			if err != nil {
				return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape: %v", name, err.Error()))
			}
			return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape", name))
		}
	case elasticsearch.MappingElementTypeGeoShapeA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape array", name))
		}
		for _, v := range arr {
			mShape, ok := v.(map[string]interface{})
			if !ok {
				return errors.New(fmt.Sprintf("geo_shape array %v contains a non-valid geo_shape: %v", name, mShape))
			}
			sShape, err := piazza.StructInterfaceToString(mShape)
			if err != nil {
				return err
			}
			var shape Geo_Shape
			err = json.Unmarshal([]byte(sShape), &shape)
			if err != nil {
				return err
			}
			if ok, err := shape.valid(); !ok {
				if err != nil {
					return errors.New(fmt.Sprintf("geo_shape array %v is not a valid geo_shape %v. Info: [%v]", name, shape, err.Error()))
				}
				return errors.New(fmt.Sprintf("geo_shape array %v contains a non-valid geo_shape: %v", name, shape))
			}
		}
	default:
		return errors.New(fmt.Sprintf("Unknown type %v: %v", typ, name))
	}
	return nil
}
