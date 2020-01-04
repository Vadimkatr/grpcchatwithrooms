package room

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"
	"time"

	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
)

type Rooms struct {
	rooms []*Room
}

func InitRooms() *Rooms {
	return &Rooms{}
}

func (rms *Rooms) CreateConnection(pconn *pb.Connect, stream pb.ChatRooms_CreateStreamServer) error {
	user := User{
		Id:   pconn.User.Id,
		Name: pconn.User.Name,
	}
	conn := &Connection{
		stream: stream,
		user:   &user,
		active: true,
		error:  make(chan error),
	}

	rm, err := findRoomByName(rms.rooms, pconn.RoomName)
	if err != nil {
		return err
	}

	rm.Connections = append(rm.Connections, conn)
	log.Printf("User %s connect to room %s\n", pconn.User.Name, pconn.RoomName)

	return <-conn.error
}

func findRoomByName(rooms []*Room, roomName string) (*Room, error) {
	for _, rm := range rooms {
		if rm.Name == roomName {
			return rm, nil
		}
	}

	return &Room{}, ErrRoomNotFound
}

func (rms *Rooms) FindRoomByName(name string) (*Room, error) {
	for _, rm := range rms.rooms {
		if rm.Name == name {
			return rm, nil
		}
	}

	return nil, ErrRoomNotFound
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
	for i, rm := range rms.rooms {
		if rm.Name == roomName {
			if rm.CreatorId == creatorId {
				rms.rooms = append(rms.rooms[:i], rms.rooms[i+1:]...)
				log.Printf("delete room: %s", roomName)
				return nil
			}
			err := ErrDelRoomPermissionDen
			log.Printf("error while deleting room: %v", err)
			return err
		}
	}

	err := ErrRoomNotFound
	log.Printf("error while deleting room: %v", err)
	return err
}

func (rms *Rooms) BroadcastRoomMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	wait := sync.Mutex{}
	rm, err := rms.FindRoomByName(msg.RoomName)
	if err != nil {
		return &pb.Close{}, err
	}
	wait.Lock()
	defer wait.Unlock()
	return rm.BroadcastMessageToRoom(ctx, msg)
}

func (rms *Rooms) GetAllRooms() []*Room {
	return rms.rooms
}
