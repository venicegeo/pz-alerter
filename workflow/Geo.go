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
	"encoding/json"
	"regexp"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

type Geo_Point struct {
	Lat float64 `json:"lat" binding:"required"`
	Lon float64 `json:"lon" binding:"required"`
}

func (p *Geo_Point) valid() bool { //TODO
	return true
}

type Geo_Shape struct {
	Type             string      `json:"type" binding:"required"`
	Coordinates      interface{} `json:"coordinates" binding:"required"`
	Tree             string      `json:"tree,omitempty"`
	Precision        string      `json:"precision,omitempty"`
	TreeLevels       string      `json:"tree_levels,omitempty"`
	Strategy         string      `json:"strategy,omitempty"`
	DistanceErrorPct float64     `json:"distance_error_pct,omitempty"`
	Orientation      string      `json:"orientation,omitempty"`
	PointsOnly       bool        `json:"points_only,omitempty"`
	Radius           string      `json:"radius,omitempty"`
}

type geo_GeometryCollection []Geo_Shape
type geo_Sub_Point []interface{}
type geo_LineString []geo_Sub_Point
type geo_Polygon [][]geo_Sub_Point
type geo_MultiPoint []geo_Sub_Point
type geo_MultiLineString []geo_LineString
type geo_MultiPolygon []geo_Polygon
type geo_Envelope []geo_Sub_Point
type geo_Circle geo_Sub_Point

func NewDefaultGeo_Shape() Geo_Shape {
	return Geo_Shape{Tree: "geohash", Precision: "meters", TreeLevels: "50m", Strategy: "recursive", DistanceErrorPct: 0.025, Orientation: "ccw", PointsOnly: false}
}

func (gs *Geo_Shape) valid() (bool, error) {
	if ok, err := gs.validTree(gs.Tree); !ok {
		return false, err
	}
	if ok, err := gs.validPrecision(gs.Precision); !ok {
		return false, err
	}
	if ok, err := gs.validTreeLevels(gs.TreeLevels); !ok {
		return false, err
	}
	if ok, err := gs.validStrategy(gs.Strategy); !ok {
		return false, err
	}
	if ok, err := gs.validDistanceErrorPct(gs.DistanceErrorPct); !ok {
		return false, err
	}
	if ok, err := gs.validOrientation(gs.Orientation); !ok {
		return false, err
	}
	switch gs.Type {
	case "geometrycollection":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_GeometryCollection
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "point":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_Sub_Point
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "linestring":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_LineString
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "polygon":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_Polygon
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "multipoint":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_MultiPoint
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "multilinestring":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_MultiLineString
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "multipolygon":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_MultiPolygon
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "envelope":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_Envelope
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	case "circle":
		str, err := piazza.StructInterfaceToString(gs.Coordinates)
		if err != nil {
			return false, err
		}
		var temp geo_Circle
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			return false, err
		}
		return temp.valid(gs)
	}
	return true, nil
}

func (gc *geo_GeometryCollection) valid(gs *Geo_Shape) (bool, error) {
	for _, v := range *gc {
		if ok, err := v.valid(); !ok {
			return false, err
		}
	}
	return true, nil
}
func (p *geo_Sub_Point) valid(gs *Geo_Shape) (bool, error) { //TODO
	if len(*p) != 2 {
		return false, nil
	}
	for _, v := range *p {
		if /*key*/ _, ok := v.(float64); !ok {
			return false, nil
		}
	}
	return true, nil
}
func (ls *geo_LineString) valid(gs *Geo_Shape) (bool, error) {
	for _, v := range *ls {

		if ok, _ := v.valid(gs); !ok {
			return false, nil
		}
	}
	return true, nil
}
func (ply *geo_Polygon) valid(gs *Geo_Shape) (bool, error) {
	if len(*ply) < 1 {
		return false, nil
	}
	for _, v := range *ply {
		if len(v) != 5 {
			return false, nil
		}
		for _, p := range v {
			if ok, err := p.valid(gs); !ok {
				return false, err
			}
		}
	}
	return true, nil
}
func (mp *geo_MultiPoint) valid(gs *Geo_Shape) (bool, error) {
	for _, p := range *mp {
		if ok, err := p.valid(gs); !ok {
			return false, err
		}
	}
	return true, nil
}
func (mls *geo_MultiLineString) valid(gs *Geo_Shape) (bool, error) {
	for _, ls := range *mls {
		if ok, err := ls.valid(gs); !ok {
			return false, err
		}
	}
	return true, nil
}
func (mply *geo_MultiPolygon) valid(gs *Geo_Shape) (bool, error) {
	for _, ply := range *mply {
		if ok, err := ply.valid(gs); !ok {
			return false, err
		}
	}
	return true, nil
}
func (e *geo_Envelope) valid(gs *Geo_Shape) (bool, error) {
	if len(*e) != 2 {
		return false, nil
	}
	for _, p := range *e {
		if ok, err := p.valid(gs); !ok {
			return false, err
		}
	}
	return true, nil
}
func (c *geo_Circle) valid(gs *Geo_Shape) (bool, error) {
	p := geo_Sub_Point(*c)
	if ok, err := p.valid(gs); !ok {
		return false, err
	}
	return gs.validRadius(gs.Radius)
}

func (gs *Geo_Shape) validDistance(distance string) (bool, error) {
	re, err := regexp.Compile(`^(([1-9][0-9]*)((in)|(inch)|(yd)|(yard)|(mi)|(miles)|(km)|(kilometers)|(m)|(meters)|(cm)|(centimeters)|(mm)|(millimeters)|$))$`)
	if err != nil {
		return false, err
	}
	return re.MatchString(distance), nil
}
func (gs *Geo_Shape) validRadius(radius string) (bool, error) {
	return gs.validDistance(radius)
}
func (gs *Geo_Shape) validTree(tree string) (bool, error) {
	return tree == "geohash" || tree == "quadtree", nil
}
func (gs *Geo_Shape) validPrecision(precision string) (bool, error) {
	re, err := regexp.Compile(`^((in)|(inch)|(yd)|(yard)|(mi)|(miles)|(km)|(kilometers)|(m)|(meters)|(cm)|(centimeters)|(mm)|(millimeters))$`)
	if err != nil {
		return false, err
	}
	return re.MatchString(precision), nil
}
func (gs *Geo_Shape) validTreeLevels(treeLevels string) (bool, error) {
	return gs.validDistance(treeLevels)
}
func (gs *Geo_Shape) validStrategy(strategy string) (bool, error) {
	return strategy == "recursive" || strategy == "term", nil
}
func (gs *Geo_Shape) validDistanceErrorPct(errorPct float64) (bool, error) {
	return errorPct >= 0 && errorPct <= 100, nil
}
func (gs *Geo_Shape) validOrientation(orientation string) (bool, error) {
	re, err := regexp.Compile(`^((right)|(ccw)|(counterclockwise)|(left)|(cw)|(clockwise))$`)
	if err != nil {
		return false, err
	}
	return re.MatchString(orientation), nil
}
