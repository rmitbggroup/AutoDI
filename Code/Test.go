package main

import (
	"ExplainPlan/Explain"
	"ExplainPlan/Plan"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)


func testOneQuery(qPath string) {
	plan := Plan.GetPlanFromFile(qPath, false)
	// Remove useless added sort, topn, limit
	newN1 := plan.RemoveSpecialNode()
	newN1.GetLogicalSignature()
	newN1.GetPhysicalSignature()
	newN1.GetLogicalMap()
}

func TestWorkload(planDir string) {
	files, _ := ioutil.ReadDir(planDir)
	for _, f := range files {
		if !(f.Name() == ".DS_Store") {
			println(f.Name())
			testOneQuery(planDir + f.Name())
		}
	}
}

func OutputJoinOrder(path string) {
	plan1 := Plan.GetPlanFromFile(path, true)
	plan1 = plan1.RemoveProjectionAndSelection()
	plan1.GetLogicalSignature()
	plan1.GetPhysicalSignature()
	println(plan1.ObtainJoinPath())
}

func TestFindDiff(path1, path2 string, queryID string, diffThreshold float64) []Explain.DiffItem {
	plan1 := Plan.GetPlanFromFile(path1, false)
	plan1 = plan1.RemoveProjectionAndSelection()
	plan1.UpdateRowCountInfo()
	plan1.GetLogicalSignature()
	plan1.GetPhysicalSignature()

	plan2 := Plan.GetPlanFromFile(path2, false)
	plan2 = plan2.RemoveProjectionAndSelection()
	plan1.UpdateRowCountInfo()
	plan2.GetLogicalSignature()
	plan2.GetPhysicalSignature()
	mMap := plan2.GetLogicalMap()
	diffs := Explain.FindDiff(plan1, plan2, mMap, false)
	overallDiff := Explain.ParseTime4NS(plan1.Runtime) - Explain.ParseTime4NS(plan2.Runtime)
	_, _ = Explain.Inference(queryID, diffs, diffThreshold, overallDiff)
	return nil
}

var jobTblNames = []string{
	"aka_name",
	"aka_title",
	"cast_info",
	"char_name",
	"comp_cast_type",
	"company_name",
	"company_type",
	"complete_cast",
	"info_type",
	"keyword",
	"kind_type",
	"link_type",
	"movie_companies",
	"movie_info",
	"movie_info_idx",
	"movie_keyword",
	"movie_link",
	"name",
	"person_info",
	"role_type",
	"title",
}

func GenPlanFromPath(qPath, pPath string, explainOnly bool) {
	bytes, err := ioutil.ReadFile(qPath)
	if err != nil {
		return
	}
	sqlss := string(bytes)
	sqlsa := strings.Split(sqlss, ";")
	for _, query := range sqlsa {
		if len(query) > 6 { // Must have `select'
			query = strings.TrimSpace(query)+";"
			query = Plan.GenQueryWithHint(query, "nth_plan(2)")
			Plan.WritePlanInfoFile(query, pPath, explainOnly)
			//Plan.GetPlanFromExplain(query)
		}

	}
}

func GenPlanFromDir(oDir, nDir , version string, explainOnly bool) {
	files, _ := ioutil.ReadDir(oDir)
	for _, f := range files {
		if f.Name() != ".DS_Store" && f.Name() != "q15.sql" && f.Name() != "q2.sql" && f.Name() != "q22.sql"{
			println(f.Name())
			GenPlanFromPath(oDir+f.Name(), nDir+"plan"+f.Name()+"-"+version+".txt", explainOnly)
		}
	}
}

const sp byte = 226

var head = []string{"id", "estRows", "estCost", "actRows", "task", "access object", "execution info", "operator info", "memory", "disk"}

func FormThePlan(filePath string, explainOnly bool) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	//line, err := buf.ReadString('\n')
	//if err != nil && err != io.EOF{
	//	fmt.Println("Read file error!", err)
	//	return
	//}
	explainRS := make([]string, 0, 10)

	//line = strings.Replace(line, "\n", "", -1)
	//explainRS = append(explainRS, line)
	//rows := strings.Split(line, ";")
	colSize := 10
	if explainOnly {
		colSize = 5
	}
	colLen := make([]int, 0, colSize)
	for i := 0; i < colSize ; i++ {
		colLen = append(colLen, len(head[i]))
	}
	specialCount := make([]int, 0, 4)
	isFinished := false
	maxSC := 0
	explainRS = append(explainRS, strings.Join(head, ";"))
	specialCount = append(specialCount, 0)
	for !isFinished {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				isFinished = true
			} else {
				fmt.Println("Read file error!", err)
				return
			}
		}
		line = strings.Replace(line, "\n", "", -1)
		explainRS = append(explainRS, line)
		rows := strings.Split(line, ";")
		rawResult := make([][]byte, 0, colSize)

		for i := 0; i < colSize ; i++ {
			rawResult = append(rawResult, []byte(rows[i]))
		}
		spc := 0
		for i := 0; i < len(rawResult[0]); i++ {
			if rawResult[0][i] == sp {
				spc += 1
			}
		}
		maxSC = Explain.SelfMaxInt(maxSC, spc)
		specialCount = append(specialCount, spc)

		for i := 0; i < colSize; i++ {
			colLen[i] = Explain.SelfMaxInt(colLen[i], len(rows[i]))
		}
	}
	offsetI := make([]int, 0, colSize)
	for i := 0; i < colSize ; i++ {
		if i == 0 {
			offsetI = append(offsetI, 0)
		} else {
			offsetI = append(offsetI, offsetI[i-1]+colLen[i-1]+3)
		}
	}
	for i, explainRow := range explainRS {
		cols := strings.Split(explainRow, ";")
		curIndex := 0
		for j, col := range cols {
			if j > 0 {
				for curIndex < offsetI[j]-1-(maxSC-specialCount[i]-specialCount[i]) {
					print(" ")
					curIndex+=1
				}
			}
			print(col)
			curIndex += len(col)
		}
		println()
	}
}
