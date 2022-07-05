package main

import (
	"ExplainPlan/Explain"
	"ExplainPlan/Plan"
	"fmt"
	"io/ioutil"
	"time"
)

func findDiffOnly(p1Path, p2Path string) float64 {

	// Stage 1: Preprocessing
	p1 := Plan.GetPlanFromFile(p1Path, false)
	p2 := Plan.GetPlanFromFile(p2Path, false)

	i := 0
	var sum int64 = 0
	for i < 5 {
		start := time.Now()
		p1 = p1.RemoveProjectionAndSelection()
		p1 = p1.RemoveSpecialNode()
		p1 = p1.ChangeJoinSide()
		p1.GetLogicalSignature()
		p1.GetPhysicalSignature()
		p1.AddLevelInfo(0)

		p2 = p2.RemoveProjectionAndSelection()
		p2 = p2.RemoveSpecialNode()
		p2 = p2.ChangeJoinSide()
		p2.GetLogicalSignature()
		p2.GetPhysicalSignature()
		p2.AddLevelInfo(0)
		_ = p1.GetLogicalMap()
		mp := p2.GetLogicalMap()

		// Stage 2: FindDiff
		_ = Explain.FindDiff(p1, p2, mp, true)
		//for _, diff := range diffs {
		//	println(diff.DiffMessage)
		//}
		timeUsed := time.Since(start)
		sum += timeUsed.Microseconds()
		i++
	}
	return float64(sum)/float64(i)
}

func inferenceOnly(queryId, p1Path, p2Path string) float64 {

	// Stage 1: Preprocessing
	p1 := Plan.GetPlanFromFile(p1Path, false)
	p2 := Plan.GetPlanFromFile(p2Path, false)

	p1 = p1.RemoveProjectionAndSelection()
	p1 = p1.RemoveSpecialNode()
	p1 = p1.ChangeJoinSide()
	p1.GetLogicalSignature()
	p1.GetPhysicalSignature()
	p1.AddLevelInfo(0)

	p2 = p2.RemoveProjectionAndSelection()
	p2 = p2.RemoveSpecialNode()
	p2 = p2.ChangeJoinSide()
	p2.GetLogicalSignature()
	p2.GetPhysicalSignature()
	p2.AddLevelInfo(0)
	mp := p2.GetLogicalMap()

	// Stage 2: FindDiff
	diffs := Explain.FindDiff(p1, p2, mp, true)
	tr1 := p1.Runtime
	tr2 := p2.Runtime
	overall := Explain.ParseTime4NS(tr1) - Explain.ParseTime4NS(tr2)

	i := 0
	var sum int64 = 0

	ratio := 0.7
	for i < 5 {
		start := time.Now()
		//p1 = p1.RemoveProjectionAndSelection()
		//p1 = p1.RemoveSpecialNode()
		//p1 = p1.ChangeJoinSide()
		//p1.GetLogicalSignature()
		//p1.GetPhysicalSignature()
		//p1.AddLevelInfo(0)
		//
		//p2 = p2.RemoveProjectionAndSelection()
		//p2 = p2.RemoveSpecialNode()
		//p2 = p2.ChangeJoinSide()
		//p2.GetLogicalSignature()
		//p2.GetPhysicalSignature()
		//p2.AddLevelInfo(0)
		//mp := p2.GetLogicalMap()
		//
		//// Stage 2: FindDiff
		//diffs := Explain.FindDiff(p1, p2, mp, true)
		//tr1 := p1.Runtime
		//tr2 := p2.Runtime
		//overall := Explain.ParseTime4NS(tr1) - Explain.ParseTime4NS(tr2)
		_,_ = Explain.Inference(queryId+".sql", diffs, ratio, overall)
		timeUsed := time.Since(start)
		sum += timeUsed.Microseconds()
		i++
	}
	return float64(sum)/float64(i)
}

func enumerateTwoDir(dir1, dir2 string) {
	files, _ := ioutil.ReadDir(dir1)
	for _, f := range files {
		if !(f.Name() == ".DS_Store") {
			fmt.Printf("%s,%f\n",f.Name(),findDiffOnly(dir1+f.Name(), dir2+f.Name()))
		}
	}
}

