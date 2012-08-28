package mud

type FlexObjHandler func(*FlexContainer, interface{})

type FlexObjHandlerPair struct {
	Add FlexObjHandler
	Remove FlexObjHandler
}

var (
	FlexObjHandlers = make(map[string]FlexObjHandlerPair)
)

/*
 * FlexContainer sorts objects into separate lists,
 * which are stored in a map, using custom FlexObjHandler 
 * functions to sort.
 *
 * FlexObjHandlers can also react in ways that are unrelated
 * to AllObjects, for example in the way that rooms handle
 * PhysicalObject(s).
 */
type FlexContainer struct {
	handlers []FlexObjHandlerPair
	AllObjects map[string][]interface{}
	Meta map[string]interface{}
}

func (f *FlexContainer) Add(o interface{}) {
	for _,handler := range(f.handlers) {
		handler.Add(f,o)
	}
}

func (f *FlexContainer) Remove(o interface{}) {
	for _,handler := range(f.handlers) {
		handler.Remove(f,o)
	}
}

func (f *FlexContainer) AddObjToCategory(category string, o interface{}) {
	f.AllObjects[category] = append(f.AllObjects[category], o)
	Log("[add]",o,"to",category)
}

func (f *FlexContainer) RemoveObjFromCategory(category string, o interface{}) {
	for i,listO := range(f.AllObjects[category]) {
		if o == listO {
			if len(f.AllObjects[category]) > 1 {
				objs := append(f.AllObjects[category][:i],
					f.AllObjects[category][i+1:]...)
				f.AllObjects[category] = objs
			} else {
				f.AllObjects[category] = []interface{}{}
			}
			Log("[rm]",o,"from",category)
			break
		}
	}
}

func MakeFlexContainer(handlers ...string) *FlexContainer {
	fc := new(FlexContainer)
	fc.AllObjects = make(map[string][]interface{})
	fc.Meta = make(map[string]interface{})
	for _, handlerName := range(handlers) {
		fc.handlers = append(fc.handlers, FlexObjHandlers[handlerName])
	}
	return fc
}