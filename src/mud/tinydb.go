package mud

import ("redis"
	"strconv"
	"strings"
	"fmt")

type Pvals map[string]interface{}
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

func (t *TinyDB) RedisGet(k string) (string, error) {
	bytes, error := t.dbConn.Get(k)
	str := string(bytes)
	return str, error
}

func (t *TinyDB) KeyExists(k string) (bool, error) {
	return t.dbConn.Exists(k)
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

func SMembersAsString(mems [][]byte) []string {
	members := make([]string,len(mems))
	for i,x := range(mems) { members[i] = string(x) }
	return members
}

func (t *TinyDB) LoadStructure(keys []string, fullDbUrl string) Pvals {
	vals := make(Pvals)
	for _,key := range(keys) {
		keyFull := FieldJoin(":",fullDbUrl,key)
		dbType, _ := t.dbConn.Type(keyFull)
		switch(dbType) {
		case redis.RT_STRING:
			val, _ := t.dbConn.Get(keyFull)
			vals[key] = string(val)
		case redis.RT_SET:
			mems, _ := t.dbConn.Smembers(keyFull)
			vals[key] = SMembersAsString(mems)
		case redis.RT_NONE:
			Log("[WARN] persistent key expected but not found:",keyFull)
		default:
			panic("Unrecognized value type in Redis middleware")
		}
	}
	return vals
}

func (t *TinyDB) AddToGlobalSet(key string, value string) {
	t.dbConn.Sadd(key, []byte(value))
}

func (t *TinyDB) RemoveFromGlobalSet(key string, value string) {
	t.dbConn.Srem(key, []byte(value))
}

func (t *TinyDB) GlobalSetGet(key string) []string {
	mems, _ := t.dbConn.Smembers(key)
	return SMembersAsString(mems)
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