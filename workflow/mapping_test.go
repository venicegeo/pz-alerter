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
	"fmt"
	"log"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const loopAmount = 5000

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
		`{ "a":
			{"aa":"string",
			"b":{
				"bb":"long",
				"c":{
					"cc":"integer",
					"d":{
						"dd":"short",
						"e":{
							"ee":"byte",
							"f":{
								"ff":"double",
								"g":{
									"gg":"float",
									"h":{
										"hh":"date",
										"i":{
											"ii":"boolean",
											"j":{
												"jj":"binary",
													"k":{
														"kk":"geo_point",
														"l":{
															"ll":"geo_shape",
															"m":{
																"mm":"ip",
																"n":{
																	"nn":"completion"
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}`,
		`{"a":{
			"dynamic":"strict",
			"properties":{
				"aa":{"type":"string"},
				"b":{
					"dynamic":"strict",
					"properties":{
						"bb":{"type":"long"},
						"c":{
							"dynamic":"strict",
							"properties":{
								"cc":{"type":"integer"},
								"d":{
									"dynamic":"strict",
									"properties":{
										"dd":{"type":"short"},
										"e":{
											"dynamic":"strict",
											"properties":{
												"ee":{"type":"byte"},
												"f":{
													"dynamic":"strict",
													"properties":{
														"ff":{"type":"double"},
														"g":{
															"dynamic":"strict",
															"properties":{
																"gg":{"type":"float"},
																"h":{
																	"dynamic":"strict",
																	"properties":{
																		"hh":{"type":"date"},
																		"i":{
																			"dynamic":"strict",
																			"properties":{
																				"ii":{"type":"boolean"},
																				"j":{
																					"dynamic":"strict",
																					"properties":{
																						"jj":{"type":"binary"},
																						"k":{
																							"dynamic":"strict",
																							"properties":{
																								"kk":{"type":"geo_point"},
																								"l":{
																									"dynamic":"strict",
																									"properties":{
																										"ll":{"type":"geo_shape"},
																										"m":{
																											"dynamic":"strict",
																											"properties":{
																												"mm":{"type":"ip"},
																												"n":{
																													"dynamic":"strict",
																													"properties":{
																														"nn":{"type":"completion"}
																													}
																												}
																											}
																										}
																									}
																								}
																							}
																						}
																					}
																				}
																			}
																		}
																	}
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}}`,
	},
}

//---------------------------------------------------------------------------

func (suite *MappingTester) Test20Mappings() {
	t := suite.T()
	assert := assert.New(t)

	for i := 0; i < loopAmount; i++ {
		for idx, pair := range mappingTestData {
			input := pair[0]
			expectedOutput := pair[1]

			{
				inputTree := map[string]interface{}{}
				expectedOutputTree := map[string]interface{}{}

				temp, err := structStringToInterface(input)
				assert.NoError(err)
				inputTree = temp.(map[string]interface{})

				temp, err = structStringToInterface(expectedOutput)
				assert.NoError(err)
				expectedOutputTree = temp.(map[string]interface{})

				outputTree, err := visitNode(inputTree)
				assert.NoError(err)

				err = doVerification(expectedOutputTree, outputTree)
				assert.NoError(err, fmt.Sprintf("failed %d\n", idx))
			}
		}
	}
}

func (suite *MappingTester) Test21Mappings() {
	t := suite.T()
	assert := assert.New(t)

	for i := 0; i < loopAmount; i++ {
		for idx, pair := range mappingTestData {
			input := pair[0]
			expectedOutput := pair[1]

			{
				inputTree := map[string]interface{}{}
				expectedOutputTree := map[string]interface{}{}

				temp, err := structStringToInterface(input)
				assert.NoError(err)
				inputTree = temp.(map[string]interface{})

				temp, err = structStringToInterface(expectedOutput)
				assert.NoError(err)
				expectedOutputTree = temp.(map[string]interface{})

				outputTree, err := StringBuildMapping(inputTree)
				assert.NoError(err)

				err = doVerification(expectedOutputTree, outputTree)
				assert.NoError(err, fmt.Sprintf("failed %d\n", idx))
			}
		}
	}
}

func doVerification(expecte map[string]interface{}, actua map[string]interface{}) error {

	expected, err := structInterfaceToString(expecte)
	if err != nil {
		return err
	}
	expected = removeWhitespace(expected)
	actual, err := structInterfaceToString(actua)
	if err != nil {
		return err
	}
	actual = removeWhitespace(actual)
	if strings.Compare(expected, actual) != 0 {
		println(expected)
		println()
		println(actual)
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
