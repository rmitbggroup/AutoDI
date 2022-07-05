package Explain

import (
	"fmt"
	"sort"
	"strings"
)

func ObtainJoinPath(diff *DiffItem) {
	println("P1: " + diff.N1.ObtainJoinPath())
	println("P2: " + diff.N2.ObtainJoinPath())
}

func PrintDiffInfo(diff *DiffItem) {
	switch diff.DiffMessage {
		case "JODiff":
			ObtainJoinPath(diff)
		case "ImpDiff":
			//println("ImpDiff")
			println(diff.N1.PhysicalSig.GetNodeImp())
			println(diff.N2.PhysicalSig.GetNodeImp())
		case "DADiff":
			//println("DADiff")
			println(diff.N1.PhysicalSig.GetNodeImp())
			println(diff.N2.PhysicalSig.GetNodeImp())
		case "OODiff":
			println("OODiff")
		default:
			panic("Not Support This Diff")
	}
}

// Currently, support JOB and TPC-H (without q2)
func buildDiffGraph(diffs []*DiffItem) *DiffItem {
	sort.SliceStable(diffs, func(i, j int) bool {
		return diffs[i].DiffLevel < diffs[j].DiffLevel
	})
	rootDiff := &DiffItem{}
	processedDiffIdx := 0
	if diffs[0].DiffLevel == diffs[1].DiffLevel {
		rootDiff.LChild = diffs[0]
		rootDiff.RChild = diffs[1]
		processedDiffIdx = 1
	} else {
		rootDiff.LChild = diffs[0]
	}

	noFinished := true
	lastAddedRoot := rootDiff
	for noFinished {
		remained := len(diffs)-(processedDiffIdx+1)
		if remained > 1 {
			if diffs[processedDiffIdx+1].DiffLevel == diffs[processedDiffIdx+2].DiffLevel {
				isInLeft := diffs[processedDiffIdx+1].N1.LogicalSig.ISIncluded(lastAddedRoot.LChild.N1.LogicalSig)
				isInRight := false
				if lastAddedRoot.RChild != nil {
					isInRight = diffs[processedDiffIdx+1].N1.LogicalSig.ISIncluded(lastAddedRoot.RChild.N1.LogicalSig)
				}
				if isInLeft == isInRight {
					panic("They are in both subtree...")
				}

				if isInLeft {
					lastAddedRoot.LChild.LChild =  diffs[processedDiffIdx+1]
					lastAddedRoot.LChild.RChild =  diffs[processedDiffIdx+2]
					lastAddedRoot = lastAddedRoot.LChild
				} else {
					lastAddedRoot.RChild.LChild =  diffs[processedDiffIdx+1]
					lastAddedRoot.RChild.RChild =  diffs[processedDiffIdx+2]
					lastAddedRoot = lastAddedRoot.RChild
				}
				processedDiffIdx += 2
			} else {
				if processedDiffIdx == 6 {
					print(0)
				}
				isInLeft := diffs[processedDiffIdx+1].N1.LogicalSig.ISIncluded(lastAddedRoot.LChild.N1.LogicalSig)
				isInRight := false
				if lastAddedRoot.RChild != nil {
					isInRight = diffs[processedDiffIdx+1].N1.LogicalSig.ISIncluded(lastAddedRoot.RChild.N1.LogicalSig)
				}
				if isInLeft == isInRight {
					if isInLeft {
						panic("They are in both subtree...")
					} else {
						panic("This diff is not included...")
					}

				}
				if isInLeft {
					lastAddedRoot.LChild.LChild =  diffs[processedDiffIdx+1]
					lastAddedRoot = lastAddedRoot.LChild
				} else {
					lastAddedRoot.RChild.LChild =  diffs[processedDiffIdx+1]
					lastAddedRoot = lastAddedRoot.RChild
				}
				processedDiffIdx += 1
			}
		} else if remained == 1 {
			isInLeft := diffs[processedDiffIdx+1].N1.LogicalSig.ISIncluded(lastAddedRoot.LChild.N1.LogicalSig)
			isInRight := false
			if lastAddedRoot.RChild != nil {
				isInRight = diffs[processedDiffIdx+1].N1.LogicalSig.ISIncluded(lastAddedRoot.RChild.N1.LogicalSig)
			}
			if isInLeft == isInRight {
				if isInLeft {
					panic("They are in both subtree...")
				} else {
					panic("This diff is not included...")
				}

			}
			if isInLeft {
				lastAddedRoot.LChild.LChild =  diffs[processedDiffIdx+1]
				lastAddedRoot = lastAddedRoot.LChild
			} else {
				lastAddedRoot.RChild.LChild =  diffs[processedDiffIdx+1]
				lastAddedRoot = lastAddedRoot.RChild
			}
			processedDiffIdx += 1
			noFinished = false
		} else {
			noFinished = false
		}

	}
	return rootDiff
}

