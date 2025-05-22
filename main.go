package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"slices"

	"encoding/json"
	pb "video3/proto"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type Todo struct {
	Id        int32  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var todos []Todo

type server struct {
	pb.UnimplementedTodosServer
}

func (s *server) GetMany(_ context.Context, in *pb.GetManyRequest) (*pb.ManyTodosResponse, error) {
	log.Printf("Received: %v", in)
	var todosResponse []*pb.Todo
	for _, t := range todos {
		todosResponse = append(todosResponse, &pb.Todo{
			Id:        t.Id,
			Title:     t.Title,
			Completed: t.Completed,
		})
	}
	return &pb.ManyTodosResponse{
		Todos: todosResponse,
	}, nil
}

func (s *server) GetOne(_ context.Context, in *pb.GetOneRequest) (*pb.Todo, error) {
	log.Printf("Received: %v", in)
	for _, t := range todos {
		if t.Id == in.Id {
			return &pb.Todo{
				Id:        t.Id,
				Title:     t.Title,
				Completed: t.Completed,
			}, nil
		}
	}
	return nil, fmt.Errorf("todo with id %d not found", in.Id)
}

func (s *server) CreateOne(_ context.Context, in *pb.CreateOneRequest) (*pb.Todo, error) {
	log.Printf("Received: %v", in)
	var newID int32 = 1
	if len(todos) > 0 {
		newID = todos[len(todos)-1].Id + 1
	}
	todo := Todo{
		Id:        newID,
		Title:     in.Title,
		Completed: false,
	}
	todos = append(todos, todo)
	return &pb.Todo{
		Id:        todo.Id,
		Title:     todo.Title,
		Completed: todo.Completed,
	}, nil
}

func (s *server) UpdateOne(_ context.Context, in *pb.UpdateOneRequest) (*pb.Todo, error) {
	log.Printf("Received: %v", in)
	for i, t := range todos {
		if t.Id == in.Id {
			todos[i].Title = in.Title
			todos[i].Completed = in.Completed
			return &pb.Todo{
				Id:        todos[i].Id,
				Title:     todos[i].Title,
				Completed: todos[i].Completed,
			}, nil
		}
	}
	return nil, fmt.Errorf("todo with id %d not found", in.Id)
}

func (s *server) DeleteOne(_ context.Context, in *pb.GetOneRequest) (*pb.DeleteOneResponse, error) {
	log.Printf("Received: %v", in)
	for i, t := range todos {
		if t.Id == in.Id {
			todos = slices.Delete(todos, i, i+1)
			return &pb.DeleteOneResponse{
				Message: fmt.Sprintf("todo with id %d deleted", in.Id),
			}, nil
		}
	}
	return nil, fmt.Errorf("todo with id %d not found", in.Id)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, using host environment variables")
	}

	file, err := os.Open(os.Getenv("DATA_JSON_URI"))
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&todos); err != nil {
		log.Fatalf("failed to decode json: %v", err)
	}

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterTodosServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
