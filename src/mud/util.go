package mud

import "regexp"

func SplitCommandString(cmd string) []string {
	re, _ := regexp.Compile(`(\S+)|(['"][^'"]+['"])`)
	return re.FindAllString(cmd, 10)
}

func Divider() string { 
	return "\n-----------------------------------------------------------\n"
}

func PlayersAsPhysObjSlice(ps map[int]Player) []PhysicalObject {
	physObjs := make([]PhysicalObject, len(ps))
	n := 0
	for _, p := range(ps) { 
		physObjs[n] = p 
		n++
	}
	return physObjs
}