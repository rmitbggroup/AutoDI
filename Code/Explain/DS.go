package Explain

import (
	"ExplainPlan/Plan"
	"strconv"
)

var DiffConfCase = 1
var DiffOptCase = 2

type DiffItem struct {
	N1 *Plan.Node
	N2 *Plan.Node
	//Weight float64
	DiffMessage string
	DiffLevel int

	//Used to construct the diff graph
	LChild *DiffItem
	RChild *DiffItem
}

func SelfMinInt(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func SelfMaxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

type Rule struct {
	DiffName string
	//DiffArgs []string
	ActionName []string
	Actions []func(n1 *Plan.Node, n2 *Plan.Node, inputType int) string
	ActionFlags []byte
}

func ParseTime4NS(runtime string) float64 {
	var newRuntime = float64(0)
	unit := runtime[len(runtime)-2:]
	switch unit {
	case "us":
		_newRuntime, _ := strconv.ParseFloat(runtime[:len(runtime)-2], 64)
		newRuntime = _newRuntime*float64(1000)
	case "ms":
		_newRuntime, _ := strconv.ParseFloat(runtime[:len(runtime)-2], 64)
		newRuntime = _newRuntime*float64(1000000)
	case "ns":
		newRuntime, _ = strconv.ParseFloat(runtime[:len(runtime)-2], 64)
	default:
		switch unit[1] {
		case 'h':
			_newRuntime, _ := strconv.ParseFloat(runtime[:len(runtime)-1], 64)
			newRuntime = _newRuntime*float64(3600000000000)
		case 'm':
			_newRuntime, _ := strconv.ParseFloat(runtime[:len(runtime)-1], 64)
			newRuntime = _newRuntime*float64(60000000000)
		case 's':
			_newRuntime, _ := strconv.ParseFloat(runtime[:len(runtime)-1], 64)
			newRuntime = _newRuntime*float64(1000000000)
		}
	}
	return newRuntime
}
type ByDiffRunTime []*DiffItem
func (s ByDiffRunTime) Len() int { return len(s) }
func (s ByDiffRunTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
// We rank the diffs from high diff time to low diff time.
func (s ByDiffRunTime) Less(i, j int) bool {
	diffI := ParseTime4NS(s[i].N2.Runtime) - ParseTime4NS(s[i].N1.Runtime)
	diffJ := ParseTime4NS(s[j].N2.Runtime) - ParseTime4NS(s[j].N1.Runtime)
	return diffI < diffJ
}