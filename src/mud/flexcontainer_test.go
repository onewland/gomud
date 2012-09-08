package mud

import "testing"

func isVowel(s string) bool {
	switch s {
	case "a", "e", "i", "o", "u":
		Log("isVowel is true")
		return true
	}
	return false
}

func init() {
	vowelCounter := new(FlexObjHandlerPair)
	vowelCounter.Add = func(fc *FlexContainer, o interface{}) {
		if s, ok := o.(string); ok && isVowel(s) {
			vowelCount := fc.Meta["vowelCount"].(int)
			vowelCount += 1
			fc.Meta["vowelCount"] = vowelCount
		}
	}
	vowelCounter.Remove = func(fc *FlexContainer, o interface{}) {
		if s, ok := o.(string); ok && isVowel(s) {
			vowelCount := fc.Meta["vowelCount"].(int)
			vowelCount -= 1
			fc.Meta["vowelCount"] = vowelCount
		}
	}
	FlexObjHandlers["VowelCounter"] = *vowelCounter
}

func categoryContainsString(c *FlexContainer, cat string, val string) bool {
	vals := c.AllObjects[cat]
	for _,v := range(vals) { 
		if v == val { return true }
	}
	return false
}

func assertContainsString(t *testing.T, c *FlexContainer, cat string, val string) {
	if !categoryContainsString(c, cat, val) {
		t.Errorf("FlexContainer should contain '%s' in '%s'",val,cat)
	}
}

func assertNotContainsString(t *testing.T, c *FlexContainer, cat string, val string) {
	if categoryContainsString(c, cat, val) {
		t.Errorf("FlexContainer should not contain '%s' in '%s'",val,cat)
	}
}

func assertVowelCount(t *testing.T, c *FlexContainer, count int) {
	i, ok := c.Meta["vowelCount"].(int)
	if !ok {
		t.Errorf("FlexContainer error recovering vowelCount")
	}
	if i != count {
		t.Errorf("FlexContainer should contain %d vowel(s), contains %d",
			count,
			i)
	}
}

func TestFlexContainerAddRemoveToCat(t *testing.T) {
	c := NewFlexContainer()
	c.AddObjToCategory("vowels","a")
	c.AddObjToCategory("vowels","e")
	c.AddObjToCategory("vowels","i")
	c.AddObjToCategory("vowels","o")
	c.AddObjToCategory("vowels","u")
	c.AddObjToCategory("vowels","y")
	assertContainsString(t, c, "vowels", "a")
	assertContainsString(t, c, "vowels", "e")
	assertContainsString(t, c, "vowels", "i")
	assertContainsString(t, c, "vowels", "o")
	assertContainsString(t, c, "vowels", "u")
	c.RemoveObjFromCategory("vowels","y")
	assertNotContainsString(t, c, "vowels", "y")
}

func TestFlexContainerAddCallsVowelCounter(t *testing.T) {
	c := NewFlexContainer("VowelCounter")
	c.Meta["vowelCount"] = 0
	c.Add("a")
	c.Add("u")
	assertVowelCount(t,c,2)
}

func TestFlexContainerRemoveCallsVowelCounter(t *testing.T) {
	c := NewFlexContainer("VowelCounter")
	c.Meta["vowelCount"] = 0
	c.Add("a")
	c.Remove("a")
	assertVowelCount(t,c,0)
}