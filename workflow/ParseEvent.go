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
	//"github.com/venicegeo/pz-gocommon/gocommon"
)

func (db *EventDB) valueIsValidType(key interface{}, value interface{}) error {
	k := fmt.Sprint(key)
	if !elasticsearch.IsValidMappingType(k) { //TODO :Array types
		return errors.New(fmt.Sprintf("Variable %s is not a valid mapping type", key))
	}

	switch elasticsearch.MappingElementTypeName(k) {
	case elasticsearch.MappingElementTypeString:
		if reflect.TypeOf(value).Kind() != reflect.String {
			return errors.New(fmt.Sprintf("Value %s is not a valid String", value))
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

	case elasticsearch.MappingElementTypeDouble: //float64
		num, ok := value.(json.Number)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Double", num))
		}
		/*floatNum*/ _, err := num.Float64()
		if err != nil {
			return err
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

	case elasticsearch.MappingElementTypeBool:
		if reflect.TypeOf(value).Kind() != reflect.Bool {
			return errors.New(fmt.Sprintf("Value %s is not a valid Boolean", value))
		}

	case elasticsearch.MappingElementTypeBinary:
		value, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid Binary", value))
		}
		/*binary*/ _, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}

	case elasticsearch.MappingElementTypeIp:
		ip, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid IP", value))
		}
		re, err := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
		if err != nil {
			return err
		}
		if !re.MatchString(ip) {
			return errors.New(fmt.Sprintf("Value %s is not a valid IP", ip))
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
	case elasticsearch.MappingElementTypeGeoShape:
		sShape, ok := value.(string)
		if !ok {
			return errors.New(fmt.Sprintf("Value %s is not a valid geo_shape", value))
		}
		//shape := NewDefaultGeo_Shape()
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
	}
	return nil
}
