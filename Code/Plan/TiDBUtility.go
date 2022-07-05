package Plan

import (
	"bufio"
	"container/list"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)
import "database/sql"

// We meet a sp at i-th, then we jump to check the (i+2)-th
const sp byte = 226
const c1 byte = 130 // '│'
const c2 byte = 156 // '├'
const c3 byte = 148 // '└'
const c4 byte = 128 // '─'
const empty byte = 32 // ' '

func isFirstChap(c byte) bool {
	return c >= 65 && c <= 90
}

func getNodeType(op string) string {
	var nodeType = ""
	switch op {
		case "HashJoin", "IndexJoin", "IndexHashJoin", "IndexMergeJoin", "MergeJoin":
			nodeType = "Join"
		case "StreamAgg", "HashAgg":
			nodeType = "Agg"
		case "TableReader", "IndexReader", "IndexLookUp":
			nodeType = "Reader"
		default:
			nodeType = op

	}
	return nodeType
}

func isNumber(c byte) bool {
	return c >= 48 && c <= 57
}

func getOP(oOp []byte) (string, string) {
	var start = -1
	var end = -1
	for i := 0; i < len(oOp); i++ {
		if start == -1 && isFirstChap(oOp[i]) {
			start = i
		}
		if end == -1 && oOp[i] == 95 {
			end = i
		}
	}
	var end2 = len(oOp)
	for i:= end+1; i < len(oOp); i++ {
		if !isNumber(oOp[i]) {
			end2 = i
			break
		}
	}
	nOp := oOp[start:end]
	id := oOp[end+1:end2]
	return string(nOp), string(id)
}

type item4GenPlan struct {
	node *Node
	sp int
}

type Stack struct {
	list *list.List
}

func NewStack() *Stack {
	list := list.New()
	return &Stack{list}
}

func (stack *Stack) Push(value interface{}) {
	stack.list.PushBack(value)
}

func (stack *Stack) Pop() interface{} {
	if e := stack.list.Back(); e!= nil {
		stack.list.Remove(e)
		return e.Value
	}

	return nil
}

func (stack *Stack) Len() int {
	return stack.list.Len()
}

func (stack *Stack) Empty() bool {
	return stack.Len() == 0
}

func getNodeFromOneRow(raw [][]byte, explainOnly bool) *Node {
	op, id := getOP(raw[0])
	node := &Node{OP:op, ID: id}
	node.NodeType = getNodeType(node.OP)
	node.EstRows, _ = strconv.ParseFloat(string(raw[1]), 64)
	node.RightEstRows = -1.0
	if explainOnly {
		node.TblName = string(raw[3])
		node.OperatorInfo = string(raw[4])
	} else {
		node.Cost, _ = strconv.ParseFloat(string(raw[2]), 64)
		node.ActRows, _ = strconv.ParseFloat(string(raw[3]), 64)
		node.OperatorInfo = string(raw[7])
		node.TblName = string(raw[5])
		// execution info can be empty
		_index := strings.Index(string(raw[6]), "time")
		if _index > -1 {
			_index1 := _index + 5
			_index2 := strings.Index(string(raw[6]), ",")
			node.Runtime = string(raw[6])[_index1: _index2]
		} else {
			node.Runtime = "0ns"
		}
	}




	return node
}

func WritePlanInfoFile(query string, filePath string, explainOnly bool) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	command := "explain analyze "
	if explainOnly {
		command = "explain "
	}
	re, err := db.Query(command + query)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	colSize := 10
	if explainOnly {
		colSize = 5
	}
	dest := make([]interface{}, colSize)
	rawResult := make([][]byte, colSize)
	for i, _ := range rawResult {
		dest[i] = &rawResult[i]
	}
	plan := make([]string, 0, 100)
	for re.Next() {
		err = re.Scan(dest...)
		if err != nil {
			panic(err)
		}
		row := make([]string, 0, colSize)
		for i := 0; i < colSize; i++ {
			row = append(row, string(rawResult[i]))
		}
		plan = append(plan, strings.Join(row,";"))
	}
	err = ioutil.WriteFile(filePath, []byte(strings.Join(plan, "\n")), 0644)
	if err != nil {
		panic(err)
	}
}

func PrintPlanFromFile(filePath string) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	sqlss := string(bytes)
	println(sqlss)
}

