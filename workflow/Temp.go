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

	"github.com/venicegeo/pz-gocommon/gocommon"
)

func isValidString(typ, name string, value interface{}) error {
	if fmt.Sprint(reflect.TypeOf(value)) != "string" {
		return errors.New(fmt.Sprintf("Value of %v is not a valid String", name))
	}
	return nil
}
func isValidStringArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid String array", name))
	}
	for _, v := range arr {
		if err := isValidString(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("String array %v contains non-valid String: %v", name, v))
		}
	}
	return nil
}
func isValidLong(typ, name string, value interface{}) error {
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
func isValidLongArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Long array", name))
	}
	for _, v := range arr {
		if err := isValidLong(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Long array %v contains a non-valid Long: %v", name, v))

		}
	}
	return nil
}
func isValidInteger(typ, name string, value interface{}) error {
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
func isValidIntegerArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Integer array", name))
	}
	for _, v := range arr {
		if err := isValidInteger(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Integer array %v contains a non-valid Integer: %v", name, v))
		}
	}
	return nil
}
func isValidShort(typ, name string, value interface{}) error {
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
func isValidShortArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Short array", name))
	}
	for _, v := range arr {
		if err := isValidShort(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Short array %v contains a non-valid Short: %v", name, v))
		}
	}
	return nil
}
func isValidByte(typ, name string, value interface{}) error {
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
func isValidByteArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Byte array", name))
	}
	for _, v := range arr {
		if err := isValidByte(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Byte array %v contains a non-valid Byte: %v", name, v))
		}
	}
	return nil
}
func isValidDouble(typ, name string, value interface{}) error {
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
func isValidDoubleArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		fmt.Println(typ, name, value, reflect.TypeOf(value))
		return errors.New(fmt.Sprintf("Value of %v is not a valid Double array", name))
	}
	for _, v := range arr {
		if err := isValidDouble(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Double array %v contains a non-valid Double: %v", name, v))
		}
	}
	return nil
}
func isValidFloat(typ, name string, value interface{}) error {
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
func isValidFloatArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Float array", name))
	}
	for _, v := range arr {
		if err := isValidFloat(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Float array %v contains a non-valid Float: %v", name, v))
		}
	}
	return nil
}
func isValidBool(typ, name string, value interface{}) error {
	if reflect.TypeOf(value).Kind() != reflect.Bool {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Boolean", name))
	}
	return nil
}
func isValidBoolArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Boolean array", name))
	}
	for _, v := range arr {
		if err := isValidBool(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Boolean array %v contains a non-valid Boolean: %v", name, v))
		}
	}
	return nil
}
func isValidBinary(typ, name string, value interface{}) error {
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
func isValidBinaryArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Binary array", name))
	}
	for _, v := range arr {
		if err := isValidBinary(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Binary array %v contains a non-valid Binary: %v", name, v))
		}

	}
	return nil
}
func isValidIp(typ, name string, value interface{}) error {
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
func isValidIpArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid IP array", name))
	}
	for _, v := range arr {
		if err := isValidIp(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("IP array %v contains a non-valid IP: %v", name, v))
		}
	}
	return nil
}
func isValidDate(typ, name string, value interface{}) error {
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
func isValidDateArray(typ, name string, value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid Date array", name))
	}
	for _, v := range arr {
		if err := isValidDate(typ, name, v); err != nil {
			return errors.New(fmt.Sprintf("Date array %v contains a non-valid Date: %v", name, v))
		}
	}
	return nil
}
func isValidGeoPoint(typ, name string, value interface{}) error {
	sPoint, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point", name))
	}
	if point, err := NewGeo_Point_FromJSON(sPoint); err != nil || !point.valid() {
		return errors.New(fmt.Sprintf("Value of %v is not a valid geo_point", name))
	}
	return nil
}
func isValidGeoPointArray(typ, name string, value interface{}) error {
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
		if err := isValidGeoPoint(typ, name, sPoint); err != nil {
			return errors.New(fmt.Sprintf("geo_point array %v contains a non-valid geo_point: %v", name, v))
		}
	}
	return nil
}
func isValidGeoShape(typ, name string, value interface{}) error {
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
	return nil
}
func isValidGeoShapeArray(typ, name string, value interface{}) error {
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
	return nil
}
