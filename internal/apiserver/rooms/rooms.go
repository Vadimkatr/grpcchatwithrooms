package rooms

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
)

type Rooms struct {
	rooms []*Room
}

func InitRooms() *Rooms {
	return &Rooms{}
}

func (rms *Rooms) BroadcastRoomMessage(ctx context.Context, msg *pb.Message) (*pb.Empty, error) {
	rm, _, err := rms.GetRoomByName(msg.RoomName)
	if err != nil {
		return &pb.Empty{}, err
	}

	err = rm.BroadcastMessageToRoom(msg)
	if err != nil {
		return &pb.Empty{}, err
	}

	return &pb.Empty{}, nil
}

func (rms *Rooms) CreateStreamConnection(pconn *pb.Connect, stream pb.ChatRooms_CreateStreamServer) (*Connection, error) {
	user := User{
		Id:   pconn.User.Id,
		Name: pconn.User.Name,
	}
	conn := &Connection{
		stream: stream,
		active: true,
		user:   &user,
		Error:  make(chan error),
	}

	rm, _, err := rms.GetRoomByName(pconn.RoomName)
	if err != nil {
		return nil, err
	}

	rm.Connections = append(rm.Connections, conn)

	return conn, nil
}

func (rms *Rooms) CloseStreamConnection(pconn *pb.Connect) error {
	user := User{
		Id:   pconn.User.Id,
		Name: pconn.User.Name,
	}
	rm, _, err := rms.GetRoomByName(pconn.RoomName)
	if err != nil {
		return err
	}

	for i, conn := range rm.Connections {
		if conn.user.Id == user.Id {
			conn.Error <- nil                                                    // send nil to close connection
			close(conn.Error)                                                    // close channel
			rm.Connections = append(rm.Connections[:i], rm.Connections[i+1:]...) // del conn from rooms connections
			return nil
		}
	}

	return ErrConnNotFound
}

func (rms *Rooms) GetRoomByName(name string) (*Room, int, error) {
	for i, rm := range rms.rooms {
		if rm.Name == name {
			return rm, i, nil
		}
	}

	return nil, 0, ErrRoomNotFound
}

func (rms *Rooms) CreateRoom(name, creatorId string) (*Room, error) {
	timestamp := time.Now()
	id := sha256.Sum256([]byte(timestamp.String() + name))

	rm := &Room{
		Connections: nil,
		Id:          hex.EncodeToString(id[:]),
		Name:        name,
		CreatorId:   creatorId,
	}

	rms.rooms = append(rms.rooms, rm)
	return rm, nil
}

func (rms *Rooms) DeleteRoom(roomName, creatorId string) error {
	rm, index, err := rms.GetRoomByName(roomName)
	if err != nil {
		return err
	}

	// only user that create rooms can delete them
	if rm.CreatorId != creatorId {
		return ErrDelRoomPermissionDen
	}

	// close all connections in room
	for _, conn := range rm.Connections {
		conn.Error <- nil // send nil to close connection
		close(conn.Error) // close channel
	}

	// delete this room from all
	rms.rooms = append(rms.rooms[:index], rms.rooms[index+1:]...)
	return nil
}

func (rms *Rooms) GetAllRooms() []*Room {
	return rms.rooms
}
