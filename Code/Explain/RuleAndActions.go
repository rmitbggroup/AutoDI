package Explain

import (
	"ExplainPlan/Plan"
	"fmt"
	"math"
)

var DARule = Rule {
	DiffName: "DADiff",
	ActionName: []string {
		"CE&CM Check",
	},
	Actions: []func(*Plan.Node, *Plan.Node, int) string {_DADiff},
	ActionFlags: []byte{1,1},
}

var ImpDiffRule = Rule {
	DiffName: "ImpDiff",
	ActionName: []string {
		"CE Check",
		"CM Check",
	},
	Actions: []func(*Plan.Node, *Plan.Node, int) string {_ce4ImpDiff, _cm4ImpDiff},
	ActionFlags: []byte{1,1},
}

var JODiffRule = Rule {
	DiffName: "JODiff",
	ActionName: []string {
		"JO Check",
	},
	Actions: []func(*Plan.Node, *Plan.Node, int) string {JODiff},
	ActionFlags: []byte{1},
}

var RuleMap = map[string] Rule {
	"DADiff": DARule,
	"ImpDiff": ImpDiffRule,
	"JODiff": JODiffRule,
}

func isN1CloserToTrueCard(n1 *Plan.Node, n2 *Plan.Node) bool {
	c1 := math.Abs(n1.EstRows-n1.ActRows) // TODO: Runtime; 98% Threshold
	c2 := math.Abs(n2.EstRows-n2.ActRows)
	return  c1 < c2
}

func isJN1CloserToTrueCard(n1 *Plan.JNode, n2 *Plan.JNode) bool {
	c1 := math.Abs(n1.EstRow-n1.ActRow)
	c2 := math.Abs(n2.EstRow-n2.ActRow)
	return  c1 < c2
}


func _ce4ImpDiff(n1 *Plan.Node, n2 *Plan.Node, intputType int) string {
	var rs = "No Report On CE For This Diff"
	n1EstRow, n2EstRow, n1LCEstRow, n1RCEstRow, n2LCEstRow, n2RCEstRow := obtainEstRow(n1, n2)
	isSameRoot := n1EstRow == n2EstRow
	isSameLeft := n1LCEstRow == n2LCEstRow
	isSameRight := n1RCEstRow == n2RCEstRow

	n1Faster := isN1RunFaster(n1, n2)

	if isSameRoot && isSameLeft && isSameRight {
		rs = "[Report] It Is Not The Problem In Cardinality Estimation For This Diff." // both in case 1 and case 2
		return rs
	}

	if !isSameRoot && isSameLeft && isSameRight {
		// We suppose a more accurate card will get a better plan.
		if n1Faster == isN1CloserToTrueCard(n1, n2) {
			if n1Faster {
				rs = "[Report] The Cardinality Estimation In P2 Is More Worse Than P1 For This Diff."
			} else {
				rs = "[Report] The Cardinality Estimation In P1 Is More Worse Than P1 For This Diff."
			}
		} else {
			// Cost model may has problem for this in P2.
			if isN1CloserToTrueCard(n1, n2) {
				rs = "[Report] The Cardinality Estimation In P1 Is Better While It Shows More Worse Performance."
			} else {
				rs = "[Report] The Cardinality Estimation In P1 Has Problem While It Shows Better Performance."
			}
		}
	}

	if !isSameRoot && (!isSameLeft || !isSameRight) {
		// This may cause the cascade checks on children's cardinalities.
		rs = "[Report] The Children's Cardinalities Are Different, Which May Cause Card Diff In This Node."
		return rs
	}

	if isSameRoot && (!isSameLeft || !isSameRight) {
		rs = "[Report] The Children's Cardinalities Are Different, Which May Cause Cost Diff In This Node."
	}

	return rs
}

