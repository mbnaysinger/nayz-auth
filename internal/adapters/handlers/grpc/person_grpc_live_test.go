package grpc

import (
	"context"
	"os"
	"testing"
	"time"

	pb "github.com/mbnaysinger/nayz-auth/pkg/grpc/pb"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Teste de integração contra o servidor gRPC local em execução.
// Roda apenas quando NAYZ_GRPC_LIVE_TEST=1 (exige serviço + banco de pé).
func TestGetPersonsByIdsLive(t *testing.T) {
	if os.Getenv("NAYZ_GRPC_LIVE_TEST") != "1" {
		t.Skip("teste live desabilitado; defina NAYZ_GRPC_LIVE_TEST=1")
	}

	conn, err := googlegrpc.NewClient("localhost:50051", googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("falha ao conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewPersonServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ids := os.Getenv("NAYZ_GRPC_TEST_IDS")
	if ids == "" {
		t.Fatal("defina NAYZ_GRPC_TEST_IDS com um UUID existente")
	}

	resp, err := client.GetPersonsByIds(ctx, &pb.GetPersonsByIdsRequest{
		PersonIds: []string{ids, "00000000-0000-0000-0000-000000000000"},
	})
	if err != nil {
		t.Fatalf("erro na chamada batch: %v", err)
	}
	if len(resp.Persons) != 1 {
		t.Fatalf("esperada 1 pessoa (ID inexistente omitido), vieram %d", len(resp.Persons))
	}
	if resp.Persons[0].Name == "" || resp.Persons[0].Id != ids {
		t.Errorf("pessoa retornada inconsistente: %+v", resp.Persons[0])
	}
}
