package Plan

import (
	"container/list"
	"crypto"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type Node struct {
	OP string
	ID string
	NodeType    string

	EstRows float64
	ActRows float64
	Cost float64
	Runtime string
	OperatorInfo string
	// It records the left child's estRow.
	RightEstRows float64
	Level int

	LChild *Node
	RChild *Node

	TblName string

	LogicalSig  *LogicalSignature
	PhysicalSig *PhysicalSignature

	HasSpecialLimit bool
	HasSpecialTopN bool
	HasSpecialSort bool
	HasSpecialAgg bool
	// record them to calculate the cost
	LSpecialNode *Node
	RSpecialNode *Node
}

type PhysicalSignature struct {
	nodeImp string
	leftImp string
	rightImp string
}

func (ps *PhysicalSignature) obtainSubtreeSig() string {
	//1 used
	return ps.nodeImp + "@#@1" + ps.leftImp + "@_@1" + ps.rightImp + "@#@"
}

func (ps *PhysicalSignature) IsSame(another *PhysicalSignature) bool {
	return ps.obtainSubtreeSig() == another.obtainSubtreeSig()
}

func (ps *PhysicalSignature) HasSameNodeImp(another *PhysicalSignature) bool {
	return ps.nodeImp == another.nodeImp
}

func (ps *PhysicalSignature) HasSameParas(another *PhysicalSignature) bool {
	hasSame := false
	nodeInfo1 := strings.Split(ps.nodeImp, "#")
	nodeInfo2 := strings.Split(another.nodeImp, "#")
	if nodeInfo1[1] == nodeInfo2[1] {
		hasSame = true
	}
	return hasSame
}

func (ps *PhysicalSignature) GetNodeImp() string {
	return ps.nodeImp
}

type LogicalSignature struct {
	ntCount map[string]int
	lsDetails *list.List
	firstItem map[string]*list.Element
}

// Different nodes has different detail formats.
type LogicalSignatureItem struct {
	nt string
	detail string
}

func (ls *LogicalSignature) Insert(d *LogicalSignatureItem) bool {
	if d == nil {
		return false
	}

	var newLS *list.Element
	first, has := ls.firstItem[d.nt]
	if has {
		ls.lsDetails.InsertAfter(d, first)
		ls.ntCount[d.nt] = ls.ntCount[d.nt] + 1
	} else {
		newLS = ls.lsDetails.PushBack(d)
		ls.ntCount[d.nt] = 1
		ls.firstItem[d.nt] = newLS
	}
	return true
}

func (ls *LogicalSignature) Append(childLS *LogicalSignature) bool {
	if childLS == nil {
		return false
	}
	for elem := childLS.lsDetails.Front(); elem != nil; elem = elem.Next() {
		d := elem.Value.(*LogicalSignatureItem)
		ls.Insert(d)
	}
	return true
}

func (ls *LogicalSignature) Signature() string {
	if ls.lsDetails.Len() == 0 {
		return ""
	}
	items := make([]string, 0, 2)
	for elem := ls.lsDetails.Front(); elem != nil; elem = elem.Next() {
		d := elem.Value.(*LogicalSignatureItem)
		items = append(items, d.nt+"#"+d.detail)
	}
	sort.Strings(items)
	hash := crypto.SHA256.New()
	hash.Write([]byte(strings.Join(items, "$#$")))
	sum := hash.Sum(nil)
	s := hex.EncodeToString(sum)
	return s
}

func (ls *LogicalSignature) IsSame(another *LogicalSignature) bool {
	isSame := true
	for k, v := range ls.ntCount {
		ntC, has := another.ntCount[k]
		if !has || ntC != v{
			isSame = false
			break
		}

		first, _ := ls.firstItem[k]
		var item = &LogicalSignatureItem{}
		for first != nil {
			item = first.Value.(*LogicalSignatureItem)
			if item.nt != k {
				break
			}
			// TODO: linear, not efficient.
			var find = false
			aFirst, _ := another.firstItem[k]
			var aItem = &LogicalSignatureItem{}
			for aFirst != nil {
				aItem = aFirst.Value.(*LogicalSignatureItem)
				if aItem.nt != k {
					break
				}
				if item.detail == aItem.detail{
					find = true
					break
				}
				aFirst = aFirst.Next()
			}
			if !find {
				isSame = false
				break
			}
			first = first.Next()
		}

		if !isSame {
			break
		}
	}
	return isSame
}

func (ls *LogicalSignature) ISIncluded(another *LogicalSignature) bool {
	included := true
	for k, v := range ls.ntCount {
		ntC, has := another.ntCount[k]
		if !has || ntC < v{
			included = false
			break
		}

		first, _ := ls.firstItem[k]
		var item = &LogicalSignatureItem{}
		for first != nil {
			item = first.Value.(*LogicalSignatureItem)
			if item.nt != k {
				break
			}
			// TODO: linear, not efficient.
			var find = false
			aFirst, _ := another.firstItem[k]
			var aItem = &LogicalSignatureItem{}
			for aFirst != nil {
				aItem = aFirst.Value.(*LogicalSignatureItem)
				if aItem.nt != k {
					break
				}
				if item.detail == aItem.detail{
					find = true
					break
				}
				aFirst = aFirst.Next()
			}
			if !find {
				included = false
				break
			}
			first = first.Next()
		}

		if !included {
			break
		}
	}
	return included
}

var accessPath  = map[string]int {
	"TableReader":1,
	"IndexReader":1,
	"IndexLookUp":1,
}

func IsAccessPath(p *Node) bool {
	_, ok := accessPath[p.OP]
	return ok
}

var ignoredLS = map[string]int {
	"Projection":1,
	"Selection":1,
}

var WorkloadType = "job"

var tpchPrimaryKey = map[string] string {
	"nation":"n_nationkey",
	"region":"r_regionkey",
	"part":"p_partkey",
	"supplier":"s_suppkey",
	"partsupp": "ps_partkey, ps_suppkey",
	"customer":"c_custkey",
	"orders":"o_oderkey",
	"lineitem":"l_orderkey, l_linenumber",
}

var jobPrimaryKey = "id"

func ObtainPrimaryKey(tableName string) string {
	key := ""
	switch WorkloadType {
	case "job":
		key = "id"
	case "tpch":
		_key, ok := tpchPrimaryKey[tableName]
		if !ok {
			if tableName[0] == 'n' {
				key, _ = tpchPrimaryKey["nation"]
			} else if tableName[0] == 'l' {
				key, _ = tpchPrimaryKey["lineitem"]
			} else {
				panic(fmt.Sprintf("Not Mapping For %s", tableName))
			}
		} else {
			key = _key
		}
	default:
		panic("No Such Workload Information!")
	}
	return key
}

var bndToOp = map[string] string {
	")": "lt",
	"]": "le",
	"(": "gt",
	"[": "ge",
}

var dbname = "test"

//[-inf,3)
func genExpr(tb string, attrs []string, r string) []string {
	exprs := make([]string, 0, 1)
	_v := strings.Split(r, ",")
	rv := _v[1]
	lv := _v[0]
	rbnd := rv[len(rv)-1:]
	lbnd := lv[0:1]
	rvsa := strings.Split(rv[:len(rv)-1], " ")
	lvsa := strings.Split(lv[1:], " ")

	for i, v := range rvsa {
		if i < len(rvsa)-1 {
			cond := "eq("+dbname+"."+tb+"."+attrs[i]+", "+v+")"
			exprs = append(exprs, cond)
		} else {
			_rop, _ := bndToOp[rbnd]
			_lop, _ := bndToOp[lbnd]
			if rvsa[i] != "+inf" {
				cond := _rop+"("+dbname+"."+tb+"."+attrs[i]+", "+v+")"
				exprs = append(exprs, cond)
			}
			if lvsa[i] != "-inf" {
				cond := _lop+"("+dbname+"."+tb+"."+attrs[i]+", "+lvsa[i]+")"
				exprs = append(exprs, cond)
			}
		}
	}
	return exprs
}

func getExprsOfTableRangeScan(p *Node) []string {
	// Suppose currently, only len(pk) == 1 in this case.
	//  └─TableRangeScan_5     | 3333.33 | cop[tikv] | table:pkt     | range:[-inf,3), keep order:false, stats:pseudo |
	condDetails := make([]string, 0, 2)
	pkAttrss := ObtainPrimaryKey(p.TblName)
	pkAttrsa := strings.Split(pkAttrss, ", ")
	idx := strings.Index(p.OperatorInfo, ", keep order")
	expr := p.OperatorInfo[6:idx]
	condsa := genExpr(p.TblName, pkAttrsa, expr)
	for _, cond := range condsa {
		condDetails = append(condDetails, cond)
	}
	return condDetails
}

func getExprsFromIndexRange(p *Node) (string, []string) {
	idx := strings.Index(p.TblName, ", index:")
	tableName := p.TblName[6:idx]
	conds := make([]string, 0, 1)
	// IndexReader->IndexFullScan
	if p.OP == "IndexFullScan" {
		return tableName, conds
	}
	idxAttrs := strings.Split(p.TblName[idx+8:len(p.TblName)-1], ", ")
	if strings.Index(p.OperatorInfo, "decided by") < 0 {
		idx = strings.Index(p.OperatorInfo, ", keep order")
		r := p.OperatorInfo[6:idx]
		conds = genExpr(tableName, idxAttrs, r)
	}
	return tableName, conds
}

func hashJoinDetails(p *Node) string {
	opInfo := p.OperatorInfo
	idx := strings.Index(opInfo, "equal")
	condss := ") " + opInfo[idx+7:len(opInfo)-2]
	condsa := strings.Split(condss, ") eq(")[1:]
	joinCols := make([]string, 0, 2)
	for _, conds := range condsa {
		cols := strings.Split(conds, ", ")
		for _, col := range cols {
			if strings.Index(col, "#") > -1 {
				joinCols = append(joinCols, "Column")
			} else {
				joinCols = append(joinCols, col)
			}
		}
	}
	sort.Strings(joinCols)
	joinDetails := strings.Join(joinCols,", ")
	return joinDetails
}

func indexJoinDetails(p *Node) string {
	opInfo := p.OperatorInfo
	joinCols := make([]string, 0, 2)
	//idx1 := strings.Index(opInfo, "outer key")
	//idx2 := strings.Index(opInfo, "inner key")
	//idx3 := strings.Index(opInfo, ", equal cond")
	//outerKeyss := opInfo[idx1+10:idx2-2]
	//outerKeya := strings.Split(outerKeyss, ", ")
	//for _, outerKey := range outerKeya {
	//	joinCols = append(joinCols, outerKey)
	//}
	//innerKeyss := opInfo[idx2+10:idx3]
	//innerKeya := strings.Split(innerKeyss, ", ")
	//for _, innerKey := range innerKeya {
	//	joinCols = append(joinCols, innerKey)
	//}
	idx3 := strings.Index(opInfo, ", equal cond:")
	joinCond := "), "+opInfo[idx3+13:len(opInfo)-1]
	condsa := strings.Split(joinCond, "), eq(")[1:]
	for _, conds := range condsa {
		cols := strings.Split(conds, ", ")
		for _, col := range cols {
			if strings.Index(col, "#") > -1 {
				joinCols = append(joinCols, "Column")
			} else {
				joinCols = append(joinCols, col)
			}
		}
	}
	sort.Strings(joinCols)
	joinDetails := strings.Join(joinCols,", ")
	return joinDetails
}

func mergeJoinDetails(p *Node) string {
	opInfo := p.OperatorInfo
	joinCols := make([]string, 0, 2)
	idx1 := strings.Index(opInfo, "left key")
	idx2 := strings.Index(opInfo, "right key")
	leftKeys := opInfo[idx1+9:idx2-2]
	rightKeys := opInfo[idx2+10:]
	leftKeya := strings.Split(leftKeys, ", ")
	for _, lk := range leftKeya {
		joinCols = append(joinCols, lk)
	}
	rightKeya := strings.Split(rightKeys, ", ")
	for _, rk := range rightKeya {
		joinCols = append(joinCols, rk)
	}
	sort.Strings(joinCols)
	joinDetails := strings.Join(joinCols,", ")
	return joinDetails
}

func aggregationDetails(p *Node) string {
	i := 0
	findFlag := false
	newOp := ""
	for i < len(p.OperatorInfo) {
		if p.OperatorInfo[i] == '#' {
			findFlag = true
		} else if findFlag && isNumber(p.OperatorInfo[i]) {

		} else if findFlag && !isNumber(p.OperatorInfo[i]) {
			findFlag = false
			newOp += fmt.Sprintf("%c",p.OperatorInfo[i])
		} else {
			newOp += fmt.Sprintf("%c",p.OperatorInfo[i])
		}
		i++
	}
	aggDetails := make([]string, 0, 2)
	if newOp[0] == 'g' { // with group by
		idx1 := strings.Index(p.OperatorInfo, "funcs:")
		aggDetails = append(aggDetails, newOp[:idx1-2])
		funcss := newOp[idx1-2:]
		funcsa := strings.Split(funcss, ", funcs:")[1:]
		for _, funcs := range funcsa {
			aggDetails = append(aggDetails, strings.Split(funcs, "->")[0])
		}
	} else { // without group by
		idx1 := strings.Index(newOp, ", funcs:")
		if idx1 < 0 {
			firstFunc := newOp[6:]
			aggDetails = append(aggDetails, strings.Split(firstFunc, "->")[0])
		} else {
			funcss := newOp[idx1-2:]
			funcsa := strings.Split(funcss, ", funcs:")[1:]
			for _, funcs := range funcsa {
				aggDetails = append(aggDetails, strings.Split(funcs, "->")[0])
			}
		}

	}

	return strings.Join(aggDetails, ", ")
}

func tableReaderDetails(p *Node) (string, string) {
	child := p.LChild
	condDetails := make([]string, 0, 2)
	tableName := ""
	scanType := "full"
	switch child.OP {
	case "Selection":
		condsa := strings.Split(child.OperatorInfo, ", ")
		for _, cond := range condsa {
			idx := strings.Index(cond,"isnull")
			if cond != "" && idx < 0{
				condDetails = append(condDetails, cond)
			}
		}

		grandson := child.LChild
		if grandson.OP == "TableRangeScan" {
			scanType = "range"
		}

		if grandson.OP == "TableRangeScan" && strings.Index(grandson.OperatorInfo, "decided by") < 0 {
			condsa :=  getExprsOfTableRangeScan(grandson)
			for _, cond := range condsa {
				condDetails = append(condDetails, cond)
			}
		}
		tableName = grandson.TblName[6:]
	case "TableRangeScan":
		scanType = "range"
		if strings.Index(child.OperatorInfo, "decided by") < 0 {
			condsa :=  getExprsOfTableRangeScan(child)
			for _, cond := range condsa {
				condDetails = append(condDetails, cond)
			}
		}
		tableName = child.TblName[6:]
	case "TableFullScan":
		tableName = child.TblName[6:]
	}
	sort.Strings(condDetails)
	return scanType, tableName+"#"+strings.Join(condDetails, ", ")
}

func indexReaderDetails(p *Node) (string, string) {
	child := p.LChild
	tblName, conds := getExprsFromIndexRange(child)
	sort.Strings(conds)
	idx := strings.Index(p.TblName, ", index:")
	idx2 := strings.Index(child.TblName, "(")
	idxName := child.TblName[idx+8:idx2]

	return idxName, tblName+"#"+strings.Join(conds, ", ")
}

func GetRangeCondition(p *Node) (string, string) {
	var idxName = "_pk"
	var rangeConds = ""
	switch  p.OP {
		case "IndexRangeScan":
			_, conds := getExprsFromIndexRange(p)
			sort.Strings(conds)
			idx := strings.Index(p.TblName, ", index:")
			idx2 := strings.Index(p.TblName, "(")
			idxName = p.TblName[idx+8:idx2]
			rangeConds = strings.Join(conds, ", ")
		case "TableRangeScan":
			rangeConds =  strings.Join(getExprsOfTableRangeScan(p), ", ")
	}
	return idxName, rangeConds
}

func indexLookUpDetails(p *Node) (string, string) {
	rChild := p.RChild // indexrange
	if rChild.OP == "Selection" {
		rChild = rChild.LChild
	}
	conds := make([]string, 0, 2)
	idx := strings.Index(rChild.TblName, ", index:")
	tableName := rChild.TblName[6:idx]
	idx2 := strings.Index(rChild.TblName, "(")
	idxName := rChild.TblName[idx+8:idx2]
	if strings.Index(rChild.OperatorInfo, "decided by") < 0 {
		_, _conds := getExprsFromIndexRange(rChild)
		for _, cond := range _conds {
			if cond != "" {
				conds = append(conds, cond)
			}
		}
	}
	if p.LChild.OP == "Selection" {
		condsa := strings.Split(p.LChild.OperatorInfo, ", ")
		for _, cond := range condsa {
			idx := strings.Index(cond,"isnull")
			if cond != "" && idx < 0 {
				conds = append(conds, cond)
			}
		}
	}
	sort.Strings(conds)
	return idxName, tableName+"#"+strings.Join(conds, ", ")
}

func (p *Node) visitChildren() bool {
	_, in := accessPath[p.OP]
	return !in
}

func (p *Node) GetLogicalSignature() {
	p.LogicalSig = &LogicalSignature{
		ntCount: make(map[string]int),
		lsDetails: list.New(),
		firstItem: make(map[string]*list.Element),
	}
	if p.LChild != nil && p.visitChildren() {
		p.LChild.GetLogicalSignature()
		p.LogicalSig.Append(p.LChild.LogicalSig)
	}
	if p.RChild != nil && p.visitChildren() {
		p.RChild.GetLogicalSignature()
		p.LogicalSig.Append(p.RChild.LogicalSig)
	}
	nodeLS := &LogicalSignatureItem{}
	switch p.OP {
		case "HashJoin":
			// inner join, equal:[eq(test.part.p_partkey, test.partsupp.ps_partkey) eq(test.partsupp.ps_supplycost, Column#50)]
			nodeLS.nt = p.NodeType
			nodeLS.detail = hashJoinDetails(p)
		case "MergeJoin":
			nodeLS.nt = p.NodeType
			nodeLS.detail = mergeJoinDetails(p)
	case "IndexJoin", "IndexHashJoin":
			// inner join, inner:TableReader_55, outer key:test.nation.n_regionkey, inner key:test.region.r_regionkey, equal cond:eq(test.nation.n_regionkey, test.region.r_regionkey)
			nodeLS.nt = p.NodeType
			nodeLS.detail = indexJoinDetails(p)
		case "TopN":
			nodeLS.nt = p.NodeType
			nodeLS.detail = topnDetails(p)
		case "HashAgg", "StreamAgg":
			// group by:test.partsupp.ps_partkey, funcs:min(test.partsupp.ps_supplycost)->Column#50, funcs:firstrow(test.partsupp.ps_partkey)->test.partsupp.ps_partkey
			nodeLS.nt = p.NodeType
			nodeLS.detail = aggregationDetails(p)
		case "TableReader":
			// Selection/TableFullScan,TableRangeScan (decided by)
			nodeLS.nt = p.NodeType
			_, nodeLS.detail = tableReaderDetails(p)
		case "IndexReader":
			nodeLS.nt = p.NodeType
			_, nodeLS.detail = indexReaderDetails(p)
		case "IndexLookUp":
			// double read
			// IndexRangeScan (decided by) TableRowIDScan
			nodeLS.nt = p.NodeType
			_, nodeLS.detail = indexLookUpDetails(p)
		case "Projection":
			// currently, ignore it
			//println("LogicalSignature Generation - Projection...")
		case "Selection":
			// currently, ignore it
			//println("LogicalSignature Generation - Single Selection...")
		case "Sort":
			nodeLS.nt = p.NodeType
			nodeLS.detail = sortDetails(p)

		default:
			panic("not support "+p.OP)
	}
	_, isIgnore := ignoredLS[p.OP]
	if !isIgnore {
		p.LogicalSig.Insert(nodeLS)
	}
}

func sortDetails(p *Node) string {
	rs := make([]string, 0, 2)
	sortItems := strings.Split(p.OperatorInfo, ", ")

	for _, sortItem := range sortItems {
		if strings.Index(sortItem, ".") > 0 {
			rs = append(rs, sortItem)
		} else {
			//Column#53:desc
			idx := strings.Index(sortItem, ":")
			if idx > 0 {
				rs = append(rs, "Column"+sortItem[idx:])
			} else {
				rs = append(rs, "Column")
			}
		}
	}

	return strings.Join(rs, ", ")
}

func topnDetails(p *Node) string {
	rs := make([]string, 0, 2)
	idx1 := strings.Index(p.OperatorInfo, ", offset:")
	opSortItem := p.OperatorInfo[:idx1]

	sortItems := strings.Split(opSortItem, ", ")
	for _, sortItem := range sortItems {
		if strings.Index(sortItem, ".") > 0 {
			rs = append(rs, sortItem)
		} else {
			//Column#53:desc
			idx := strings.Index(sortItem, ":")
			if idx > 0 {
				rs = append(rs, "Column"+sortItem[idx:])
			} else {
				rs = append(rs, "Column")
			}
		}
	}
	rs = append(rs, p.OperatorInfo[idx1+2:])
	return strings.Join(rs, ", ")
}

func (p *Node) GetPhysicalSignature() {
	p.PhysicalSig = &PhysicalSignature{}
	if p.LChild != nil && p.visitChildren() {
		p.LChild.GetPhysicalSignature()
		p.PhysicalSig.leftImp = p.LChild.PhysicalSig.obtainSubtreeSig()
	}
	if p.RChild != nil && p.visitChildren() {
		p.RChild.GetPhysicalSignature()
		p.PhysicalSig.rightImp = p.RChild.PhysicalSig.obtainSubtreeSig()
	}

	nodeImp := ""
	switch p.OP {
		case "TopN":
			nodeImp = "TopN#"+topnDetails(p)
		case "HashJoin":
			detail := hashJoinDetails(p)
			nodeImp = "HashJoin@inner join#"+detail
		case "MergeJoin":
			detail := mergeJoinDetails(p)
			nodeImp = "MergeJoin@inner join#"+detail
		case "IndexJoin":
			detail := indexJoinDetails(p)
			nodeImp = "indexJoin@inner join#"+detail
		case "IndexHashJoin":
			detail := indexJoinDetails(p)
			nodeImp = "IndexHashJoin@inner join#"+detail
		case "HashAgg":
			detail := aggregationDetails(p)
			nodeImp = "HashAgg#"+detail
		case "StreamAgg":
			detail := aggregationDetails(p)
			nodeImp = "StreamAgg#"+detail
		case "TableReader":
			scanType, detail := tableReaderDetails(p)
			nodeImp = "TableScan@"+scanType+"#"+detail
		case "IndexReader":
			indexName, detail := indexReaderDetails(p)
			nodeImp = "IndexScan@"+indexName+"#"+detail
		case "IndexLookUp":
			indexName, detail := indexLookUpDetails(p)
			nodeImp = "IndexScan@"+indexName+"#"+detail
		case "Projection":
			//println("PhysicalSignature Generation - Projection...")
		case "Selection":
			//println("PhysicalSignature Generation - Single Selection...")
		case "Sort":
			nodeImp = "Sort#"+sortDetails(p)
		default:
			panic("not support "+p.OP)

	}
	p.PhysicalSig.nodeImp = nodeImp
}

func (p *Node) ObtainJoinPath () string {
	if p.NodeType == "Join" {
		switch p.OP {
			case "IndexJoin", "IndexHashJoin", "IndexMergeJoin":
				return fmt.Sprintf("(%s,%s,%f)", p.LChild.ObtainJoinPath(), p.RChild.ObtainJoinPath(), p.EstRows)
			case "HashJoin":
				return fmt.Sprintf("(%s,%s,%f)", p.LChild.ObtainJoinPath(), p.RChild.ObtainJoinPath(), p.EstRows)
			default:
				panic("No Support This Join Type.")
		}
	} else if p.NodeType == "Reader" {
		return strings.Split(p.PhysicalSig.nodeImp, "#")[1]
	} else {
		return p.LChild.ObtainJoinPath()
	}
}

type JNode struct {
	NodeType string
	ID string
	EstRow float64
	Cost float64
	ActRow float64
	Runtime string

	LChild *JNode
	RChild *JNode

	Signature []string
}

func (p *Node) ObtainJoinTree() *JNode {
	if p.NodeType != "Join" && p.NodeType != "Reader" {
		return p.LChild.ObtainJoinTree()
	}
	jNode := &JNode{Signature: make([]string,0 , 2)}
	jNode.Cost = p.Cost
	jNode.Runtime = p.Runtime
	jNode.EstRow = p.EstRows
	jNode.ActRow = p.ActRows
	jNode.ID = p.ID
	jNode.NodeType = p.NodeType
	if p.NodeType == "Join" {
		//switch p.OP {
		//	case "IndexJoin", "IndexMergeJoin", "IndexHashJoin":
		//
		//	case "HashJoin":
		//		jNode.LChild = p.LChild.ObtainJoinTree()
		//		jNode.RChild =  p.RChild.ObtainJoinTree()
		//default:
		//	panic(fmt.Sprintf("Not Support Operator: %s!", p.OP))
		//}
		jNode.LChild = p.LChild.ObtainJoinTree()
		jNode.RChild = p.RChild.ObtainJoinTree()
		jNode.Signature = append(jNode.Signature, jNode.LChild.Signature...)
		jNode.Signature = append(jNode.Signature, jNode.RChild.Signature...)
	} else {
		tblName := strings.Split(p.PhysicalSig.nodeImp, "#")[1]
		jNode.Signature = append(jNode.Signature, tblName)
	}

	return jNode
}

func (jn *JNode) ObtainMap() map[string] *JNode {
	mMap := make(map[string] *JNode)
	sort.Strings(jn.Signature)
	sig := strings.Join(jn.Signature, ";")
	mMap[sig] = jn
	if jn.LChild != nil {
		lMap := jn.LChild.ObtainMap()
		for k, v := range lMap {
			mMap[k] = v
		}
	}
	if jn.RChild != nil {
		rMap := jn.RChild.ObtainMap()
		for k, v := range rMap {
			mMap[k] = v
		}
	}
	return mMap
}

func (jn *JNode) ObtainSigString() string {
	sort.Strings(jn.Signature)
	return strings.Join(jn.Signature, ";")
}

func (p *Node) RemoveProjectionAndSelection() *Node {
	newNode := p
	if IsAccessPath(p) { return newNode }
	if p.NodeType == "Projection" || p.NodeType == "Selection" {
		newNode = p.LChild.RemoveProjectionAndSelection()
	} else {
		if p.LChild != nil {
			newNode.LChild = p.LChild.RemoveProjectionAndSelection()
		}
		if p.RChild != nil {
			newNode.RChild = p.RChild.RemoveProjectionAndSelection()
		}
	}
	return newNode
}

func (p *Node) needChangeJoin() bool {
	return p.OP == "IndexJoin" || p.OP == "IndexHashJoin" || p.OP == "IndexMergeJoin"
}

func (p *Node) ChangeJoinSide() *Node {
	newNode := new(Node)
	*newNode = *p
	if IsAccessPath(p) { return newNode }
	if p.needChangeJoin() {
		newNode.LChild = p.RChild.ChangeJoinSide()
		newNode.RChild = p.LChild.ChangeJoinSide()
	} else {
		if p.LChild != nil {
			newNode.LChild = p.LChild.ChangeJoinSide()
		}
		if p.RChild != nil {
			newNode.RChild = p.RChild.ChangeJoinSide()
		}
	}
	return newNode
}

func (p *Node) AddLevelInfo(parentLevel int) {
	p.Level = parentLevel + 1
	if p.NodeType == "Reader" { return }
	if p.LChild != nil {
		p.LChild.AddLevelInfo(parentLevel+1)
	}
	if p.RChild != nil {
		p.RChild.AddLevelInfo(parentLevel+1)
	}
}

func (p *Node) UpdateRowCountInfo() {
	if p.OP == "IndexJoin" || p.OP == "IndexMergeJoin" || p.OP == "IndexHashJoin" {
		if  p.RChild.EstRows == 0.0 && p.LChild.EstRows > 0.0 {
			p.RChild.EstRows = p.EstRows/p.LChild.EstRows
		}
		if  p.LChild.EstRows == 0.0 && p.RChild.EstRows > 0.0 {
			p.LChild.EstRows = p.EstRows/p.RChild.EstRows
		}

		p.RChild.RightEstRows = p.LChild.EstRows

		if p.LChild.NodeType != "Reader" {
			p.LChild.UpdateRowCountInfo()
		}
		if p.RChild.NodeType != "Reader" {
			p.RChild.UpdateRowCountInfo()
		}
	} else {
		if p.LChild != nil {
			p.LChild.UpdateRowCountInfo()
		}
		if p.RChild != nil {
			p.RChild.UpdateRowCountInfo()
		}
	}
}

// TODO: Implement it.
func (p *Node) GetLogicalMap2() map[string] []*Node {

	return nil
}

// We suppose this function is called after GetLogicalSignature.
// TODO: how to differentiate the node with same sig, esp. for the table read.
func (p *Node) GetLogicalMap() map[string] *Node {
	mMap := make(map[string]*Node)
	sig := p.LogicalSig.Signature()
	_, ok := ignoredLS[p.OP]
	if sig != "" && !ok {
		mMap[sig] = p
	}
	if p.LChild != nil && p.visitChildren() {
		//sig := p.LChild.LogicalSig.Signature()
		lMap := p.LChild.GetLogicalMap()
		for k, v := range lMap {
			// TODO: fix conflict
			//o, has := mMap[k]
			//if has {
			//	println(v.OP)
			//	println(v.PhysicalSig.nodeImp)
			//	println("--------")
			//	println(o.OP)
			//	println(o.PhysicalSig.nodeImp)
			//}
			mMap[k] = v
		}
	}

	if p.RChild != nil && p.visitChildren() {
		rMap := p.RChild.GetLogicalMap()
		for k, v := range rMap {
			// todo fix conflict
			//o, has := mMap[k]
			//if has {
			//	println(v.OP)
			//	println(v.PhysicalSig.nodeImp)
			//	println("--------")
			//	println(o.OP)
			//	println(o.PhysicalSig.nodeImp)
			//}
			mMap[k] = v
		}
	}
	return mMap
}

func (p *Node) IsMergeHashJoin() bool {
	return p.OP == "HashJoin" || p.OP == "MergeJoin"
}

// Stream_Sort/Merge_Sort
// PartialAgg/PartialTopN/PartialLimit
func (p *Node) RemoveSpecialNode() *Node {
	newNode := p
	switch p.OP {
		case "TableReader", "IndexLookUp", "IndexReader":
			if newNode.LChild.NodeType == "Agg" {
				newNode.LChild = p.LChild.LChild
				newNode.HasSpecialAgg = true
				newNode.LSpecialNode = p.LChild
			}
			if newNode.LChild.NodeType == "TopN" {
				newNode.LChild = p.LChild.LChild
				newNode.HasSpecialTopN = true
				newNode.LSpecialNode = p.LChild
			}
			if newNode.LChild.NodeType == "Limit" {
				newNode.LChild = p.LChild.LChild
				newNode.HasSpecialTopN = true
				newNode.LSpecialNode = p.LChild
			}
		case "MergeJoin":
			if newNode.LChild.OP == "Sort" {
				newNode.HasSpecialSort = true
				newNode.LChild = p.LChild.LChild.RemoveSpecialNode()
				newNode.LSpecialNode = p.LChild
			} else {
				newNode.LChild = p.LChild.RemoveSpecialNode()
			}
			if newNode.RChild.OP == "Sort"{
				newNode.HasSpecialSort = true
				newNode.RChild = p.RChild.LChild.RemoveSpecialNode()
				newNode.RSpecialNode = p.LChild
			} else {
				newNode.RChild = p.RChild.RemoveSpecialNode()
			}
		case "StreamAgg":
			if newNode.LChild.OP == "Sort" {
				newNode.HasSpecialSort = true
				newNode.LChild = p.LChild.LChild.RemoveSpecialNode()
				newNode.LSpecialNode = p.LChild
			} else {
				newNode.LChild = p.LChild.RemoveSpecialNode()
			}
		default:
			if p.LChild != nil {
				newNode.LChild = p.LChild.RemoveSpecialNode()
			}
			if p.RChild != nil {
				newNode.RChild = p.RChild.RemoveSpecialNode()
			}
	}
	return newNode
}

// Handle The Sort Case
func (p *Node) SortRelatedNode() []*Node {
	rs := make([]*Node, 0 ,1)
	switch p.OP {
		case "Sort", "IndexRangeScan", "TableRangScan", "TopN":
			rs = append(rs, p)
		default:
			if p.LChild != nil {
				rs = append(rs, p.LChild.SortRelatedNode()...)
			}
			if p.RChild != nil {
				rs = append(rs, p.RChild.SortRelatedNode()...)
			}
	}
	return rs
}

