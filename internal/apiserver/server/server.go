package server

import (
	"context"
	"log"

	"github.com/Vadimkatr/grpcchatwithrooms/internal/apiserver/rooms"
	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
)

type Server struct {
	Rooms *rooms.Rooms
}

func InitServer() (*Server, error) {
	return &Server{
		Rooms: rooms.InitRooms(),
	}, nil
}

func (s *Server) CreateNewRoom(ctx context.Context, pconn *pb.CreateRoom) (*pb.Room, error) {
	_, err := s.Rooms.FindRoomByName(pconn.RoomName)
	if err != rooms.ErrRoomNotFound {
		log.Printf("error while creating room: %v", rooms.ErrRoomIsExist)
		return &pb.Room{}, rooms.ErrRoomIsExist
	}

	rm, err := s.Rooms.CreateRoom(pconn.RoomName)
	if err != nil {
		log.Printf("error while creating room: %v", err)
		return &pb.Room{}, err
	}

	log.Printf("Create new room: %v", rm)

	return &pb.Room{
		Id:   rm.Id,
		Name: rm.Name,
	}, nil
}

func (s *Server) CreateStream(pconn *pb.Connect, stream pb.ChatRooms_CreateStreamServer) error {
	err := s.Rooms.CreateConnection(pconn, stream)
	if err != nil {
		log.Printf("Error while creating stream: %v\n", err)
		return err
	}

	return nil
}

func (s *Server) BroadcastRoomMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	return s.Rooms.BroadcastRoomMessage(ctx, msg)
}

func (s *Server) GetAllRooms(ctx context.Context, epmty *pb.Empty) (*pb.ListRoom, error) {
	rms := s.Rooms.GetAllRooms()
	listRooms := &pb.ListRoom{}
	for _, rm := range rms {
		listRooms.Rooms = append(
			listRooms.Rooms,
			&pb.Room{
				Id:   rm.Id,
				Name: rm.Name,
			},
		)
	}

	return listRooms, nil
}