// We fix the cost of the children here for index[merge/hash]join and mergejoin.
func obtainCost(n1 *Plan.Node, n2 *Plan.Node) (float64, float64, float64, float64, float64, float64) {
	n1EstCost := n1.Cost
	n2EstCost := n2.Cost

	var n1LCEstCost = 0.0
	var n2LCEstCost = 0.0
	var n1RCEstCost = 0.0
	var n2RCEstCost = 0.0

	if n1.LChild != nil {
		n1LCEstCost = n1.LChild.Cost
		if n1.LChild.RightEstRows != -1.0 {
			n1LCEstCost = n1.LChild.Cost * n1.LChild.RightEstRows
		}
	}

	if n2.LChild != nil {
		n2LCEstCost = n2.LChild.Cost
		if n2.LChild.RightEstRows != -1.0 {
			n2LCEstCost = n2.LChild.Cost * n2.LChild.RightEstRows
		}
	}

	if n1.RChild != nil {
		n1RCEstCost = n1.RChild.Cost
		if n1.RChild.RightEstRows != -1.0 {
			n1RCEstCost = n1.RChild.Cost * n1.RChild.RightEstRows
		}
	}

	if n2.RChild != nil {
		n2RCEstCost = n2.RChild.Cost
		if n2.RChild.RightEstRows != -1.0 {
			n2RCEstCost = n2.RChild.Cost * n2.RChild.RightEstRows
		}
	}


	return n1EstCost, n2EstCost, n1LCEstCost, n1RCEstCost, n2LCEstCost, n2RCEstCost
}

// Possible the ce of the right child in the index*join.
// However, there is little possibility to choose a bad index.
// We still return that estRow.
func obtainEstRow(n1 *Plan.Node, n2 *Plan.Node) (float64, float64, float64, float64, float64, float64) {
	n1EstRow := n1.EstRows
	n2EstRow := n2.EstRows

	var n1LCEstRow = 0.0
	var n2LCEstRow = 0.0
	var n1RCEstRow = 0.0
	var n2RCEstRow = 0.0
	if n1.LChild != nil {
		n1LCEstRow = n1.LChild.EstRows
	}
	if n1.RChild != nil {
		n1RCEstRow = n1.RChild.EstRows
	}
	if n2.LChild != nil {
		n1LCEstRow = n2.LChild.EstRows
	}
	if n2.RChild != nil {
		n1RCEstRow = n2.RChild.EstRows
	}

	return n1EstRow, n2EstRow, n1LCEstRow, n1RCEstRow, n2LCEstRow, n2RCEstRow
}

func isN1RunFaster(n1, n2 *Plan.Node) bool {
	return ParseTime4NS(n1.Runtime) < ParseTime4NS(n2.Runtime)
}

func _cm4ImpDiff(n1 *Plan.Node, n2 *Plan.Node, inputType int) string {
	var rs = "No Report On CM For This Diff"

	n1EstRow, n2EstRow, n1LCEstRow, n1RCEstRow, n2LCEstRow, n2RCEstRow := obtainEstRow(n1, n2)
	isSameRoot := n1EstRow == n2EstRow
	isSameLeft := n1LCEstRow == n2LCEstRow
	isSameRight := n1RCEstRow == n2RCEstRow

	n1Cost, n2Cost, n1LCCost, n1RCCost, n2LCCost, n2RCCost := obtainCost(n1, n2)
	remainedCost1 := n1Cost - (n1LCCost + n1RCCost)
	remainedCost2 := n2Cost - (n2LCCost + n2RCCost)
	n1Faster := isN1RunFaster(n1, n2)
	n1FasterExpected := n1Cost < n2Cost
	n1NodeFasterExpected := remainedCost1 < remainedCost2
	if isSameRoot && isSameLeft && isSameRight {
		if n1Faster == n1FasterExpected { // rank the same
			if n1NodeFasterExpected == n1Faster {
				rs = "[Report] With Same Cardinality Estimation & Same Rank In Cost and Runtime, This Can Happen With " +
					"(1) User Defined Hints, (2) Different Search Space; (3) New Operator Introduced"
			} else {
				rs = "[Report] The Cost Model Is Not Accurate At Least At This Node."
				if n1Faster {
					rs += " P2 Has Been Underestimated or P1 Has Been Overestimated."
				} else {
					rs += " P1 Has Been Underestimated or P2 Has Been Overestimated."
				}
			}
		} else { // rank different
			rs = "[Report] Cost Mode Does Not Reflect The Runtime."
			if n1NodeFasterExpected == n1Faster {
				rs += " Inaccurate Cost Is In The Children. Need to Further Check."
			} else {
				rs = "[Report] The Cost Model Is Not Accurate At Least At This Node."
				if n1Faster {
					rs += " P2 Has Been Underestimated or P1 Has Been Overestimated."
				} else {
					rs += " P1 Has Been Underestimated or P2 Has Been Overestimated."
				}
			}
		}
		return rs
	}

	if !isSameRoot {
		rs = "[Report] May Be No Problem On CM And Need To First Fix Diff In CE."
		return rs
	}

	if !isSameLeft || !isSameRight {
		if n1Faster == n1FasterExpected {
			if n1Faster == n1NodeFasterExpected {
				rs = "[Report] With Same Cardinality Estimation & Same Rank In Cost and Runtime, This Can Happen With " +
					"(1) User Defined Hints, (2) Different Search Space; (3) New Operator Introduced"
			} else {
				rs = "[Report] The Cost Model Is Not Accurate At Least At This Node."
				if n1Faster {
					rs += " P2 Has Been Underestimated or P1 Has Been Overestimated."
				} else {
					rs += " P1 Has Been Underestimated or P2 Has Been Overestimated."
				}
			}
		} else {
			if n1Faster == n1NodeFasterExpected {
				rs = "[Report] The Children Cost Is Not Accurate, Which May Cause This Diff."
			} else {
				rs = "[Report] The Cost Model Is Not Accurate At Least At This Node."
				if n1Faster {
					rs += " P2 Has Been Underestimated or P1 Has Been Overestimated."
				} else {
					rs += " P1 Has Been Underestimated or P2 Has Been Overestimated."
				}
			}
		}
	}

	return rs
}

