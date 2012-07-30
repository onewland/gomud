package mud

import ("time")

// Map of [type name] -> [field names that persist]
var PersistentKeys map[string][]string

func init() {
	PersistentKeys = make(map[string][]string)
}

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