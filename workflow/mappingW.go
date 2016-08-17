package workflow

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

type objIdent struct {
	Index int
	Type  string
	Name  string
}
type objPair struct {
	OpenIndex   int
	ClosedIndex int
	Range       int
	Name        string
}

var closed, open = "closed", "open"

func pairsContain(pairs []objPair, index int) bool {
	for _, pair := range pairs {
		if index >= pair.OpenIndex && index <= pair.ClosedIndex {
			return true
		}
	}
	return false
}

func stringBuildMapping(obj map[string]interface{}) (map[string]interface{}, error) {
	str, err := piazza.StructInterfaceToString(obj)
	if err != nil {
		return nil, err
	}
	str = piazza.RemoveWhitespace(str)
	str = str[1 : len(str)-1]
	//-------------Find all open and closed quotes------------------------------
	quotes := []objPair{}
	idents := []objIdent{}
	qO, qCount := -1, 0
	oC, cC := 0, 0
	for i := 0; i < len(str); i++ {
		char := piazza.CharAt(str, i)
		if char == "\"" {
			if i != 0 {
				charBefore := piazza.CharAt(str, i-1)
				if charBefore != "\\" {
					qCount++
					if qO == -1 {
						qO = i
					} else {
						quotes = append(quotes, objPair{qO, i, 0, ""})
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
			idents = append(idents, objIdent{i, open, ""})
		} else if char == "}" && qO == -1 {
			cC++
			idents = append(idents, objIdent{i, closed, ""})
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
		char := piazza.CharAt(str, i)
		if char == "{" && !pairsContain(quotes, i) {
			oC++
			idents = append(idents, objIdent{i, open, ""})
		} else if char == "}" && !pairsContain(quotes, i) {
			cC++
			idents = append(idents, objIdent{i, closed, ""})
		}
	}

	//-------------Match brackets into pairs------------------------------------
	pairs := []objPair{}
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
		pairs = append(pairs, objPair{k, v, 0, ""})
	}
	//-------------Add properties after each open index-------------------------
	propertiesAddition := `"dynamic":"strict","properties":{`
	for i := 0; i < len(pairs); i++ {
		str = piazza.InsertString(str, propertiesAddition, pairs[i].OpenIndex+1)
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
		str = piazza.InsertString(str, `}`, pairs[i].ClosedIndex)
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
	quotes = []objPair{}
	qO, qCount = -1, 0
	for i := 0; i < len(str); i++ {
		char := piazza.CharAt(str, i)
		if char == "\"" {
			if i != 0 {
				charBefore := piazza.CharAt(str, i-1)
				if charBefore != "\\" {
					qCount++
					if qO == -1 {
						qO = i
					} else {
						quotes = append(quotes, objPair{qO, i, 0, ""})
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
		if piazza.CharAt(str, i) == "[" && !pairsContain(quotes, i) {
			squareBracketOpen = true
		} else if piazza.CharAt(str, i) == "]" && !pairsContain(quotes, i) {
			squareBracketOpen = false
		}
		if ((piazza.CharAt(str, i) == "}" || piazza.CharAt(str, i) == ",") && !squareBracketOpen) && !pairsContain(quotes, i) {
			temp += "\n" + piazza.CharAt(str, i) + "\n"
		} else if (piazza.CharAt(str, i) == "{" && !squareBracketOpen) && !pairsContain(quotes, i) {
			temp += piazza.CharAt(str, i) + "\n"
		} else {
			temp += piazza.CharAt(str, i)
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
	res, err := piazza.StructStringToInterface(temp)
	return res.(map[string]interface{}), err
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