func _DADiff(n1 *Plan.Node, n2 *Plan.Node, inputType int) string {
	rs := "No Report For This Diff"

	n1Faster := isN1RunFaster(n1,n2)
	n1ExpectedFaster := n1.Cost < n2.Cost

	//idx vs idx-CE
	// TODO: combine two cases & runtime
	if n1.OP == "IndexLookUp" && n2.OP == "IndexLookUp" {
		idx1, idxCond1 := Plan.GetRangeCondition(n1.RChild)
		idx2, idxCond2 := Plan.GetRangeCondition(n2.RChild)
		if isN1CloserToTrueCard(n1.RChild, n2.RChild) != n1Faster {
			if n1Faster {
				rs = fmt.Sprintf("[Report] Bad Cardinality Estimation On Index: %s, Condition: %s, Q-Error: %f",
					idx1, idxCond1, math.Abs(n1.EstRows-n1.ActRows)/n1.ActRows)
			} else {
				rs = fmt.Sprintf("[Report] Bad Cardinality Estimation On Index: %s, Condition: %s, Q-Error: %f",
					idx2, idxCond2, math.Abs(n2.EstRows-n2.ActRows)/n2.ActRows)
			}
		} else {
			if n1ExpectedFaster == n1Faster {
				//if n1Faster {
				//	rs = "[Report] With Same Rank In Cost, CE, and Runtime, This Can Happen In P2 With " +
				//		"(1) User Defined Hints, (2) Different Index Set"
				//} else {
					rs = "[Report] With Same Rank In Cost, CE, and Runtime, This Can Happen In P1 With " +
						"(1) User Defined Hints, (2) Different Index Set"
				//}
			} else {
				rs = "[Report] The Cost Model Is Not Accurate At Least At This Node."
				if n1Faster {
					rs += " P2 Has Been Underestimated or P1 Has Been Overestimated."
				} else {
					rs += " P1 Has Been Underestimated or P2 Has Been Overestimated."
				}
			}
		}
		return rs
	}


	//idx vs full-Cost Model
	if n1.OP != n2.OP {
		if isN1CloserToTrueCard(n1, n2) != n1Faster {
			if n1Faster {
				if n1.OP == "TableReader" { // TableRangeScan/
					rs = "[Report] Bad Cardinality Estimation On P1"
				} else {
					idx, idxCond := Plan.GetRangeCondition(n1.RChild)
					rs = fmt.Sprintf("[Report] Bad Cardinality Estimation On Index: %s, Condition: %s, Q-Error: %f",
						idx, idxCond, math.Abs(n1.EstRows-n1.ActRows)/n1.ActRows)
				}

			} else {
				if n1.OP == "TableReader" {
					idx, idxCond := Plan.GetRangeCondition(n2.RChild)
					rs = fmt.Sprintf("[Report] Bad Cardinality Estimation On Index: %s, Condition: %s, Q-Error: %f",
						idx, idxCond, math.Abs(n2.EstRows-n2.ActRows)/n2.ActRows)

				} else {
					rs = "[Report] Bad Cardinality Estimation On P1"
				}
			}
		} else {
			if n1ExpectedFaster == n1Faster {
				if n1Faster {
					rs = "[Report] With Same Rank In Cost, CE, and Runtime, This Can Happen In P2 With " +
						"(1) User Defined Hints, (2) Different Index Set"
				} else {
					rs = "[Report] With Same Rank In Cost, CE, and Runtime, This Can Happen In P1 With " +
						"(1) User Defined Hints, (2) Different Index Set"
				}
			} else {
				rs = "[Report] The Cost Model Is Not Accurate At Least At This Node."
				if n1Faster {
					rs += " P2 Has Been Underestimated or P1 Has Been Overestimated."
				} else {
					rs += " P1 Has Been Underestimated or P2 Has Been Overestimated."
				}
			}
		}
	}

	return rs
}