func GetPlanFromFile(filePath string, explainOnly bool) *Node {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return nil
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	line, err := buf.ReadString('\n')
	if err != nil && err != io.EOF{
		fmt.Println("Read file error!", err)
		return nil
	}
	line = strings.Replace(line, "\n", "", -1)
	row := strings.Split(line, ";")
	colSize := 10
	if explainOnly {
		colSize = 5
	}
	rawResult := make([][]byte, 0, colSize)
	for i := 0; i < colSize ; i++ {
		rawResult = append(rawResult, []byte(row[i]))
	}
	root := getNodeFromOneRow(rawResult, explainOnly)
	cur := root
	s := NewStack()
	// the first level nodes/children with 0 space.
	lastSC := -1
	isFinished := false
	for !isFinished {
		line, err = buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				isFinished = true
			} else {
				fmt.Println("Read file error!", err)
				return nil
			}
		}
		line = strings.Replace(line, "\n", "", -1)
		row = strings.Split(line, ";")
		rawResult = make([][]byte, 0, colSize)
		for i := 0; i < colSize ; i++ {
			rawResult = append(rawResult, []byte(row[i]))
		}
		var Sp byte = 0
		var spaceCount = 0
		for i := 0; i < len(rawResult[0]); {
			switch rawResult[0][i] {
			case sp: i+=2
			case empty: spaceCount+=1; i+=1
			case c2, c3:
				Sp = rawResult[0][i]; i+=1
			case c4:
				if Sp == c2 {
					rc := getNodeFromOneRow(rawResult, explainOnly)
					cur.RChild = rc
					_it := item4GenPlan{node: cur, sp: lastSC}
					s.Push(_it)
					cur = rc
					lastSC = spaceCount
				}
				if Sp == c3 {
					lc := getNodeFromOneRow(rawResult, explainOnly)
					if spaceCount > lastSC {
						cur.LChild = lc
						_it := item4GenPlan{node: cur, sp: lastSC}
						s.Push(_it)
						cur = lc
						lastSC = spaceCount
					} else if spaceCount == lastSC {
						v := s.Pop().(item4GenPlan)
						v.node.LChild = lc
						s.Push(v)
						cur = lc
						lastSC = spaceCount
					} else {
						var v item4GenPlan
						v = s.Pop().(item4GenPlan)
						for v.sp >= spaceCount {
							v = s.Pop().(item4GenPlan)
						}
						v.node.LChild = lc
						s.Push(v)
						cur = lc
						lastSC = spaceCount
					}
				}
				i+=1
			case c1: i+=1
			default:
				i = len(rawResult[0])

			}
		}

	}
	return root
}

// id-estRows-actRows-task-access_object-execution_info-operator_info-memory-disk
func GetPlanFromExplain(query string, explainOnly bool) *Node{
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	command := "explain analyze "
	if explainOnly {
		command = "explain "
	}
	re, err := db.Query(command + query)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	colSize := 10
	if explainOnly {
		colSize = 5
	}
	dest := make([]interface{}, colSize)
	rawResult := make([][]byte, colSize)
	for i, _ := range rawResult {
		dest[i] = &rawResult[i]
	}

	re.Next()
	err = re.Scan(dest...)
	if err != nil {
		panic(err)
	}
	root := getNodeFromOneRow(rawResult, explainOnly)
	cur := root
	s := NewStack()
	// the first level nodes/children with 0 space.
	lastSC := -1
	for re.Next() {
		re.Scan(dest...)
		var Sp byte = 0
		var spaceCount = 0
		for i := 0; i < len(rawResult[0]); {
			switch rawResult[0][i] {
				case sp: i+=2
				case empty: spaceCount+=1; i+=1
				case c2, c3:
					Sp = rawResult[0][i]; i+=1
				case c4:
					if Sp == c2 {
						rc := getNodeFromOneRow(rawResult, explainOnly)
						cur.RChild = rc
						_it := item4GenPlan{node: cur, sp: lastSC}
						s.Push(_it)
						cur = rc
						lastSC = spaceCount
					}
					if Sp == c3 {
						lc := getNodeFromOneRow(rawResult, explainOnly)
						if spaceCount > lastSC {
							cur.LChild = lc
							_it := item4GenPlan{node: cur, sp: lastSC}
							s.Push(_it)
							cur = lc
							lastSC = spaceCount
						} else if spaceCount == lastSC {
							v := s.Pop().(item4GenPlan)
							v.node.LChild = lc
							s.Push(v)
							cur = lc
							lastSC = spaceCount
						} else {
							var v item4GenPlan
							v = s.Pop().(item4GenPlan)
							for v.sp >= spaceCount {
								v = s.Pop().(item4GenPlan)
							}
							v.node.LChild = lc
							s.Push(v)
							cur = lc
							lastSC = spaceCount
						}
					}
					i+=1
				case c1: i+=1
				default:
					i = len(rawResult[0])

			}
		}
	}
	return root
}

func ObtainQueryHint(query string) string {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	re, err := db.Query("explain format='hint' " + query)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	dest := make([]interface{}, 1)
	rawResult := make([][]byte, 1)
	for i, _ := range rawResult {
		dest[i] = &rawResult[i]
	}

	re.Next()
	re.Scan(dest...)
	return string(rawResult[0])
}

func GenQueryWithHint(query, hint string) string {
	return  "select /*+ " + hint + " */ " + query[6:]
}