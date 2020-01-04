package room

import (
	"context"
	"log"

	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
	"sync"
)

type User struct {
	Id   string
	Name string
}

type Connection struct {
	stream pb.ChatRooms_CreateStreamServer
	user   *User
	active bool
	error  chan error
}

type Room struct {
	Connections []*Connection
	Id          string
	Name        string
	CreatorId   string
}

func (rm *Room) BroadcastMessageToRoom(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range rm.Connections {
		wait.Add(1)

		go func(msg *pb.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				err := conn.stream.Send(msg)
				log.Println("Sending message to: ", conn.stream)

				if err != nil {
					log.Printf("Error with Stream: %v - Error: %v\n", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &pb.Close{}, nil
}
