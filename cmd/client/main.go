package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"google.golang.org/grpc"
	"sync"
	"time"

	"fmt"
	pb "github.com/Vadimkatr/grpcchatwithrooms/internal/proto"
	"io"
	"log"
	"os"
)

var client pb.ChatRoomsClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func main() {
	timestamp := time.Now()

	name := flag.String("N", fmt.Sprintf("NoName%d", time.Now().Unix()), "The name of the user")
	flag.Parse()

	id := sha256.Sum256([]byte(timestamp.String() + *name))

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	client = pb.NewChatRoomsClient(conn)
	user := &pb.User{
		Id:   hex.EncodeToString(id[:]),
		Name: *name,
	}

	for {
		menu := fmt.Sprintf("\nMenu:\n" +
			"1) Create room;\n" +
			"2) List room;\n" +
			"3) Connect to existing rooom;\n" +
			"4) Delete room;\n" +
			"5) Exit\n")
		fmt.Printf("%s", menu)
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error : %v.\n", err) // something went wrong
			return
		}

		text = text[:len(text)-1]
		switch text {
		case "1":
			{
				fmt.Println("Enter new room name:")
				reader := bufio.NewReader(os.Stdin)
				roomName, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Creating room failed: %v.\n", err) // something went wrong
					break
				}
				roomName = roomName[:len(roomName)-1]

				rmId, err := createRoom(user, roomName)
				if err != nil {
					fmt.Printf("Creating room failed: %v.\n", err)
				} else {
					fmt.Printf("New room with id %s created.\n", rmId)
				}
			}
		case "2":
			printListRooms()
		case "3":
			{
				done := make(chan int)
				fmt.Println("Enter room name to connect:")
				reader := bufio.NewReader(os.Stdin)
				roomName, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Connecting room failed: %v.\n", err) // something went wrong
					break
				}

				roomName = roomName[:len(roomName)-1]
				err = connectToRoom(user, roomName)
				if err != nil {
					fmt.Printf("Error while connecting to room: %v.\n", err)
					continue
				}

				wait.Add(1)
				go func() {
					defer wait.Done()

					fmt.Printf("You connected to room \"%s\". Type something:\n", roomName)
					reader := bufio.NewReader(os.Stdin)
					for {
						text, err := reader.ReadString('\n') // read msg from stdin
						if err != nil {
							if err == io.EOF { // when user type Ctrl+D to get EOF
								if err := disconnectFromRoom(user, roomName); err != nil {
									fmt.Sprintf("Error while disconnecting from room: \"%v\".\n", err)
								}
								fmt.Printf("You left room.\n")
								break
							} else {
								log.Fatal(err) // something went wrong
								return
							}
						}

						text = text[:len(text)-1]
						msg := &pb.Message{
							UserName:  user.Name,
							UserId:    user.Id,
							RoomName:  roomName,
							RoomId:    "",
							Content:   text,
							Timestamp: timestamp.String(),
						}

						if msg.GetContent() == "/exit" || msg.GetContent() == "/menu" {
							if err := disconnectFromRoom(user, roomName); err != nil {
								fmt.Sprintf("Error while disconnecting from room: %v.\n", err)
							}
							fmt.Printf("You left room.\n")
							break
						}

						if _, err := client.BroadcastRoomMessage(context.Background(), msg); err != nil {
							fmt.Printf("Error Sending Message: %v", err)
							break
						}

					}
				}()

				go func() {
					wait.Wait()
					close(done)
				}()

				<-done
			}
		case "4":
			{
				fmt.Println("Enter room name to delete:")
				reader := bufio.NewReader(os.Stdin)
				roomName, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Deleting room failed: %v.\n", err) // something went wrong
					break
				}

				roomName = roomName[:len(roomName)-1]
				err = deleteRoom(user, roomName)
				if err != nil {
					fmt.Printf("Error while deleting room: %v.\n", err)
				} else {
					fmt.Printf("Room deleted.\n")
				}
			}
		case "5":
			{
				fmt.Printf("Bye bye...")
				return
			}
		default:
			fmt.Printf("Oh no, try again...")
		}
	}
}

func createRoom(user *pb.User, roomName string) (string, error) {
	rm, err := client.CreateNewRoom(context.Background(), &pb.CreateOrDelRoom{
		User:     user,
		RoomName: roomName,
	})
	if err != nil {
		return "", err
	}

	return rm.Id, nil
}

func printListRooms() {
	rms, err := client.GetAllRooms(context.Background(), &pb.Empty{})
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(rms.Rooms) == 0 {
		fmt.Printf("There are no rooms\n")
	} else {
		fmt.Printf("Rooms:\n")
		for i, rm := range rms.Rooms {
			fmt.Printf("%d: \"%s\"\n", i+1, rm.Name)
		}
	}
}

// connectToRoom - send request to create stream and run goroutines that print incoming message from connected room
func connectToRoom(user *pb.User, roomName string) error {
	stream, err := client.CreateStream(context.Background(), &pb.Connect{
		User:     user,
		RoomName: roomName,
		Active:   true,
	})
	if err != nil {
		return err
	}

	md, err := stream.Header()
	if err, ok := md["error"]; ok {
		return errors.New(err[0])
	}

	wait.Add(1)
	go func(str pb.ChatRooms_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				fmt.Printf("Error reading message: %v\n", err)
				break
			}

			fmt.Printf("%s: %s\n", msg.GetUserName(), msg.GetContent())
		}
	}(stream)

	return nil
}

// disconnectFromRoom - send request to close user stream
func disconnectFromRoom(user *pb.User, roomName string) error {
	_, err := client.CloseStream(context.Background(), &pb.Connect{
		User:     user,
		RoomName: roomName,
		Active:   false,
	})

	return err
}

// deleteRoom - send request to delete room
func deleteRoom(user *pb.User, roomName string) error {
	_, err := client.DeleteRoom(context.Background(), &pb.CreateOrDelRoom{
		User:     user,
		RoomName: roomName,
	})
	if err != nil {
		return err
	}

	return nil
}
