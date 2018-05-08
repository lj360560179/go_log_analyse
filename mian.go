package main

import (
	"fmt"
	"strings"
)

type LogProcess struct {
	path  string // 日志路径
	dsn string   // infludb dsn
}

func (lp *LogProcess) Read(){
	path := lp.path
	fmt.Println(path)
}

func (lp *LogProcess) Process() {
	log := "hello world"
	fmt.Println(strings.ToUpper(log))
}

func (lp *LogProcess) Write()  {
	dsn := lp.dsn
	fmt.Println(dsn)
}

func main() {

	lp := &LogProcess{
		path: "test path",
		dsn: "test dsn",
	}

	// read log file
	lp.Read()
	// process log
	lp.Process()
	// write data
	lp.Write()
}