// Selection, Agg
func OODiff(n1 *Plan.Node, n2 *Plan.Node, inputType int) string {
	return ""
}

func handleFirstDiffJoin(j1 *Plan.JNode, j2 *Plan.JNode, j1Map map[string]*Plan.JNode, j2Map map[string]*Plan.JNode) string {
	rs := ""
	lTbl1Node := j1.LChild
	rTbl1Node := j1.RChild
	lTbl2Node := j2.LChild
	rTbl2Node := j2.RChild

	lTbl1 := j1.LChild.ObtainSigString()
	rTbl1 := j1.RChild.ObtainSigString()
	lTbl2 := j2.LChild.ObtainSigString()
	rTbl2 := j2.RChild.ObtainSigString()

	lTbl1Prime, _ := j2Map[lTbl1]
	rTbl1Prime, _ := j2Map[rTbl1]
	lTbl2Prime, _ := j1Map[lTbl2]
	rTbl2Prime, _ := j1Map[rTbl2]

	actDiffl1 := lTbl1Node.ActRow != lTbl1Prime.ActRow
	actDiffr1 := rTbl1Node.ActRow != rTbl1Prime.ActRow
	estDiffl1 := lTbl1Node.EstRow != lTbl1Prime.EstRow
	estDiffr1 := rTbl1Node.EstRow != rTbl1Prime.EstRow

	actDiffl2 := lTbl2Node.ActRow != lTbl2Prime.ActRow
	actDiffr2:= rTbl2Node.ActRow != rTbl2Prime.ActRow
	estDiffl2 := lTbl2Node.EstRow != lTbl2Prime.EstRow
	estDiffr2 := rTbl2Node.EstRow != rTbl2Prime.EstRow

	if actDiffl1 || actDiffr1 ||actDiffl2 || actDiffr2 {
		//panic("We Cannot Support That The Number Of True Returned Rows IS Different.")
		rs += "We Cannot Support That The Number Of True Returned Rows IS Different."
		return rs
	}

	if estDiffl1 {
		rs += fmt.Sprintf("Estimated Cardinality On Table:%s Is Different.", lTbl1)
		if isJN1CloserToTrueCard(lTbl1Node, lTbl1Prime) {
			rs += " P1 IS Closer To The True Card."
		} else {
			rs += " P2 IS Closer To The True Card."
		}
	}
	if estDiffr1 {
		rs += fmt.Sprintf("Estimated Cardinality On Table:%s Is Different", rTbl1)
		if isJN1CloserToTrueCard(rTbl1Node, rTbl1Prime) {
			rs += " P1 IS Closer To The True Card."
		} else {
			rs += " P2 IS Closer To The True Card."
		}
	}

	if estDiffl2 {
		rs += fmt.Sprintf("Estimated Cardinality On Table:%s Is Different.", lTbl2)
		if isJN1CloserToTrueCard(lTbl2Node, lTbl2Prime) {
			rs += " P1 IS Closer To The True Card."
		} else {
			rs += " P2 IS Closer To The True Card."
		}
	}
	if estDiffr2 {
		rs += fmt.Sprintf("Estimated Cardinality On Table:%s Is Different", rTbl2)
		if isJN1CloserToTrueCard(rTbl2Node, rTbl2Prime) {
			rs += " P1 IS Closer To The True Card."
		} else {
			rs += " P2 IS Closer To The True Card."
		}
	}
	return rs
}

