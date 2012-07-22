package mud

import ("time")

type Persister interface {
	PersistentValues() map[string]string
	Save() string
}

func (u *Universe) HandlePersist() {
	for {
		for _,p := range(u.Persistents) {
			go p.Save()
		}
		time.Sleep(300 * time.Millisecond)
	}
}