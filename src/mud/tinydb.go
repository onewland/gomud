package mud

import ("redis"
	"strconv"
	"strings"
	"fmt")

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

func (t *TinyDB) RedisSet(k string, v interface{}) {
	switch ty := v.(type) {
	case string:
		vbyte := []byte(ty)
		t.dbConn.Set(k,vbyte)
	case []Persister:
		t.dbConn.Del(k)
		for _,p := range(ty) {
			t.dbConn.Sadd(k, []byte(p.DBFullName()))
		}
	default:
		panic(fmt.Sprintf("Unrecognized interface %T in RedisSet",ty))
	}
}

func (t *TinyDB) LoadStructure(fullyQualifiedID string) (interface{}, string) {
	return t, ""
}

func (t *TinyDB) SaveStructure(className string, vals map[string]interface{}) string {
	var returnId string
	if theId, ok := vals["id"].(string); ok && theId != "" {
		// Already exists, just update data
		returnId = theId
	} else {
		// Insert anew
		newId, err := t.dbConn.Incr(FieldJoin(":",className,"idCounter"))
		if(err == nil) {
			newIdStr := strconv.Itoa(int(newId))
			vals["id"] = newIdStr
			returnId = newIdStr
		} else {
			panic(err)
		}
	}

	for k,v := range(vals) {
		t.RedisSet(FieldJoin(":",className,returnId,k),v)
	}
	
	return returnId
}

func (t *TinyDB) Flush() {
	t.dbConn.Flushdb()
}