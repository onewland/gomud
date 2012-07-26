package mud

import ("time")

type Persister interface {
	DBFullName() string
	PersistentValues() map[string]interface{}
	Save() string
}

func (u *Universe) HandlePersist() {
	for {
		for _,p := range(u.Persistents) {
			p.Save()
		}
		time.Sleep(300 * time.Millisecond)
	}
}