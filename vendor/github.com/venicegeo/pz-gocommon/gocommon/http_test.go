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

package piazza

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//--------------------------

func TestHttp(t *testing.T) {
	assert := assert.New(t)

	assert.True(!false)
}

func TestToJsonString(t *testing.T) {
	assert := assert.New(t)

	m := map[string]interface{}{"k": "v"}

	h := Http{}

	s := h.toJsonString(m)
	assert.Equal(`{"k":"v"}`, RemoveWhitespace(s))
}

func TestPreflight(t *testing.T) {
	assert := assert.New(t)

	exampleurl := "http://www.example.com"

	preflight := func(verb string, url string, obj string) error {
		assert.Equal(exampleurl+"/", url)
		return nil
	}

	postflight := func(code int, obj string) error {
		assert.Equal(200, code)
		return nil
	}

	h := Http{BaseUrl: exampleurl, Preflight: preflight, Postflight: postflight}

	var s string
	code, err := h.Get("/", s)
	assert.NoError(err)
	assert.Equal(200, code)
	code, err = h.Get2("/", nil, s)
	assert.NoError(err)
	assert.Equal(200, code)
	code, err = h.Verb("GET", "/", nil, s)
	assert.NoError(err)
	assert.Equal(200, code)
	assert.NotNil(h.PzGet2("/", nil))
	SimplePreflight("", "", "")
	SimplePostflight(9000, "")
}
