package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func httpUpgrader(r *http.Request) bool {
	return true
}

// NewManager inits the manager
func NewManager() *Manager {
	return &Manager{
		clientMap: map[string][]string{},
		mapEntry:  []*ClientEntry{},
	}
}

// GetRouter Configure the routes for the service
func (m *Manager) GetRouter() *httprouter.Router {
	log.Infof("Initializing the router")
	router := httprouter.New()
	router.GET("/websocket/connect", m.webSocketConnectHandler)
	return router
}

func (m *Manager) webSocketConnectHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Infof("Configuring Client Connection")

	ctx := r.Context()
	upgrader := websocket.Upgrader{CheckOrigin: httpUpgrader}
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Errorf("Error: %s \n", err)
		return
	}

	defer ws.Close()
	ticker := time.NewTicker(10000)

	writeChan := make(chan string, 50)
	clientUUID, _ := uuid.NewV4()
	uuid := clientUUID.String()

	cE := &ClientEntry{
		ClientUUID:   uuid,
		writeChannel: writeChan,
	}

	go transmitLast10(writeChan)

	log.Infof("Client with uuid, %s, registered", uuid)

	m.mapEntry = append(m.mapEntry, cE)
	m.clientMap["/"] = append(m.clientMap["/"], uuid)

	for {
		select {
		case m := <-writeChan:
			err = ws.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Errorf("Error in writing to WS: %s", err)
			}

		case <-ticker.C:
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Errorf("Err in writing to WS: %s", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (m *Manager) fileWatch(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					m.publishToClient(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func (m *Manager) publishToClient(path string) {
	fileHandle, err := os.Open(path)

	if err != nil {
		panic("Cannot open file")
		os.Exit(1)
	}
	defer fileHandle.Close()

	line := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()
	for {
		cursor--
		fileHandle.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		fileHandle.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) {
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line)

		if cursor == -filesize {
			break
		}
	}

	for _, cl := range m.mapEntry {
		fmt.Println(cl.ClientUUID)
		cl.writeChannel <- line
	}
}

func transmitLast10(writeChan chan string) {
	fileHandle, err := os.Open("logger.log")
	if err != nil {
		panic("Cannot open file")
		os.Exit(1)
	}
	defer fileHandle.Close()

	line := ""
	finalLine := ""
	var cursor int64 = 0
	stat, _ := fileHandle.Stat()
	filesize := stat.Size()
	count := 1
	for {
		cursor--
		fileHandle.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		fileHandle.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) && count > 10 {
			break
		} else if cursor != -1 && (char[0] == 10 || char[0] == 13) && count <= 10 {
			finalLine = line
			count++
		}

		line = fmt.Sprintf("%s%s", string(char), line)
		if cursor == -filesize {
			finalLine = line
			break
		}
	}

	writeChan <- finalLine
}
