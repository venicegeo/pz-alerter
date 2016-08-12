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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MappingTester struct {
	suite.Suite
}

func (suite *MappingTester) SetupSuite() {
}

func (suite *MappingTester) TearDownSuite() {
}

//---------------------------------------------------------------------------

var mappingTestData = []([2]string){
	{ // 0
		`{
            "a": "float",
            "b": {
                "ba:": {
                    "baa": "integer"
                },
                "bb": "boolean"
            },
            "c": "string"
        }`,
		`{
            "a": {
                "type":"float"
            },
            "b": {
                "dynamic":"strict",
                "properties": {
                    "ba:": {
                        "dynamic":"strict",
                        "properties":{
                            "baa":{
                                "type": "integer"
                            }
                        }
                    },
                    "bb": {
                        "type":"boolean"
                    }
                }
            },
            "c": {
                "type":"string"
            }
        }`,
	},
	{ //  1
		`{
          "my_type":{
              "region":"string",
              "manager":{
                  "name":{
                      "first":"string",
                      "last":"string"
                    },
                    "age":"integer",
                    "loc":{
                        "lat":"double","lon":"double"
                    }
                }
            }
        }`,
		`{
            "my_type":{
                "dynamic":"strict",
                "properties":{
                    "region":{"type":"string"},
                    "manager":{
		                "dynamic":"strict",
                        "properties":{
                            "name":{
				                "dynamic":"strict",
                                "properties":{
                                    "first":{"type":"string"},
                                    "last":{"type":"string"}
                                }
                            },
                            "age":{"type":"integer"},
                            "loc":{
				                "dynamic":"strict",
                                "properties":{
                                    "lat":{"type":"double"},
                                    "lon":{"type":"double"}
                                }
                            }
                        }
                    }
                }
            }
        }`,
	},
	{ // 2
		`{"name":"string","age":"integer"}`,
		`{"name":{"type":"string"},"age":{"type":"integer"}}`,
	},
	{ // 3
		`{"name":"string","age":"integer","height":"float"}`,
		`{"name":{"type":"string"},"age":{"type":"integer"},"height":{"type":"float"}}`,
	},
	{ // 4
		`{
			"big_object":{
				"big_value":"string",
				"medium_object_1":{
					"medium_v_1":"integer",
					"medium_v_2":"long",
					"small_object_1":{
						"what":"binary",
						"is":"geo_point",
						"this":"ip"
					}
				},
				"another_value":"geo_shape",
				"medium_object_2":{
					"does":"float",
					"this":"boolean",
					"scale":"date",
					"well":"geo_point",
					"enough":{"punk":"completion"}
				}
			}
		}`,
		`{"big_object":{
			"dynamic":"strict",
			"properties":{
				"big_value":{"type":"string"},
				"medium_object_1":{
                	"dynamic":"strict",
					"properties":{
						"medium_v_1":{"type":"integer"},
						"medium_v_2":{"type":"long"},
						"small_object_1":{
			                "dynamic":"strict",
							"properties":{
								"what":{"type":"binary"},
								"is":{"type":"geo_point"},
								"this":{"type":"ip"}
							}
						}
					}
				},
				"another_value":{"type":"geo_shape"},
				"medium_object_2":{
	                "dynamic":"strict",
					"properties":{
						"does":{"type":"float"},
						"this":{"type":"boolean"},
						"scale":{"type":"date"},
						"well":{"type":"geo_point"},
						"enough":{
			                "dynamic":"strict",
							"properties":{
								"punk":{"type":"completion"}
							}
						}
					}
				}
			}
		}
	}`,
	},
	{ // 5
		`{ "a":{"aa":"string","b":{"bb":"long","c":{"cc":"integer","d":{"dd":"short","e":{"ee":"byte","f":{"ff":"double","g":{"gg":"float","h":{"hh":"date","i":{"ii":"boolean","j":{"jj":"binary","k":{"kk":"geo_point","l":{"ll":"geo_shape","m":{"mm":"ip","n":{"nn":"completion"}}}}}}}}}}}}}} }`,
		`{"a":{"properties":{"aa":{"type":"string"},"b":{"properties":{"bb":{"type":"long"},"c":{"properties":{"cc":{"type":"integer"},"d":{"properties":{"dd":{"type":"short"},"e":{"properties":{"ee":{"type":"byte"},"f":{"properties":{"ff":{"type":"double"},"g":{"properties":{"gg":{"type":"float"},"h":{"properties":{"hh":{"type":"date"},"i":{"properties":{"ii":{"type":"boolean"},"j":{"properties":{"jj":{"type":"binary"},"k":{"properties":{"kk":{"type":"geo_point"},"l":{"properties":{"ll":{"type":"geo_shape"},"m":{"properties":{"mm":{"type":"ip"},"n":{"properties":{"nn":{"type":"completion"}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}`,
	},
}

//---------------------------------------------------------------------------

func (suite *MappingTester) xTest20Mappings() {
	t := suite.T()
	assert := assert.New(t)

	for idx, pair := range mappingTestData {
		input := pair[0]
		expectedOutput := pair[1]

		{
			log.Printf("======================================================")
			log.Printf("test %d", idx)

			inputTree := map[string]interface{}{}

			err := json.Unmarshal([]byte(input), &inputTree)
			assert.NoError(err)

			outputTree, err := visitNode(inputTree)
			assert.NoError(err)

			err = doVerification(expectedOutput, outputTree)
			assert.NoError(err, fmt.Sprintf("failed %d\n", idx))
		}
	}
}

func doVerification(str string, obj map[string]interface{}) error {

	var expected bytes.Buffer
	err := json.Compact(&expected, []byte(str))
	if err != nil {
		return err
	}

	actualBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var actual bytes.Buffer
	err = json.Compact(&expected, actualBytes)
	if err != nil {
		return err
	}

	s1 := expected.String()
	s2 := actual.String()
	//log.Printf("%d %d %d", len(s1), len(s2), strings.Compare(s1, s2))
	/*for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			log.Printf("%d: %s %s", i, string(s1[i]), string(s2[i]))
		}
	}*/
	if strings.Compare(s1, s2) != 0 {
		//printBytes(expected.Bytes())
		//printBytes(actual)
		return fmt.Errorf("nope")
	}
	return nil
}

func printBytes(data []byte) error {
	var obj interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		panic(err)
	}

	byts, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		panic(err)
	}

	log.Printf("%s", string(byts))

	return nil
}
