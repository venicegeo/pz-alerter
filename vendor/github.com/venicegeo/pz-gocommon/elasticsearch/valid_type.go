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

package elasticsearch

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"time"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

const ipRegex = `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`

func IsValidString(typ, name string, value interface{}) error {
	if fmt.Sprint(reflect.TypeOf(value)) != "string" {
		return errors.New(fmt.Sprintf("Value of %v is not a valid String", name))
	}
	return nil
}
func IsValidStringArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid String array", name))
	}
	for _, v := range arr {
		if err := IsValidString(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("String array %v contains non-valid String: %v", name, v))
		}
	}
	return nil
}
func IsValidLong(typ, name string, value interface{}) error {
	num, ok := value.(json.Number)
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Long", name))
	}
	/*intNum*/ _, err := num.Int64()
	if err != nil {
		return err
	}
	return nil
}
func IsValidLongArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Long array", name))
	}
	for _, v := range arr {
		if err := IsValidLong(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Long array %v contains a non-valid Long: %v", name, v))

		}
	}
	return nil
}
func IsValidInteger(typ, name string, value interface{}) error {
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
	return nil
}
func IsValidIntegerArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Integer array", name))
	}
	for _, v := range arr {
		if err := IsValidInteger(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Integer array %v contains a non-valid Integer: %v", name, v))
		}
	}
	return nil
}
func IsValidShort(typ, name string, value interface{}) error {
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
	return nil
}
func IsValidShortArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Short array", name))
	}
	for _, v := range arr {
		if err := IsValidShort(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Short array %v contains a non-valid Short: %v", name, v))
		}
	}
	return nil
}
func IsValidByte(typ, name string, value interface{}) error {
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
	return nil
}
func IsValidByteArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Byte array", name))
	}
	for _, v := range arr {
		if err := IsValidByte(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Byte array %v contains a non-valid Byte: %v", name, v))
		}
	}
	return nil
}
func IsValidDouble(typ, name string, value interface{}) error {
	num, ok := value.(json.Number)
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Double", name))
	}
	/*floatNum*/ _, err := num.Float64()
	if err != nil {
		return err
	}
	return nil
}
func IsValidDoubleArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		fmt.Println(typ, name, value, reflect.TypeOf(value))
		return errors.New(fmt.Sprintf("Value of %v is not a valid Double array", name))
	}
	for _, v := range arr {
		if err := IsValidDouble(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Double array %v contains a non-valid Double: %v", name, v))
		}
	}
	return nil
}
func IsValidFloat(typ, name string, value interface{}) error {
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
	return nil
}
func IsValidFloatArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Float array", name))
	}
	for _, v := range arr {
		if err := IsValidFloat(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Float array %v contains a non-valid Float: %v", name, v))
		}
	}
	return nil
}
func IsValidBool(typ, name string, value interface{}) error {
	if reflect.TypeOf(value).Kind() != reflect.Bool {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Boolean", name))
	}
	return nil
}
func IsValidBoolArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Boolean array", name))
	}
	for _, v := range arr {
		if err := IsValidBool(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Boolean array %v contains a non-valid Boolean: %v", name, v))
		}
	}
	return nil
}
func IsValidBinary(typ, name string, value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Binary", name))
	}
	/*binary*/ _, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err
	}
	return nil
}
func IsValidBinaryArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Binary array", name))
	}
	for _, v := range arr {
		if err := IsValidBinary(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Binary array %v contains a non-valid Binary: %v", name, v))
		}

	}
	return nil
}
func IsValidIp(typ, name string, value interface{}) error {
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
	return nil
}
func IsValidIpArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid IP array", name))
	}
	for _, v := range arr {
		if err := IsValidIp(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("IP array %v contains a non-valid IP: %v", name, v))
		}
	}
	return nil
}
func IsValidDate(typ, name string, value interface{}) error {
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
	return nil
}
func IsValidDateArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Date array", name))
	}
	for _, v := range arr {
		if err := IsValidDate(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Date array %v contains a non-valid Date: %v", name, v))
		}
	}
	return nil
}
func IsValidGeoPoint(typ, name string, value interface{}) error {
	sPoint, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point", name))
	}
	if point, err := NewGeo_Point_FromJSON(sPoint); err != nil || !point.valid() {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point", name))
	}
	return nil
}
func IsValidGeoPointArray(typ, name string, value interface{}) error {
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
		if err := IsValidGeoPoint(typ, name, sPoint); err != nil {
			return errors.New(fmt.Sprintf("geo_point array %v contains a non-valid geo_point: %v", name, v))
		}
	}
	return nil
}
func IsValidGeoShape(typ, name string, value interface{}) error {
	sShape, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape", name))
	}
	shape, err := NewGeo_Shape_FromJSON(sShape)
	if err != nil {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape", name))
	}
	if ok, err := shape.valid(); !ok || err != nil {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape", name))
	}
	return nil
}
func IsValidGeoShapeArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_shape array", name))
	}
	for _, v := range arr {
		mShape, ok := v.(map[string]interface{})
		if !ok {
			return errors.New(fmt.Sprintf("geo_shape array %v contains a non-valid geo_shape: %v", name, v))
		}
		sShape, err := piazza.StructInterfaceToString(mShape)
		if err != nil {
			return err
		}
		if err := IsValidGeoShape(typ, name, sShape); err != nil {
			return errors.New(fmt.Sprintf("geo_shape array %v contains a non-valid geo_shape: %v", name, v))
		}
	}
	return nil
}
