package server

import (
	"context"
	"google.golang.org/grpc/metadata"
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

func (s *Server) CreateNewRoom(ctx context.Context, pconn *pb.CreateOrDelRoom) (*pb.Room, error) {
	_, _, err := s.Rooms.GetRoomByName(pconn.RoomName)
	if err == nil {
		log.Printf("Error while creating room: %v.\n", rooms.ErrRoomIsExist)
		return &pb.Room{}, rooms.ErrRoomIsExist
	}

	rm, err := s.Rooms.CreateRoom(pconn.RoomName, pconn.User.Id)
	if err != nil {
		log.Printf("Error while creating room: %v.\n", err)
		return &pb.Room{}, err
	}

	log.Printf("Create new room: %v.\n", rm)
	return &pb.Room{
		Id:        rm.Id,
		Name:      rm.Name,
		CreatorId: pconn.User.Id,
	}, nil
}

func (s *Server) CreateStream(pconn *pb.Connect, stream pb.ChatRooms_CreateStreamServer) error {
	conn, err := s.Rooms.CreateStreamConnection(pconn, stream)
	md := metadata.MD{}
	if err != nil {
		log.Printf("Error while creating stream: user %s can't connect to room %s because err: %v.\n",
			pconn.User.Name, pconn.RoomName, err)
		md.Set("error", err.Error())
		errh := stream.SendHeader(md)
		if errh != nil {
			log.Printf("Error sending header: \"%v\"; user %s try's connect to room %s and get error: \"%v\".\n",
				errh, pconn.User.Name, pconn.RoomName, err)
		}
		return err
	}

	errh := stream.SendHeader(md)
	if errh != nil {
		log.Printf("Error sending header: \"%v\"; user %s try's connect to room %s.\n",
			errh, pconn.User.Name, pconn.RoomName)
		return errh
	}
	log.Printf("User %s connected to rooms %s.\n", pconn.User.Name, pconn.RoomName)
	return <-conn.Error
}

func (s *Server) CloseStream(ctx context.Context, pconn *pb.Connect) (*pb.Empty, error) {
	err := s.Rooms.CloseStreamConnection(pconn)
	if err != nil {
		log.Printf("Error while closing connection: %v.\n", err)
		return &pb.Empty{}, err
	}

	log.Printf("User %s closed own connection.\n", pconn.User.Name)
	return &pb.Empty{}, nil
}

func (s *Server) DeleteRoom(ctx context.Context, pconn *pb.CreateOrDelRoom) (*pb.Empty, error) {
	err := s.Rooms.DeleteRoom(pconn.RoomName, pconn.User.Id)
	if err != nil {
		log.Printf("Error while deleting room: %v.\n", err)
		return &pb.Empty{}, err
	}

	log.Printf("User %s deleted room %s.\n", pconn.User.Name, pconn.RoomName)
	return &pb.Empty{}, nil
}

func (s *Server) BroadcastRoomMessage(ctx context.Context, msg *pb.Message) (*pb.Empty, error) {
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
