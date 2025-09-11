package inmemory

type Hashtable interface {
	Set(string, string)
	Get(string) (string, bool)
	Del(string)
}

type hashtable struct {
	data map[string]string
}

func NewHashtable() Hashtable {
	data := make(map[string]string)
	return &hashtable{data: data}
}

func (ht *hashtable) Set(key, value string) {
	ht.data[key] = value
}

func (ht *hashtable) Get(key string) (string, bool) {
	value, found := ht.data[key]
	return value, found
}

func (ht *hashtable) Del(key string) {
	delete(ht.data, key)
}