func overallTest() {
	queryPath := "Data/tpch_queries/"
	planPath := "Data/tpch_explain_2th/"
	planVersion := "5.2.0"
	explainOnly := true
	testSQL := "SELECT  MIN(t.title) AS typical_european_movie FROM company_type AS ct,      info_type AS it,      movie_companies AS mc,      movie_info AS mi,      title AS t WHERE ct.kind = 'production companies'   AND mc.note LIKE '%(theatrical)%'   AND mc.note LIKE '%(France)%'   AND mi.info IN ('Sweden',                   'Norway',                   'Germany',                   'Denmark',                   'Swedish',                   'Denish',                   'Norwegian',                   'German')   AND t.production_year > 2005   AND t.id = mi.movie_id   AND t.id = mc.movie_id   AND mc.movie_id = mi.movie_id   AND ct.id = mc.company_type_id   AND it.id = mi.info_type_id;"
	//"SELECT MIN(mi.info) AS movie_budget,        MIN(mi_idx.info) AS movie_votes,        MIN(n.name) AS male_writer,        MIN(t.title) AS violent_movie_title FROM cast_info AS ci,      info_type AS it1,      info_type AS it2,      keyword AS k,      movie_info AS mi,      movie_info_idx AS mi_idx,      movie_keyword AS mk,      name AS n,      title AS t WHERE ci.note IN ('(writer)',                   '(head writer)',                   '(written by)',                   '(story)',                   '(story editor)')   AND it1.info = 'genres'   AND it2.info = 'votes'   AND k.keyword IN ('murder',                     'blood',                     'gore',                     'death',                     'female-nudity')   AND mi.info = 'Horror'   AND n.gender = 'm'   AND t.id = mi.movie_id   AND t.id = mi_idx.movie_id   AND t.id = ci.movie_id   AND t.id = mk.movie_id   AND ci.movie_id = mi.movie_id   AND ci.movie_id = mi_idx.movie_id   AND ci.movie_id = mk.movie_id   AND mi.movie_id = mi_idx.movie_id   AND mi.movie_id = mk.movie_id   AND mi_idx.movie_id = mk.movie_id   AND n.id = ci.person_id   AND it1.id = mi.info_type_id   AND it2.id = mi_idx.info_type_id   AND k.id = mk.keyword_id;"
	queryId := "33"+"a"
	_ = queryId
	p1Path := "Exp/job"+queryId+"_p1.txt"
	p2Path := "Exp/job"+queryId+"_p2.txt"

	_ = explainOnly
	_ = queryPath
	_ = planPath
	_ = planVersion
	_ = testSQL
	_ = p1Path
	_ = p2Path



	// Test Whole Framework
	// Stage 1: Preprocessing
	p1 := Plan.GetPlanFromFile(p1Path, false)
	p2 := Plan.GetPlanFromFile(p2Path, false)
	tc1 := p1.Cost
	tc2 := p2.Cost
	tr1 := p1.Runtime
	tr2 := p2.Runtime
	_ = tc1
	_ = tc2
	_ = tr1
	_ = tr2
	p1 = p1.RemoveProjectionAndSelection()
	p1 = p1.RemoveSpecialNode()
	p1 = p1.ChangeJoinSide()
	p1.GetLogicalSignature()
	p1.GetPhysicalSignature()
	p1.AddLevelInfo(0)

	p2 = p2.RemoveProjectionAndSelection()
	p2 = p2.RemoveSpecialNode()
	p2 = p2.ChangeJoinSide()
	p2.GetLogicalSignature()
	p2.GetPhysicalSignature()
	p2.AddLevelInfo(0)
	mp := p2.GetLogicalMap()
	// Stage 2: FindDiff
	diffs := Explain.FindDiff(p1, p2, mp, false)
	// Stage 3: Inference

	overall := Explain.ParseTime4NS(tr1) - Explain.ParseTime4NS(tr2)
	ratio := 0.7
	_ = overall
	_ = ratio
	mainIdx, rs := Explain.Inference(queryId+".sql", diffs, ratio,overall)
	// Stage 4: Print Report
	//Explain.PrintDiff(queryId+".sql", diffs, ratio, overall)
	Explain.PrintReport(queryId+".sql", diffs, mainIdx, rs, tc1, tc2, tr1, tr2)
}

func printPlanReadable(p1Path, p2Path string) {
	println("P1")
	FormThePlan(p1Path, false)
	println()
	println()
	println("P2")
	FormThePlan(p2Path, false)
}

func main() {
	queries := []string {"5", "10", "17", "9", "13", "15", "24", "25", "33"}
	for _, query := range queries {
		queryId := query+"a"
		_ = queryId
		p1Path := "Exp/job"+queryId+"_p1.txt"
		p2Path := "Exp/job"+queryId+"_p2.txt"
		fmt.Printf("%s,%f\n", queryId, inferenceOnly(queryId, p1Path, p2Path))
		//fmt.Printf("%s,%f\n", queryId, findDiffOnly(p1Path, p2Path))
	}
}