func handleInternalDiff(jn1 *Plan.JNode, jn2 *Plan.JNode, joinTree1Map map[string]*Plan.JNode, joinTree2Map map[string]*Plan.JNode) string {
	rs := ""
	tbl1 := ""
	tbl2 := ""
	var tbl1Node *Plan.JNode = nil
	var tbl2Node *Plan.JNode = nil
	if jn1.LChild.NodeType == "Reader" {
		tbl1Node = jn1.LChild
		tbl1 = jn1.LChild.ObtainSigString()
	} else {
		tbl1Node = jn1.RChild
		tbl1 = jn1.RChild.ObtainSigString()
	}
	jn1RCPrime, ok := joinTree2Map[tbl1]
	if !ok {
		panic(fmt.Sprintf("No Find Table %s In P2", tbl1))
	}
	if jn2.LChild.NodeType == "Reader" {
		tbl2Node = jn2.LChild
		tbl2 = jn2.LChild.ObtainSigString()
	} else {
		tbl2Node = jn2.RChild
		tbl2 = jn2.RChild.ObtainSigString()
	}
	jn2RCPrime, ok := joinTree1Map[tbl2]
	if !ok {
		panic(fmt.Sprintf("No Find Table %s In P1", tbl2))
	}
	actDiff1 := tbl1Node.ActRow != jn1RCPrime.ActRow
	actDiff2 := tbl2Node.ActRow != jn2RCPrime.ActRow
	estDiff1 := tbl1Node.EstRow != jn1RCPrime.EstRow
	estDiff2 := tbl2Node.EstRow != jn2RCPrime.EstRow
	if actDiff1 || actDiff2 {
		panic("We Cannot Support That The Number Of True Returned Rows IS Different.")
	}

	if estDiff1 {
		rs += fmt.Sprintf("Estimated Cardinality On Table:%s Is Different.", tbl1)
		if isJN1CloserToTrueCard(tbl1Node, jn1RCPrime) {
			rs += " P1 IS Closer To The True Card."
		} else {
			rs += " P2 IS Closer To The True Card."
		}
	}
	if estDiff2 {
		rs += fmt.Sprintf("Estimated Cardinality On Table:%s Is Different", tbl2)
		if isJN1CloserToTrueCard(tbl2Node, jn2.RChild) {
			rs += " P1 IS Closer To The True Card."
		} else {
			rs += " P2 IS Closer To The True Card."
		}
	}

	if !estDiff1 && !estDiff2 {
		diff1 := jn1.EstRow < jn2.EstRow
		diff2 := jn1.ActRow < jn2.ActRow
		if diff1 != diff2 {
			rs += " Join Size Estimation Is Not Accurate."
			if diff1 {
				rs += fmt.Sprintf(" The Size Of  %s IS Underestimated", jn1.ID)
			} else {
				rs += fmt.Sprintf(" The Size Of  %s IS Underestimated", jn2.ID)
			}
		}
	}
	return rs
}

// Only Work On Left-Depth Tree.
func JODiff(n1 *Plan.Node, n2 *Plan.Node, inputType int) string {
	rs := ""
	joinTree1 := n1.ObtainJoinTree()
	joinTree2 := n2.ObtainJoinTree()

	joinTree1Map := joinTree1.ObtainMap()
	joinTree2Map := joinTree2.ObtainMap()

	stack1 := Plan.NewStack() // stack to save n1
	stack2 := Plan.NewStack() // stack to save n2

	cur1 := joinTree1
	cur2 := joinTree2
	for cur1 != nil && cur2 != nil {
		stack1.Push(cur1)
		stack2.Push(cur2)
		if cur1.LChild.NodeType == "Join" {
			cur1 = cur1.LChild
		} else if cur1.RChild.NodeType == "Join" {
			cur1 = cur1.RChild
		} else {
			cur1 = nil
		}
		if cur2.LChild.NodeType == "Join" {
			cur2 = cur2.LChild
		} else if cur2.RChild.NodeType == "Join" {
			cur2 = cur2.RChild
		} else {
			cur2 = nil
		}
	}

	var jn1 *Plan.JNode = nil
	var jn2 *Plan.JNode = nil
	isFirst := true
	for !stack1.Empty() && !stack2.Empty() {
		jn1 = stack1.Pop().(*Plan.JNode)
		jn2 = stack2.Pop().(*Plan.JNode)
		//if jn1.ObtainSigString() == jn2.ObtainSigString() {
		//	isFirst = false
		//	continue
		//}
		// In this case, the first two join table is different.
		if isFirst {
			rs = handleFirstDiffJoin(jn1, jn2, joinTree1Map, joinTree2Map)
			isFirst = false
			break
		}
		// Check CE
		rs = handleInternalDiff(jn1, jn2, joinTree1Map, joinTree2Map)
		break // currently, we only report the first diff order.
	}

	return rs
}