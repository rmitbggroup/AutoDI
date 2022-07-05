package Explain

import (
	"ExplainPlan/Plan"
	"fmt"
	"sort"
	"strings"
)

func obtainPossibleSortColumns(s *Plan.Node) string {
	rs := ""
	opInfo := s.OperatorInfo
	switch s.OP {
		case "Sort", "TopN":
			if s.OP == "TopN" {
				idx := strings.Index(opInfo, ", offset:")
				opInfo = opInfo[:idx]
			}
			sortColumns := make([]string, 0, 1)
			sortItems := strings.Split(opInfo, ", ")
			for _, sortItem := range sortItems {
				column := strings.Split(sortItem, ":")[0]
				if strings.Index(column, ".") > 0 {
					sortColumns = append(sortColumns, strings.Split(column, ".")[2])
				} else {
					sortColumns = append(sortColumns, "Column")
				}
			}
			rs = strings.Join(sortColumns, ", ")
		case "IndexRangeScan":
			if strings.Index(opInfo, "keep order: true") < 0 {
				return rs
			}
			ab := s.TblName
			idx1 := strings.Index(ab, "(")
			idx2 := strings.Index(ab, ")")
			return ab[idx1+1:idx2]
		case "TableRangeScan":
			idx := strings.Index(s.TblName, ":")
			tblName := s.TblName[idx+1:]
			pk := Plan.ObtainPrimaryKey(tblName)
			return pk

	}
	return rs
}

// Return the sort mapping. Currently, we think the sorted columns from the same table.
func obtainSortMapping(s *Plan.Node, cands []*Plan.Node) string {
	sSortInfo := obtainPossibleSortColumns(s)
	for _, cand := range cands {
		if sSortInfo == obtainPossibleSortColumns(cand) {
			return fmt.Sprintf("Sort is supported with node %s_%s", cand.OP, cand.ID)
		}
	}
	// This case should not be happened.
	return "Not Find"
}

func FindDiff(n1 *Plan.Node, n2 *Plan.Node, logicalMap map[string] *Plan.Node, explainOnly bool) []*DiffItem {
	rs := make([]*DiffItem, 0, 2)
	// Process Sort case
	sortedRelatdNodes1 := n1.SortRelatedNode()
	sortedRelatdNodes2 := n2.SortRelatedNode()
	n1IsSort := n1.OP == "Sort"
	n2IsSort := n2.OP == "Sort"
	newN1 := n1
	newN2 := n2
	if n1IsSort != n2IsSort {
		if n1IsSort {
			rs = append(rs, &DiffItem{DiffMessage: "SortDiff#"+obtainSortMapping(n1, sortedRelatdNodes2)})
			newN1 = n1.LChild
		} else {
			rs = append(rs, &DiffItem{DiffMessage: "SortDiff#"+obtainSortMapping(n2, sortedRelatdNodes1)})
			newN2 = n2.LChild
		}
	}
	// Handle the plan tree with same node type & count.
	rs = append(rs, findDiff(newN1, newN2, logicalMap)...)
	if !explainOnly {
		sort.Sort(ByDiffRunTime(rs))
	}
	return rs
}

