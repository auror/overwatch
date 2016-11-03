package log

import (
	"log"
	"os"
)

var f *os.File

func init() {
//	var err error
//	f, err = os.OpenFile("/Users/sarath.prabath/Personal/Overwatch/overwatch.log", os.O_APPEND, 0666)
//	if err != nil {
//	    log.Fatalf("error opening file: %v", err)
//	}
//	log.SetOutput(f)
}

func Close() {
	f.Close()
}

func Info(v ...interface{}) {
	log.Println(v...)
}

func Debug(v ...interface{}) {
	log.Println(v...)
}

func Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Error(v ...interface{}) {
	log.Println(v...)
}

func Fatal(v ...interface{}) {
	log.Println(v...)
}