func PrintDiff(queryID string, diffs []*DiffItem, diffRatio float64, overallDiff float64) {
	if len(diffs) > 0 {
		println("=======> Query: " + queryID + " <========")
	}

	mainIdx := 0

	diffThreshold := overallDiff*diffRatio
	curMain := new(DiffItem)
	*curMain = *diffs[0]

	for i, diff := range diffs[1:] {
		if ParseTime4NS(diff.N1.Runtime) - ParseTime4NS(diff.N2.Runtime) > diffThreshold && diff.DiffLevel > curMain.DiffLevel {
			*curMain = *diff
			mainIdx = i+1
		}
	}

	// stage2: output reason behind other diffs.
	remainDiffs := make([]*DiffItem, 0, 2)
	for i, diff := range diffs {
		if i != mainIdx {
			remainDiffs = append(remainDiffs, diff)
		}
	}

	diff := diffs[mainIdx]

	if diff.DiffMessage == "JODiff" {
		println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\n", diff.DiffMessage,
			diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
			diff.N1.ObtainJoinPath(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.ObtainJoinPath()))
	} else {
		println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\n", diff.DiffMessage,
			diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
			diff.N1.PhysicalSig.GetNodeImp(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.PhysicalSig.GetNodeImp()))
	}

	for _, diff := range remainDiffs {
			if diff.DiffMessage == "JODiff" {
				println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\n", diff.DiffMessage,
					diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
					diff.N1.ObtainJoinPath(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.ObtainJoinPath()))
			} else {
				println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\n", diff.DiffMessage,
					diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
					diff.N1.PhysicalSig.GetNodeImp(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.PhysicalSig.GetNodeImp()))
			}
		}

}

func PrintReport(queryID string, diffs []*DiffItem, mainIdx int, report []string, tc1 float64, tc2 float64,
	tr1 string, tr2 string) {
	if len(diffs) > 0 {
		println("=======> Query: " + queryID + " <========")
		println(fmt.Sprintf("%d Diffs Have Been Found Between Two Plans.\nP1's cost is %f and runtime is %s.\nP2's cost is %f and runtime is %s.\n",
			len(diffs),tc1, tr1, tc2, tr2 ))
	}
	j := 0
	if len(diffs) > 1 {
		println("------- Main Diff Report -------")
		diff := diffs[mainIdx]
		if diff.DiffMessage == "JODiff" {
			println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\nPossible Reasons:%s\n", diff.DiffMessage,
				diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
				diff.N1.ObtainJoinPath(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.ObtainJoinPath(), report[j]))
		} else {
			println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\nPossible Reasons:%s\n", diff.DiffMessage,
				diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
				diff.N1.PhysicalSig.GetNodeImp(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.PhysicalSig.GetNodeImp(), report[j]))
		}
		j++
		println("------- Other Diffs Report -------")

	}
	for i, diff := range diffs {
		if i != mainIdx {
			if diff.DiffMessage == "JODiff" {
				println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\nPossible Reasons:%s\n", diff.DiffMessage,
					diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
					diff.N1.ObtainJoinPath(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.ObtainJoinPath(), report[j]))
			} else {
				println(fmt.Sprintf("Diff Type: %s, Node ID:(%s, %s), Level: %d\n%s with cost %f and runtime %s \n%s with cost %f and runtime %s \n\t%s:%s\n\t%s:%s\nPossible Reasons:%s\n", diff.DiffMessage,
					diff.N1.OP+"_"+diff.N1.ID, diff.N2.OP+"_"+diff.N2.ID, diff.DiffLevel, diff.N1.OP+"_"+diff.N1.ID, diff.N1.Cost, diff.N1.Runtime, diff.N2.OP+"_"+diff.N2.ID, diff.N2.Cost, diff.N2.Runtime, diff.N1.OP+"_"+diff.N1.ID,
					diff.N1.PhysicalSig.GetNodeImp(), diff.N2.OP+"_"+diff.N2.ID, diff.N2.PhysicalSig.GetNodeImp(), report[j]))
			}
			j++
		}
	}
	if j < len(report) {
		println(report[j])
		j++
	}
}

