package memory

import (
	"fmt"

	"github.com/SergeyShpak/owngame/server/src/model/layers"
	"github.com/SergeyShpak/owngame/server/src/types"
)

type memoryRoomLayer struct {
	rooms       *rooms
	roomPlayers *roomPlayers
}

func NewMemoryRoomLayer() (layers.RoomsDataLayer, error) {
	m := &memoryRoomLayer{
		rooms:       newRooms(),
		roomPlayers: newRoomPlayers(),
	}
	return m, nil
}

func (m *memoryRoomLayer) CreateRoom(r *types.RoomCreateRequest, roomToken string) error {
	rMeta := &roomMeta{
		Name:            r.Name,
		Password:        r.Password,
		MaxPlayersCount: 3,
	}
	if ok := m.rooms.PutRoomMeta(rMeta); !ok {
		return fmt.Errorf("room %s already exists", r.Name)
	}
	return nil
}

func (m *memoryRoomLayer) CheckPassword(roomName string, password string) error {
	meta, ok := m.rooms.GetRoomMeta(roomName)
	if !ok {
		return fmt.Errorf("rooms %s not found", roomName)
	}
	if meta.Password != password {
		return fmt.Errorf("password validation for room %s failed, expected: %s, actual: %s", roomName, meta.Password, password)
	}
	return nil
}

func (m *memoryRoomLayer) JoinRoom(roomName string, login string) (types.PlayerRole, error) {
	meta, ok := m.rooms.GetRoomMeta(roomName)
	if !ok {
		return types.PlayerRoleParticipant, fmt.Errorf("room %s not found", roomName)
	}
	if ok := m.roomPlayers.PutHost(roomName, login); ok {
		return types.PlayerRoleHost, nil
	}
	if ok := m.roomPlayers.AddParticipant(meta, login); ok {
		return types.PlayerRoleParticipant, nil
	}
	return types.PlayerRoleParticipant, fmt.Errorf("failed to join the room %s", roomName)
}

type rooms keyValStore

func newRooms() *rooms {
	r := (rooms)(*newKeyValStore())
	return &r
}

type roomMeta struct {
	Name            string
	Password        string
	MaxPlayersCount int
}

func (r *rooms) PutRoomMeta(meta *roomMeta) bool {
	kvs := (*keyValStore)(r)
	roomName := meta.Name
	return kvs.Put(roomName, meta)
}

func (r *rooms) GetRoomMeta(roomName string) (*roomMeta, bool) {
	kvs := (*keyValStore)(r)
	roomMetaIface, ok := kvs.Get(roomName)
	if !ok {
		return nil, false
	}
	roomMeta := (roomMetaIface).(*roomMeta)
	return roomMeta, true
}

type roomPlayers keyValStore

func newRoomPlayers() *roomPlayers {
	rp := (roomPlayers)(*newKeyValStore())
	return &rp
}

type players struct {
	Host      string
	Players   []string
	Observers []string
}

func newPlayers() *players {
	p := &players{
		Players:   make([]string, 0, 3),
		Observers: make([]string, 0),
	}
	return p
}

func (rp *roomPlayers) PutHost(roomName string, hostToken string) bool {
	kvs := (*keyValStore)(rp)
	ok := kvs.Alter(roomName, func(playersIface interface{}, exist bool) (interface{}, bool) {
		if !exist {
			p := newPlayers()
			p.Host = hostToken
			return p, true
		}
		p := playersIface.(*players)
		if len(p.Host) != 0 {
			return nil, false
		}
		p.Host = hostToken
		return p, true
	})
	return ok
}

func (rp *roomPlayers) AddParticipant(meta *roomMeta, login string) bool {
	kvs := (*keyValStore)(rp)
	ok := kvs.Alter(meta.Name, func(playersIface interface{}, exist bool) (interface{}, bool) {
		if !exist {
			return nil, false
		}
		p := playersIface.(*players)
		if meta.MaxPlayersCount == len(p.Players) {
			return nil, false
		}
		p.Players = append(p.Players, login)
		return p, true
	})
	return ok
}
