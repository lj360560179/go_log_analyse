package main

import (
	"time"
	"strings"
	"os"
	"bufio"
	"io"
	"regexp"
	"log"
	"net/url"
	"strconv"
	"github.com/influxdata/influxdb/client/v2"
	"flag"
	"net/http"
	"encoding/json"
)

type Reader interface {
	Read(rc chan []byte)
}

type ReadFromFile struct {
	path string
}

func (r *ReadFromFile) Read(rc chan []byte) {
	f, err := os.Open(r.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Seek(0, 2)
	buf := bufio.NewReader(f)

	for {
		line, err := buf.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
		} else if err != nil {
			panic(err)
		} else {
			rc <- line[:len(line)-1]
		}
	}
}

type Writer interface {
	Write(wc chan *Log)
}

type WriteToInfluxdb struct {
	dsn string
}

func (w *WriteToInfluxdb) Write(wc chan *Log) {
	dsnSli := strings.Split(w.dsn, "@")

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     dsnSli[0],
		Username: dsnSli[1],
		Password: dsnSli[2],
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dsnSli[3],
		Precision: dsnSli[4],
	})
	if err != nil {
		log.Fatal(err)
	}

	for v := range wc {
		tags := map[string]string{
			"Path": v.Path,
			"Method": v.Method,
			"Scheme": v.Scheme,
			"Status": v.Status,
		}
		fields := map[string]interface{}{
			"bytesSent":   v.BytesSent,
			"upstreamTime": v.UpstreamTime,
			"RequestTime":   v.RequestTime,
		}

		pt, err := client.NewPoint("log", tags, fields, v.TimeLocal)
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

		if err := c.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

type Log struct {
	TimeLocal                    time.Time
	BytesSent                    int
	Path, Method, Scheme, Status string
	UpstreamTime, RequestTime    float64
}

type LogProcess struct {
	rc chan []byte
	wc chan *Log
	r  Reader
	w  Writer
}

func (l *LogProcess) Process() {
	re := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(
\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([d\.-]+)`)

	loc, _ := time.LoadLocation("PRC")
	for v := range l.rc {
		str := string(v)
		ret := re.FindStringSubmatch(str)
		if len(ret) != 14 {
			log.Println(str)
			continue
		}

		msg := &Log{}
		//buzdweimaoshi 2006
		t, err := time.ParseInLocation("02/Jan/2006:15:04:05 +0000", ret[4], loc)
		if err != nil {
			log.Println(ret[4])
		}
		msg.TimeLocal = t

		byteSent, _ := strconv.Atoi(ret[8])
		msg.BytesSent = byteSent

		reqSli := strings.Split(ret[6], " ")
		if len(reqSli) != 3 {
			log.Println(ret[6])
			continue
		}
		msg.Method = reqSli[0]
		msg.Scheme = reqSli[2]
		u, err := url.Parse(reqSli[1])
		if err != nil {
			log.Println(reqSli[1])
			continue
		}
		msg.Path = u.Path
		msg.Status = ret[7]
		upTime, _ := strconv.ParseFloat(ret[12], 64)
		reqTime, _ := strconv.ParseFloat(ret[13], 64)
		msg.UpstreamTime = upTime
		msg.RequestTime = reqTime

		l.wc <- msg
	}
}

type SystemInfo struct {
	LogLine int `json:"logline"` // 总数
	Tps float64 `json:"tps"`
	ReadChanLen int `json:"readchanlen"` // read chan 长度
	WriteChanLen int `json:"writechanlen"` // write chan 长度
	RunTime string `json:"runtime"` // 运行总时间
	ErrNum int `json:"errnum"` // 错误数
}

type Monitor struct {
	startTime time.Time
	data SystemInfo
}

func (m *Monitor) start(lp *LogProcess) {
	http.HandleFunc("/monitor", func(writer http.ResponseWriter, request *http.Request) {
		m.data.RunTime = time.Now().Sub(m.startTime).String()
		m.data.ReadChanLen = len(lp.rc)
		m.data.WriteChanLen = len(lp.wc)

		ret, _ := json.MarshalIndent(m.data, "", "\t")

		io.WriteString(writer, string(ret))
	})

	http.ListenAndServe(":9091", nil)
}

func main() {
	var path, dsn string
	flag.StringVar(&path, "path", "./log.log", "file path")
	flag.StringVar(&dsn, "dsn", "http://localhost:8086@log@logpass@log@s", "influxdb dsn")
	flag.Parse()


	r := &ReadFromFile{
		path: path,
	}

	w := &WriteToInfluxdb{
		dsn: dsn,
	}

	l := &LogProcess{
		rc: make(chan []byte, 200),
		wc: make(chan *Log),
		r:  r,
		w:  w,
	}
	
	go l.r.Read(l.rc)
	for i := 0; i < 2; i++ {
		go l.Process()
	}
	for i := 0; i < 2; i++ {
		go l.w.Write(l.wc)
	}

	m := &Monitor{
		startTime: time.Now(),
		data: SystemInfo{},
	}
	m.start(l)

}
