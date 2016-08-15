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
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type ObjIdent struct {
	Index int
	Type  string
	Name  string
}
type ObjPair struct {
	OpenIndex   int
	ClosedIndex int
	Range       int
	Name        string
}

var closed, open = "closed", "open"

func StringBuildMapping(obj map[string]interface{}) (map[string]interface{}, error) {
	str, err := structInterfaceToString(obj)
	if err != nil {
		return nil, err
	}
	str = removeWhitespace(str)
	str = str[1 : len(str)-1]
	//-------------Find all open and closed quotes------------------------------
	quotes := []ObjPair{}
	idents := []ObjIdent{}
	qO, qCount := -1, 0
	oC, cC := 0, 0
	for i := 0; i < len(str); i++ {
		char := charAt(str, i)
		if char == "\"" {
			if i != 0 {
				charBefore := charAt(str, i-1)
				if charBefore != "\\" {
					qCount++
					if qO == -1 {
						qO = i
					} else {
						quotes = append(quotes, ObjPair{qO, i, 0, ""})
						qO = -1
					}
				}
			} else {
				qCount++
				if qO == -1 {
					qO = i
				}
			}
		} else if char == "{" && qO == -1 {
			oC++
			idents = append(idents, ObjIdent{i, open, ""})
		} else if char == "}" && qO == -1 {
			cC++
			idents = append(idents, ObjIdent{i, closed, ""})
		}
	}
	if len(quotes)*2 != qCount {
		return nil, fmt.Errorf("Not enough quotes: %d %d*2", qCount, len(quotes))
	}
	if oC != cC {
		return nil, fmt.Errorf("Not correct brackets: %d != %d", oC, cC)
	}
	//-------------Find all open and closed brackets----------------------------
	for i := 0; i < len(str); i++ {
		char := charAt(str, i)
		if char == "{" && !pairsContain(quotes, i) {
			oC++
			idents = append(idents, ObjIdent{i, open, ""})
		} else if char == "}" && !pairsContain(quotes, i) {
			cC++
			idents = append(idents, ObjIdent{i, closed, ""})
		}
	}

	//-------------Match brackets into pairs------------------------------------
	pairs := []ObjPair{}
	pairMap := map[int]int{}
	for len(idents) > 0 {
		for i := 0; i < len(idents)-1; i++ {
			a := idents[i]
			b := idents[i+1]
			if a.Type == open && b.Type == closed {
				pairMap[a.Index] = b.Index
				idents = append(idents[:i], idents[i+1:]...)
				idents = append(idents[:i], idents[i+1:]...)
				break
			}
		}
	}
	//-------------Sort pairs based off open bracket index----------------------
	keys := []int{}
	for k, _ := range pairMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		v := pairMap[k]
		pairs = append(pairs, ObjPair{k, v, 0, ""})
	}
	//-------------Add properties after each open index-------------------------
	propertiesAddition := `"dynamic":"strict","properties":{`
	for i := 0; i < len(pairs); i++ {
		str = insertString(str, propertiesAddition, pairs[i].OpenIndex+1)
		for j := i + 1; j < len(pairs); j++ {
			pairs[j].OpenIndex += len(propertiesAddition)
		}
		for j := 0; j < len(pairs); j++ {
			if pairs[j].ClosedIndex >= pairs[i].OpenIndex {
				pairs[j].ClosedIndex += len(propertiesAddition)
			}
		}
	}
	//-------------Add a closed bracket at each closed index--------------------
	for i := 0; i < len(pairs); i++ {
		str = insertString(str, `}`, pairs[i].ClosedIndex)
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].OpenIndex >= pairs[i].ClosedIndex {
				pairs[j].OpenIndex++
			}
		}
		for j := 0; j < len(pairs); j++ {
			if pairs[j].ClosedIndex >= pairs[i].ClosedIndex {
				pairs[j].ClosedIndex++
			}
		}
	}
	//-------------Find all open and closed quotes------------------------------
	quotes = []ObjPair{}
	qO, qCount = -1, 0
	for i := 0; i < len(str); i++ {
		char := charAt(str, i)
		if char == "\"" {
			if i != 0 {
				charBefore := charAt(str, i-1)
				if charBefore != "\\" {
					qCount++
					if qO == -1 {
						qO = i
					} else {
						quotes = append(quotes, ObjPair{qO, i, 0, ""})
						qO = -1
					}
				}
			} else {
				qCount++
				if qO == -1 {
					qO = i
				}
			}
		}
	}
	if len(quotes)*2 != qCount {
		return nil, fmt.Errorf("Not enough quotes: %s %s*2", qCount, len(quotes))
	}
	//-------------Seperate pieces of mapping onto seperate lines---------------
	temp := ""
	squareBracketOpen := false
	for i := 0; i < len(str); i++ {
		if charAt(str, i) == "[" && !pairsContain(quotes, i) {
			squareBracketOpen = true
		} else if charAt(str, i) == "]" && !pairsContain(quotes, i) {
			squareBracketOpen = false
		}
		if ((charAt(str, i) == "}" || charAt(str, i) == ",") && !squareBracketOpen) && !pairsContain(quotes, i) {
			temp += "\n" + charAt(str, i) + "\n"
		} else if (charAt(str, i) == "{" && !squareBracketOpen) && !pairsContain(quotes, i) {
			temp += charAt(str, i) + "\n"
		} else {
			temp += charAt(str, i)
		}
	}
	lines := strings.Split(temp, "\n")
	//-------------Format individual values-------------------------------------
	fixed := []string{}
	j := 0
	for _, line := range lines {
		if strings.HasPrefix(line, `"`) && !strings.HasSuffix(line, "{") {
			if strings.Contains(line, `":`) {
				toSplit := []int{}
				for i := 1; i < len(line); i++ {
					test := line[i-1 : i+1]
					if test == `":` {
						isGood := pairsContain(quotes, i+j-1)
						if isGood {
							toSplit = append(toSplit, i)
						}
					}
				}
				if len(toSplit) > 1 {
					return nil, errors.New("BAD CODE")
				}
				if len(toSplit) == 1 {
					formatedLine, err := formatKeyValue(line, toSplit[0])
					if err != nil {
						return nil, err
					}
					fixed = append(fixed, formatedLine)
				} else {
					fixed = append(fixed, line)
				}
			} else {
				fixed = append(fixed, line)
			}
		} else {
			fixed = append(fixed, line)
		}
		j += len(line)
	}
	temp = ""
	for _, line := range fixed {
		temp += line
	}
	temp = "{" + temp + "}"
	res, err := structStringToInterface(temp)
	return res.(map[string]interface{}), err
}

