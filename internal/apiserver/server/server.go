package server

import (
	"context"
	"log"

	"github.com/Vadimkatr/grpcchatwithrooms/internal/apiserver/room"
	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
)

type Server struct {
	Rooms *room.Rooms
}

func InitServer() (*Server, error) {
	return &Server{
		Rooms: room.InitRooms(),
	}, nil
}

func (s *Server) CreateNewRoom(ctx context.Context, pconn *pb.CreateOrDelRoom) (*pb.Room, error) {
	_, err := s.Rooms.FindRoomByName(pconn.RoomName)
	if err != room.ErrRoomNotFound {
		log.Printf("error while creating room: %v", room.ErrRoomIsExist)
		return &pb.Room{}, room.ErrRoomIsExist
	}

	rm, err := s.Rooms.CreateRoom(pconn.RoomName, pconn.User.Id)
	if err != nil {
		log.Printf("error while creating room: %v", err)
		return &pb.Room{}, err
	}

	log.Printf("Create new room: %v", rm)

	return &pb.Room{
		Id:        rm.Id,
		Name:      rm.Name,
		CreatorId: pconn.User.Id,
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

func (s *Server) DeleteRoom(ctx context.Context, pconn *pb.CreateOrDelRoom) (*pb.Empty, error) {
	return &pb.Empty{}, s.Rooms.DeleteRoom(pconn.RoomName, pconn.User.Id)
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
