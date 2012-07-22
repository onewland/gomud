package mud

import ("redis"
	"strconv"
	"strings")

type TinyDB struct {
	dbConn redis.Client
}

func NewTinyDB(client redis.Client) *TinyDB {
	db := new(TinyDB)
	db.dbConn = client
	return db
}

func FieldJoin(sep string, args... string) string {
	return strings.Join(args, sep)
}

func (t *TinyDB) RedisSet(k string, v string) {
	vbyte := []byte(v)
	t.dbConn.Set(k,vbyte)
}

func (t *TinyDB) SaveStructure(className string, vals map[string]string) string {
	var returnId string
	if theId, ok := vals["id"]; ok && theId != "" {
		// Already exists, just update data
		returnId = theId
	} else {
		// Insert anew
		newId, err := t.dbConn.Incr(FieldJoin(":",className,"idCounter"))
		if(err == nil) {
			newIdStr := strconv.Itoa(int(newId))
			vals["id"] = newIdStr
			returnId = newIdStr
		}
	}

	for k,v := range(vals) {
		t.RedisSet(FieldJoin(":",className,returnId,k),v)
	}
	
	return returnId
}