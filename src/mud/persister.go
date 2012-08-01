package mud

import ("time"; "strings"; "strconv";)

type Loader func(universe *Universe, id int) interface{}

var Loaders map[string]Loader
// Map of [type name] -> [field names that persist]
var PersistentKeys map[string][]string

func init() {
	PersistentKeys = make(map[string][]string)
	Loaders = make(map[string]Loader)
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

// Presumes Loaders has been populated with correct Loader, which especially
// for internal structures is not currently necessarily true
func LoadArbitrary(universe *Universe, fullDbUrl string) interface{} {
	args := strings.SplitN(fullDbUrl, ":", 2)
	dbType := args[0]
	dbIdInt, err := strconv.Atoi(args[1])
	if err != nil { 
		panic("Non-numeric ID in LoadArbitrary") 
	}
	loader := Loaders[dbType]
	return loader(universe, dbIdInt)
}