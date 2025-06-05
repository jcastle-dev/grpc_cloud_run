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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	port = flag.Int("port", 5000, "The server port")
)

type Todo struct {
	Id        int32  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var todos []Todo

type Server struct {
	pb.UnimplementedTodosServer
}

func (server *Server) GetMany(_ context.Context, in *pb.GetManyRequest) (*pb.GetManyResponse, error) {
	log.Printf("Received: %v", in)
	var todosResponse []*pb.Todo
	for _, todo := range todos {
		todosResponse = append(todosResponse, &pb.Todo{
			Id:        todo.Id,
			Title:     todo.Title,
			Completed: todo.Completed,
		})
	}
	return &pb.GetManyResponse{
		Todos: todosResponse,
	}, nil
}

func (server *Server) GetOne(_ context.Context, in *pb.GetOneRequest) (*pb.Todo, error) {
	log.Printf("Received: %v", in)
	for _, todo := range todos {
		if todo.Id == in.Id {
			return &pb.Todo{
				Id:        todo.Id,
				Title:     todo.Title,
				Completed: todo.Completed,
			}, nil
		}
	}
	error := fmt.Errorf("todo with id %d not found", in.Id)
	return nil, status.Errorf(codes.NotFound, error.Error())
}

func (server *Server) CreateOne(_ context.Context, in *pb.CreateOneRequest) (*pb.Todo, error) {
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

func (server *Server) UpdateOne(_ context.Context, in *pb.UpdateOneRequest) (*pb.Todo, error) {
	log.Printf("Received: %v", in)
	for index, todo := range todos {
		if todo.Id == in.Id {
			todos[index].Title = in.Title
			todos[index].Completed = in.Completed
			return &pb.Todo{
				Id:        todos[index].Id,
				Title:     todos[index].Title,
				Completed: todos[index].Completed,
			}, nil
		}
	}
	error := fmt.Errorf("todo with id %d not found", in.Id)
	return nil, status.Errorf(codes.NotFound, error.Error())
}

func (server *Server) DeleteOne(_ context.Context, in *pb.GetOneRequest) (*pb.DeleteOneResponse, error) {
	log.Printf("Received: %v", in)
	for index, todo := range todos {
		if todo.Id == in.Id {
			todos = slices.Delete(todos, index, index+1)
			return &pb.DeleteOneResponse{
				Message: fmt.Sprintf("todo with id %d deleted", in.Id),
			}, nil
		}
	}
	error := fmt.Errorf("todo with id %d not found", in.Id)
	return nil, status.Errorf(codes.NotFound, error.Error())
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
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to start listener: %v", err)
	}
	server := grpc.NewServer()
	reflection.Register(server)
	pb.RegisterTodosServer(server, &Server{})
	log.Printf("server listening at %v", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatalf("server failure: %v", err)
	}
}