func Inference(queryID string, diffs []*DiffItem, diffRatio float64, overallDiff float64) (int, []string) {
	rs := make([]string, 0, 2)
	mainIdx := 0
	if len(diffs) > 1 {
		// stage-1: find the most important reason, i.e., the lowest one with different runtime
		diffThreshold := overallDiff*diffRatio
		curMain := new(DiffItem)
		*curMain = *diffs[0]

		for i, diff := range diffs[1:] {
			if ParseTime4NS(diff.N1.Runtime) - ParseTime4NS(diff.N2.Runtime) > diffThreshold && diff.DiffLevel > curMain.DiffLevel {
				*curMain = *diff
				mainIdx = i+1
			}
		}
		rs = append(rs, inference(queryID, []*DiffItem{curMain})[0])

		// stage2: output reason behind other diffs.
		remainDiffs := make([]*DiffItem, 0, 2)
		for i, diff := range diffs {
			if i != mainIdx {
				remainDiffs = append(remainDiffs, diff)
			}
		}
		rs = append(rs, inference(queryID, remainDiffs)...)

		// stage3: analyze the diff relationship.
		woJODiffs := make([]*DiffItem, 0, 2)
		for _, diff := range diffs {
			if diff.DiffMessage != "JODiff" {
				woJODiffs = append(woJODiffs, diff)
			}
		}
		//diffGraph := buildDiffGraph(woJODiffs)
		//// 1) JoinType influences data access path.
		//rs = append(rs, obtainTDRelationship(diffGraph)...)
		// 2) data access path influences join.

	} else {
		rs = inference(queryID, diffs)
	}
	return mainIdx, rs
}

func obtainTDRelationship(diffNode *DiffItem) []string {
	rs := make([]string, 0, 1)
	if diffNode.DiffMessage == "ImpDiff" {
		if diffNode.N1.NodeType == "Join" && (!diffNode.N1.IsMergeHashJoin() || !diffNode.N2.IsMergeHashJoin()){
			if diffNode.LChild != nil && diffNode.LChild.DiffMessage == "DADiff" {
				if !diffNode.N1.IsMergeHashJoin() {
					isDecide := strings.Index(diffNode.LChild.N1.OperatorInfo, "decided")
					if isDecide > -1 {
						rs = append(rs, fmt.Sprintf("Reader With ID %s In P1 IS Decided By The Upper Join.", diffNode.LChild.N1.ID))
					}
				}
				if !diffNode.N2.IsMergeHashJoin() {
					isDecide := strings.Index(diffNode.LChild.N2.OperatorInfo, "decided")
					if isDecide > -1 {
						rs = append(rs, fmt.Sprintf("Reader With ID %s In P2 IS Decided By The Upper Join.", diffNode.LChild.N2.ID))
					}
				}
			}
			if diffNode.RChild != nil && diffNode.RChild.DiffMessage == "DADiff" {
				if !diffNode.N1.IsMergeHashJoin() {
					isDecide := strings.Index(diffNode.LChild.N1.OperatorInfo, "decided")
					if isDecide > -1 {
						rs = append(rs, fmt.Sprintf("Reader With ID %s In P1 IS Decided By The Upper Join.", diffNode.LChild.N1.ID))
					}
				}
				if !diffNode.N2.IsMergeHashJoin() {
					isDecide := strings.Index(diffNode.LChild.N1.OperatorInfo, "decided")
					if isDecide > -1 {
						rs = append(rs, fmt.Sprintf("Readers With ID %s In P1 And With ID %s In P2 Are Decided By The Upper Join.", diffNode.LChild.N1.ID, diffNode.LChild.N2.ID))
					}
				}
			}
		}
	}
	if diffNode.LChild != nil {
		rs = append(rs, obtainTDRelationship(diffNode.LChild)...)
	}
	if diffNode.RChild != nil {
		rs = append(rs, obtainTDRelationship(diffNode.RChild)...)
	}
	return rs
}

func inference(queryID string, diffs []*DiffItem) []string {
	rs := make([]string, 0, 2)
	for i := 0; i < len(diffs); i++ {
		rule, ok := RuleMap[diffs[i].DiffMessage]
		if ok {
			iRs := make([]string, 0, 2)
			for j := 0; j < len(rule.Actions); j++ {
				if rule.ActionFlags[j] == byte(1) {
					iRs = append(iRs, rule.ActionName[j]+"#"+rule.Actions[j](diffs[i].N1, diffs[i].N2, 1))
				}
			}
			rs = append(rs, strings.Join(iRs, "\n"))
		} else {
			rs = append(rs, "No Rule For This Diff")
		}
	}
	return rs
}