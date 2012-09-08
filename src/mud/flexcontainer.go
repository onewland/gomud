package mud

/* 
 FlexObjHandler is used to create extensible behaviors for
 FlexContainer. In general, the FlexObjHandler function will
 use a type assertion to determine what "categories" an input 
 object belongs to so it can respond appropriately.

 FlexObjHandlers can also react in ways that are unrelated
 to AllObjects, for example in the way that rooms handle
 PhysicalObject(s).
 */
type FlexObjHandler func(*FlexContainer, interface{})

/*
 FlexObjHandlerPair(s) are FlexObjHandler(s) pairs for Add
 and Remove functions for FlexContainer.
 */
type FlexObjHandlerPair struct {
	Add FlexObjHandler
	Remove FlexObjHandler
}

var (
/*
 FlexObjHandlers is a global map of FlexObjHandlerPair names 
 to FlexObjHandlerPairs. These are the keys used in 
 MakeFlexContainer.
 */
	FlexObjHandlers = make(map[string]FlexObjHandlerPair)
)

/*
 FlexContainer files objects into separate lists,
 which are stored in a map named AllObjects, using
 custom FlexObjHandler functions to sort.

 Add and Remove delegate to FlexObjHandlerPairs, which
 for the most part will dispatch to AddObjToCategory
 and RemoveObjFromCategory.
 */
type FlexContainer struct {
	handlers []FlexObjHandlerPair
	AllObjects map[string][]interface{}
	Meta map[string]interface{}
}

/* 
 Add object to FlexContainer. It is possible that
 this function will do nothing (no handlers respond to it), 
 and if so that will not be indicated.
 */
func (f *FlexContainer) Add(o interface{}) {
	for _,handler := range(f.handlers) {
		handler.Add(f,o)
	}
}

/* 
 Remove object from FlexContainer. It is possible that
 this function will do nothing (no handlers, or object hasn't 
 been added), and if so that will not be indicated.
 */
func (f *FlexContainer) Remove(o interface{}) {
	for _,handler := range(f.handlers) {
		handler.Remove(f,o)
	}
}

/* 
 Add object to specific FlexContainer category. This
 always performs an action and does not trigger handlers
 (it is, in fact, used by handlers) 
 */
func (f *FlexContainer) AddObjToCategory(category string, o interface{}) {
	f.AllObjects[category] = append(f.AllObjects[category], o)
	Log("[add]",o,"to",category)
}

/* 
 Remove object from specific FlexContainer category. This
 always performs an action and does not trigger handlers
 (it is, in fact, used by handlers) 
*/
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

/* 
 Creates a new FlexContainer with handlers defined by in
 FlexObjHandlers by input strings 
 */ 
func NewFlexContainer(handlers ...string) *FlexContainer { 
	fc := new(FlexContainer) 
	fc.AllObjects = make(map[string][]interface{}) 
	fc.Meta = make(map[string]interface{})
	for _, handlerName := range(handlers) { 
		fc.handlers = append(fc.handlers, FlexObjHandlers[handlerName])
	} 
	return fc 
}