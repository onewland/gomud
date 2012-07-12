package mud

var TimeListenerList []TimeListener

type TimeListener interface {
	Ping() chan int
}