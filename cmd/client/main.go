package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
			if err == io.EOF { // when user type Ctrl+D to get EOF
				break
			} else {
				log.Fatal(err) // something went wrong
				return
			}
		}
		text = text[:len(text)-1]
		switch text {
		case "1":
			{
				rmId, err := createRoom(user)
				if err != nil {
					log.Printf("creating room failed: %v\n", err)
				} else {
					log.Printf("new room with id %s created!\n", rmId)
				}
			}
		case "2":
			getListRooms()
		case "3":
			{
				done := make(chan int)

				fmt.Println("Enter room name to connect:")
				reader := bufio.NewReader(os.Stdin)
				roomName, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF { // when user type Ctrl+D to get EOF
						break
					} else {
						log.Fatal(err) // something went wrong
						break // TODO: fix
					}
				}
				roomName = roomName[:len(roomName)-1]

				err = connectToRoom(user, roomName)
				if err != nil {
					fmt.Errorf("error while connecting to room: %v", err)
					continue
				}

				wait.Add(1)
				go func() {
					defer wait.Done()

					fmt.Printf("You connected to room %s. Type something:\n", roomName)
					scanner := bufio.NewScanner(os.Stdin)
					for scanner.Scan() {
						msg := &pb.Message{
							UserName:  user.Name,
							UserId:    user.Id,
							RoomName:  roomName,
							RoomId:    "",
							Content:   scanner.Text(),
							Timestamp: timestamp.String(),
						}

						if msg.GetContent() == "/exit" || msg.GetContent() == "/menu" {
							_, err := client.CloseStream(context.Background(),  &pb.Connect{
								User:     user,
								RoomName: roomName,
								Active:   true,
							})
							if err != nil {
								fmt.Println("Do something:", err)
							}
							fmt.Printf("You left room.\n")
							break
						}

						_, err := client.BroadcastRoomMessage(context.Background(), msg)
						if err != nil {
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
			deleteRoom(user)
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

func createRoom(user *pb.User) (string, error) {
	fmt.Println("Enter new room name:")
	reader := bufio.NewReader(os.Stdin)
	roomName, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF { // when user type Ctrl+D to get EOF
			return "", err
		} else {
			log.Fatal(err) // something went wrong
			return "", err
		}
	}
	roomName = roomName[:len(roomName)-1]

	rm, err := client.CreateNewRoom(context.Background(), &pb.CreateOrDelRoom{
		User:     user,
		RoomName: roomName,

	})
	if err != nil {
		return "", err
	}

	return rm.Id, nil
}

func getListRooms() {
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
			fmt.Printf("%d: %s\n", i+1, rm.Name)
		}
	}
}

func connectToRoom(user *pb.User, roomName string) error {
	stream, err := client.CreateStream(context.Background(), &pb.Connect{
		User:     user,
		RoomName: roomName,
		Active:   true,
	})
	if err != nil {
		return err
	}

	wait.Add(1)
	go func(str pb.ChatRooms_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading message: %v\n", err)
					break
				}
			}

			fmt.Printf("%s: %s\n", msg.GetUserName(), msg.GetContent())
		}
	}(stream)

	return nil
}

func deleteRoom(user *pb.User) error {
	fmt.Println("Enter room name to delete:")
	reader := bufio.NewReader(os.Stdin)
	roomName, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF { // when user type Ctrl+D to get EOF
			return err
		} else {
			log.Fatal(err) // something went wrong
			return err
		}
	}
	roomName = roomName[:len(roomName)-1]

	_, err = client.DeleteRoom(context.Background(), &pb.CreateOrDelRoom{
		User:     user,
		RoomName: roomName,

	})
	if err != nil {
		return err
	}

	return nil
}
