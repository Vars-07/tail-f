package main

// Manager struct
type Manager struct {
	clientMap map[string][]string
	mapEntry  []*ClientEntry
}

// ClientEntry Struct
type ClientEntry struct {
	ClientUUID   string
	writeChannel chan string
}
