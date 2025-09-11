package inmemory

type Hashtable struct {
	data map[string]string
}

func NewHashtable() *Hashtable {
	data := make(map[string]string)
	return &Hashtable{data: data}
}

func (ht *Hashtable) Set(key, value string) {
	ht.data[key] = value
}

func (ht *Hashtable) Get(key string) (string, bool) {
	value, found := ht.data[key]
	return value, found
}

func (ht *Hashtable) Del(key string) {
	delete(ht.data, key)
}
