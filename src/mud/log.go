package mud

import "log"

func Log(v ...interface{}) {
	log.Println(v...)
}