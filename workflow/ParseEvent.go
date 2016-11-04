// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
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

func (db *EventDB) valueIsValidType(key interface{}, value interface{}) error {
	k := fmt.Sprint(key)
	if !elasticsearch.IsValidMappingType(k) {
		return errors.New(fmt.Sprintf("Variable %s is not a valid mapping type", key))
	}
	if elasticsearch.IsValidArrayTypeMapping(k) {
		if !piazza.ValueIsValidArray(value) {
			return errors.New(fmt.Sprintf("Value %s was passed into an array field", value))
		}
	} else {
		if piazza.ValueIsValidArray(value) {
			return errors.New(fmt.Sprintf("Value %s was passed into a non-array field", value))
		}
	}

	switch elasticsearch.MappingElementTypeName(k) {
	case elasticsearch.MappingElementTypeString:
		if reflect.TypeOf(value).Kind() != reflect.String {
			return errors.New(fmt.Sprintf("Value %s is not a valid String", value))
		}
	case elasticsearch.MappingElementTypeStringA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid String array", value))
		}
		for _, v := range arr {
			if reflect.TypeOf(v).Kind() != reflect.String {
				return errors.New(fmt.Sprintf("Value %s is not a valid String", value))
			}
		}

	case elasticsearch.MappingElementTypeLong: //int64
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Long", value))
		}
		/*intNum*/ _, err := num.Int64()
		if err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeLongA: //int64A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Long array", value))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Long", value))
			}
			/*intNum*/ _, err := num.Int64()
			if err != nil {
				return err
			}
		}

	case elasticsearch.MappingElementTypeInteger: //int32
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Integer", value))
		}
		intNum, err := num.Int64()
		if err != nil {
			return err
		}
		if int64(int32(intNum)) != intNum {
			return errors.New(fmt.Sprintf("Value %d is outside the range of Integer", intNum))
		}
	case elasticsearch.MappingElementTypeIntegerA: //int32A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Integer array", value))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Integer", value))
			}
			intNum, err := num.Int64()
			if err != nil {
				return err
			}
			if int64(int32(intNum)) != intNum {
				return errors.New(fmt.Sprintf("Value %d is outside the range of Integer", intNum))
			}
		}

	case elasticsearch.MappingElementTypeShort: //int16
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Short", value))
		}
		intNum, err := num.Int64()
		if err != nil {
			return err
		}
		if int64(int16(intNum)) != intNum {
			return errors.New(fmt.Sprintf("Value %d is outside the range of Short", intNum))
		}
	case elasticsearch.MappingElementTypeShortA: //int16A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Short array", value))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Short", value))
			}
			intNum, err := num.Int64()
			if err != nil {
				return err
			}
			if int64(int16(intNum)) != intNum {
				return errors.New(fmt.Sprintf("Value %d is outside the range of Short", intNum))
			}
		}

	case elasticsearch.MappingElementTypeByte: //int8
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Byte", value))
		}
		intNum, err := num.Int64()
		if err != nil {
			return err
		}
		if int64(int8(intNum)) != intNum {
			return errors.New(fmt.Sprintf("Value %d is outside the range of Byte", intNum))
		}
	case elasticsearch.MappingElementTypeByteA: //int8A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Byte array", value))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Byte", value))
			}
			intNum, err := num.Int64()
			if err != nil {
				return err
			}
			if int64(int8(intNum)) != intNum {
				return errors.New(fmt.Sprintf("Value %d is outside the range of Byte", intNum))
			}
		}

	case elasticsearch.MappingElementTypeDouble: //float64
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Double", value))
		}
		/*floatNum*/ _, err := num.Float64()
		if err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeDoubleA: //float64A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Double array", value))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Double", value))
			}
			/*floatNum*/ _, err := num.Float64()
			if err != nil {
				return err
			}
		}

	case elasticsearch.MappingElementTypeFloat: //float32
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Float", value))
		}
		floatNum, err := num.Float64()
		if err != nil {
			return err
		}
		if floatNum > 3.4*math.Pow10(38) || floatNum < -3.4*math.Pow10(38) {
			return errors.New(fmt.Sprintf("Value %f is outside the range of Float", floatNum))
		}
	case elasticsearch.MappingElementTypeFloatA: //float32A
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Float array", value))
		}
		for _, v := range arr {
			num, ok := v.(json.Number)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Float", value))
			}
			floatNum, err := num.Float64()
			if err != nil {
				return err
			}
			if floatNum > 3.4*math.Pow10(38) || floatNum < -3.4*math.Pow10(38) {
				return errors.New(fmt.Sprintf("Value %f is outside the range of Float", floatNum))
			}
		}

	case elasticsearch.MappingElementTypeBool:
		if reflect.TypeOf(value).Kind() != reflect.Bool {
			return errors.New(fmt.Sprintf("Value %s is not a valid Boolean", value))
		}
	case elasticsearch.MappingElementTypeBoolA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Boolean array", value))
		}
		for _, v := range arr {
			if reflect.TypeOf(v).Kind() != reflect.Bool {
				return errors.New(fmt.Sprintf("Value %s is not a valid Boolean", value))
			}
		}

	case elasticsearch.MappingElementTypeBinary:
		v, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Binary", value))
		}
		/*binary*/ _, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
	case elasticsearch.MappingElementTypeBinaryA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Binary array", value))
		}
		for _, vA := range arr {
			v, ok := vA.(string)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid Binary", value))
			}
			/*binary*/ _, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}
		}

	case elasticsearch.MappingElementTypeIp:
		ip, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid IP", value))
		}
		re, err := regexp.Compile(ipRegex)
		if err != nil {
			return err
		}
		if !re.MatchString(ip) {
			return errors.New(fmt.Sprintf("Value %s is not a valid IP", ip))
		}
	case elasticsearch.MappingElementTypeIpA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid IP array", value))
		}
		for _, v := range arr {
			ip, ok := v.(string)
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid IP", value))
			}
			re, err := regexp.Compile(ipRegex)
			if err != nil {
				return err
			}
			if !re.MatchString(ip) {
				return errors.New(fmt.Sprintf("Value %s is not a valid IP", ip))
			}
		}

	case elasticsearch.MappingElementTypeDate:
		stringDate, okString := value.(string)
		milliDate, okNumber := value.(json.Number)
		if !okString && !okNumber {
			return errors.New(fmt.Sprintf("Value %s is not a valid Date", value))
		}
		if okString {
			_, err1 := time.Parse("2006-01-02T15:04:05Z07:00", stringDate)
			_, err2 := time.Parse("2006-01-02", stringDate)
			if err1 != nil && err2 != nil {
				return err2
			}
		} else {
			num, err := milliDate.Int64()
			if err != nil {
				return err
			}
			if num <= 0 {
				return errors.New(fmt.Sprintf("Value %d is not a valid Date", num))
			}
		}
	case elasticsearch.MappingElementTypeDateA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Date array", value))
		}
		for _, v := range arr {
			stringDate, okString := v.(string)
			milliDate, okNumber := v.(json.Number)
			if !okString && !okNumber {
				return errors.New(fmt.Sprintf("Value %s is not a valid Date", value))
			}
			if okString {
				_, err1 := time.Parse("2006-01-02T15:04:05Z07:00", stringDate)
				_, err2 := time.Parse("2006-01-02", stringDate)
				if err1 != nil && err2 != nil {
					return err2
				}
			} else {
				num, err := milliDate.Int64()
				if err != nil {
					return err
				}
				if num <= 0 {
					return errors.New(fmt.Sprintf("Value %d is not a valid Date", num))
				}
			}
		}

	case elasticsearch.MappingElementTypeGeoPoint:
		sPoint, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_point", value))
		}
		var point Geo_Point
		err := json.Unmarshal([]byte(sPoint), &point)
		if err != nil {
			return err
		}
		if !point.valid() {
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_point", value))
		}
	case elasticsearch.MappingElementTypeGeoPointA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_point array", value))
		}
		for _, v := range arr {
			mPoint, ok := v.(map[string]interface{})
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid geo_point", value))
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
				return errors.New(fmt.Sprintf("Value %s is not a valid geo_point", value))
			}
		}

	case elasticsearch.MappingElementTypeGeoShape:
		sShape, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape", value))
		}
		shape := Geo_Shape{}
		err := json.Unmarshal([]byte(sShape), &shape)
		if err != nil {
			return err
		}
		if ok, err := shape.valid(); !ok {
			if err != nil {
				return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape: %s", value, err.Error()))
			}
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape", value))
		}
	case elasticsearch.MappingElementTypeGeoShapeA:
		arr, ok := value.([]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape array", value))
		}
		for _, v := range arr {
			mShape, ok := v.(map[string]interface{})
			if !ok {
				return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape", value))
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
					return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape: %s", value, err.Error()))
				}
				return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape", value))
			}
		}
	default:
		return errors.New(fmt.Sprintf("Unknown type %s", k))
	}
	return nil
}
