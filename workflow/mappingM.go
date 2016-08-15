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

import "fmt"

//type T map[string]T  // TODO

func BuildMappingTesting(input map[string]interface{}) (map[string]interface{}, error) {
	return visitNodeTesting(input)
}

func visitNodeTesting(inputObj map[string]interface{}) (map[string]interface{}, error) {

	outputObj := map[string]interface{}{}

	for k, v := range inputObj {
		//fmt.Printf("%s: %#v\n", k, v)
		switch t := v.(type) {

		case string:
			tree, err := handleLeafTesting(k, v)
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree

		case map[string]interface{}:
			tree, err := handleNonleafTesting(k, v)
			if err != nil {
				return nil, err
			}
			outputObj[k] = tree

		default:
			return nil, fmt.Errorf("unexpected type %T\n", t)
		}
	}

	return outputObj, nil
}

func handleNonleafTesting(k string, v interface{}) (map[string]interface{}, error) {
	//fmt.Printf("Handling nonleaf %s: %#v\n", k, v)

	subtree, err := visitNodeTesting(v.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	wrapperTree := map[string]interface{}{}
	wrapperTree["dynamic"] = "strict"
	wrapperTree["properties"] = subtree

	return wrapperTree, err
}

func handleLeafTesting(k string, v interface{}) (map[string]interface{}, error) {
	//fmt.Printf("Handling leaf %s: %s\n", k, v)

	tree := map[string]interface{}{}
	tree["type"] = v
	return tree, nil
}

