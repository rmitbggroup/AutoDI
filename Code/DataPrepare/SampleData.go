package DataPrepare

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

func obtainRandomP() float32 {
	rand.Seed(time.Now().UnixNano())
	_ri := rand.Intn(1000)
	_rf := float32(_ri)
	return _rf/1000.0
}

func SampleDataForDir(rate float32, oPath string, nPath string) {
	files, _ := ioutil.ReadDir(oPath)
	for _, f := range files {
		if !(f.Name() == ".DS_Store") {
			println(f.Name())
			SampleDataForFile(rate, oPath+f.Name(), nPath+f.Name())
		}

	}
}

func RandomGenerateData(nPath string) {

}

func SampleDataForFile(rate float32, oPath string, nPath string) {
	file, err := os.OpenFile(oPath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return
	}
	defer file.Close()

	TOTALLINE := 0
	newFile := make([]string, 0 , 10000)
	buf := bufio.NewReader(file)
	line, err := buf.ReadString('\n')
	//line = line[:len(line)-1]
	TOTALLINE += 1
	if err != nil && err != io.EOF{
		fmt.Println("Read file error!", err)
		return
	}

	p := obtainRandomP()
	if p < rate {
		newFile = append(newFile, line)
	}

	isFinished := false
	for !isFinished {
		//if TOTALLINE % 10000 == 0 {
		//	println(TOTALLINE)
		//}
		line, err = buf.ReadString('\n')
		TOTALLINE += 1
		if err != nil {
			if err == io.EOF {
				isFinished = true
			} else {
				fmt.Println("Read file error!", err)
				return
			}
		}
		//line = line[:len(line)-1]
		//println(line)
		p := obtainRandomP()
		if p < rate {
			newFile = append(newFile, line)
		}
	}
	println(len(newFile))
	println(TOTALLINE)
	println("--------")
	err = ioutil.WriteFile(nPath, []byte(strings.Join(newFile, "")), 0644)
	if err != nil {
		panic(err)
	}
}
