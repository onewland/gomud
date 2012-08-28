package mud

func ifIsTimeListener(o interface{}, ifTrue func(TimeListener)) {
	oAsTL, isTL := o.(TimeListener)
	
	if(isTL) { ifTrue(oAsTL) }
}

func init() {
	containerHelper := new(FlexObjHandlerPair)
	containerHelper.Add = func(fc *FlexContainer, o interface{}) {
		ifIsTimeListener(o, func(TimeListener) {
			fc.AddObjToCategory("TimeListeners",o)
		})
	}
	containerHelper.Remove = func(fc *FlexContainer, o interface{}) {
		ifIsTimeListener(o, func(TimeListener) {
			fc.RemoveObjFromCategory("TimeListeners",o)
		})
	}
	FlexObjHandlers["TimeListeners"] = *containerHelper
}

type TimeListener interface {
	Ping() chan int
}