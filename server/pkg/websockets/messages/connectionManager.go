package messages

import (
	"sync"

	"github.com/gorilla/websocket"
)

var connectionManager = &ConnectionManager{}

// ConnectionManager gerencia conexões WebSocket ativas.
type ConnectionManager struct {
	connections sync.Map // Map thread-safe.
}

// AddConnection adiciona uma conexão para o usuário.
func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn) {
	cm.connections.Store(userID, conn)
}

// GetConnection retorna a conexão de um usuário.
func (cm *ConnectionManager) GetConnection(userID int64) (*websocket.Conn, bool) {
	conn, ok := cm.connections.Load(userID)
	if ok {
		return conn.(*websocket.Conn), true
	}
	return nil, false
}

// RemoveConnection remove a conexão de um usuário.
func (cm *ConnectionManager) RemoveConnection(userID int64) {
	cm.connections.Delete(userID)
}
