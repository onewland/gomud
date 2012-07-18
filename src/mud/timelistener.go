package mud

type TimeListener interface {
	Ping() chan int
}