package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

var connectionManager = &ConnectionManager{}

// Gerenciador de conexões WebSocket (thread-safe)
type ConnectionManager struct {
	connections sync.Map // Armazena conexões WebSocket por ID de usuário
}

func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn) {
	cm.connections.Store(userID, conn)
}

func (cm *ConnectionManager) GetConnection(userID int64) (*websocket.Conn, bool) {
	conn, ok := cm.connections.Load(userID)
	if ok {
		return conn.(*websocket.Conn), true
	}
	return nil, false
}

func (cm *ConnectionManager) RemoveConnection(userID int) {
	cm.connections.Delete(userID)
}
