package grpc

import (
	"context"

	"github.com/mbnaysinger/nayz-auth/internal/core/services"
	pb "github.com/mbnaysinger/nayz-auth/pkg/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PersonGrpcHandler struct {
	pb.UnimplementedPersonServiceServer
	personService *services.PersonService
}

func NewPersonGrpcHandler(personService *services.PersonService) *PersonGrpcHandler {
	return &PersonGrpcHandler{personService: personService}
}

func (h *PersonGrpcHandler) GetPerson(ctx context.Context, req *pb.GetPersonRequest) (*pb.GetPersonResponse, error) {
	if req.PersonId == "" {
		return nil, status.Error(codes.InvalidArgument, "person_id is required")
	}

	person, err := h.personService.GetPerson(ctx, req.PersonId)
	if err != nil {
		if err == services.ErrPersonNotFound {
			return nil, status.Error(codes.NotFound, "person not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var userId string
	if person.UserID != nil {
		userId = *person.UserID
	}

	return &pb.GetPersonResponse{
		Id:         person.ID,
		UserId:     userId,
		Name:       person.Name,
		Identifier: person.Identifier,
		IsActive:   person.IsActive,
	}, nil
}

// GetPersonsByIds resolve um lote de pessoas em uma única chamada.
// IDs inexistentes são simplesmente omitidos da resposta (cabe ao cliente tratar ausências).
func (h *PersonGrpcHandler) GetPersonsByIds(ctx context.Context, req *pb.GetPersonsByIdsRequest) (*pb.GetPersonsByIdsResponse, error) {
	persons, err := h.personService.GetPersonsByIDs(ctx, req.PersonIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.GetPersonsByIdsResponse{Persons: make([]*pb.Person, 0, len(persons))}
	for _, person := range persons {
		var userId string
		if person.UserID != nil {
			userId = *person.UserID
		}
		resp.Persons = append(resp.Persons, &pb.Person{
			Id:         person.ID,
			UserId:     userId,
			Name:       person.Name,
			Identifier: person.Identifier,
			IsActive:   person.IsActive,
		})
	}
	return resp, nil
}
