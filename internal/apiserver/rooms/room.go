package rooms

import (
	"log"
	"sync"

	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
)

type User struct {
	Id   string
	Name string
}

type Connection struct {
	stream pb.ChatRooms_CreateStreamServer
	active bool
	user   *User
	Error  chan error
}

type Room struct {
	Connections []*Connection
	Id          string
	Name        string
	CreatorId   string
}

func (rm *Room) BroadcastMessageToRoom(msg *pb.Message) error {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range rm.Connections {
		wait.Add(1)

		go func(msg *pb.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				err := conn.stream.Send(msg)
				log.Printf("Room %s: user %s sending message to stream: %v.\n",
					msg.RoomName, msg.UserName, conn.stream)

				if err != nil {
					log.Printf("Error with Stream: %v - Error: %v\n", conn.stream, err)
					conn.active = false
					conn.Error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return nil
}