func findDiff(n1 *Plan.Node, n2 *Plan.Node, logicalMap map[string] *Plan.Node) []*DiffItem {
	rs := make([]*DiffItem, 0, 2)
	if n1 == nil || n2 == nil {
		return rs
	}
	if n1.PhysicalSig.IsSame(n2.PhysicalSig){
		return rs
	}
	if n1.LogicalSig.IsSame(n2.LogicalSig) && !n1.PhysicalSig.IsSame(n2.PhysicalSig) {
		if n1.NodeType == n2.NodeType && n1.PhysicalSig.HasSameNodeImp(n2.PhysicalSig) {
			if n1.LChild != nil && !n1.LChild.LogicalSig.IsSame(n2.LChild.LogicalSig) || n1.RChild != nil && !n1.RChild.LogicalSig.IsSame(n2.RChild.LogicalSig) {
				rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "JODiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
				s := Plan.NewStack()
				if n1.LChild != nil {
					s.Push(n1.LChild)
				}
				if n1.RChild != nil {
					s.Push(n1.RChild)
				}

				for !s.Empty() {
					cur := s.Pop().(*Plan.Node)
					curPrime, ok := logicalMap[cur.LogicalSig.Signature()]
					if ok { // We do not need to visit its children.
						rs = append(rs, findDiff(cur, curPrime, logicalMap)...)
					} else {
						if cur.LChild != nil {
							s.Push(cur.LChild)
						}
						if cur.RChild != nil {
							s.Push(cur.RChild)
						}
					}
				}
			} else {
				rs = append(rs, findDiff(n1.LChild, n2.LChild, logicalMap)...)
				rs = append(rs, findDiff(n1.RChild, n2.RChild, logicalMap)...)
			}
			return rs
		}
		if n1.NodeType == n2.NodeType && n1.PhysicalSig.HasSameParas(n2.PhysicalSig) {
			if n1.NodeType == "Reader" {
				rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "DADiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
				return rs
			} else
			if n1.NodeType == "Join" {
				rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "ImpDiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
				if n1.LChild != nil && !n1.LChild.LogicalSig.IsSame(n2.LChild.LogicalSig) || n1.RChild != nil && !n1.RChild.LogicalSig.IsSame(n2.RChild.LogicalSig) {
					rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "JODiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
					s := Plan.NewStack()
					if n1.LChild != nil {
						s.Push(n1.LChild)
					}
					if n1.RChild != nil {
						s.Push(n1.RChild)
					}

					for !s.Empty() {
						cur := s.Pop().(*Plan.Node)
						curPrime, ok := logicalMap[cur.LogicalSig.Signature()]
						if ok { // We do not need to visit its children.
							rs = append(rs, findDiff(cur, curPrime, logicalMap)...)
						} else {
							if cur.LChild != nil {
								s.Push(cur.LChild)
							}
							if cur.RChild != nil {
								s.Push(cur.RChild)
							}
						}
					}
				} else {
					rs = append(rs, findDiff(n1.LChild, n2.LChild, logicalMap)...)
					rs = append(rs, findDiff(n1.RChild, n2.RChild, logicalMap)...)
				}
				return rs
			} else {
				rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "ImpDiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
				rs = append(rs, findDiff(n1.LChild, n2.LChild, logicalMap)...)
				rs = append(rs, findDiff(n1.RChild, n2.RChild, logicalMap)...)
				return rs
			}
		}
	}
	// We come here when node type is different or node type is the same while operator is different both in imp and paras
	if !n1.LogicalSig.IsSame(n2.LogicalSig) {
		if n1.NodeType == n2.NodeType && n1.NodeType == "Join" {
			rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "JODiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
		} else {
			rs = append(rs, &DiffItem{N1: n1, N2: n2, DiffMessage: "OODiff", DiffLevel: SelfMaxInt(n1.Level, n2.Level)})
			_ = n1.LogicalSig.IsSame(n2.LogicalSig)
		}

		s := Plan.NewStack()
		if n1.LChild != nil {
			s.Push(n1.LChild)
		}
		if n1.RChild != nil {
			s.Push(n1.RChild)
		}

		for !s.Empty() {
			cur := s.Pop().(*Plan.Node)
			curPrime, ok := logicalMap[cur.LogicalSig.Signature()]
			if ok { // We do not need to visit its children.
				rs = append(rs, findDiff(cur, curPrime, logicalMap)...)
			} else {
				if cur.LChild != nil {
					s.Push(cur.LChild)
				}
				if cur.RChild != nil {
					s.Push(cur.RChild)
				}
			}
		}
	}

	// Note: the case with different LogicalSignature has been handled in `Join Order' and `Operator Oder'
	return rs
}