package mud

import ("time"
	"fmt")

type Persister interface {
	PersistentValues() map[string]string
	Save() string
}

func HandlePersist(persistents []Persister) {
	fmt.Println("persistents = ", persistents)
	for {
		for _,p := range(persistents) {
			go p.Save()
		}
		time.Sleep(300 * time.Millisecond)
	}
}