func pairsContain(pairs []ObjPair, index int) bool {
	for _, pair := range pairs {
		if index >= pair.OpenIndex && index <= pair.ClosedIndex {
			return true
		}
	}
	return false
}
func formatKeyValue(str string, whereToSplit int) (string, error) {
	parts := []string{str[:whereToSplit], str[whereToSplit+1:]}
	value := strings.Replace(parts[1], "\"", "", -1)
	key := strings.Replace(parts[0], "\"", "", -1)
	dynamic := key == "dynamic"
	if dynamic {
		if value != "strict" && value != "true" && value != "false" {
			return "", fmt.Errorf(" %s was not recognized as a valid dynamic type", value)
		}
		return str, nil
	}
	res := fmt.Sprintf(`%s:{"type":%s}`, parts[0], parts[1])
	return res, nil
}

func structStringToInterface(stru string) (interface{}, error) {
	data := []byte(stru)
	source := (*json.RawMessage)(&data)
	var res interface{}
	err := json.Unmarshal(*source, &res)
	return res, err
}
func structInterfaceToString(stru interface{}) (string, error) {
	data, err := json.MarshalIndent(stru, " ", "   ")
	return string(data), err
}
func charAt(str string, index int) string {
	return str[index : index+1]
}
func removeWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
func insertString(str, insert string, index int) string {
	return str[:index] + insert + str[index:]
}

