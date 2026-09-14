package plugin

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

// notServingHealthServer is a minimal health server implementation
// that returns NOT_SERVING for all health checks.
type notServingHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (h *notServingHealthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	}, nil
}

// TestGRPC_Ping_ChecksServingStatus verifies that Ping() returns an error
// when the gRPC health check returns a non-SERVING status, not just on transport errors.
// This test will fail if Ping() only checks for transport errors and ignores the response status.
func TestGRPC_Ping_ChecksServingStatus(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, &notServingHealthServer{})

	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer conn.Close()

	client := &GRPCClient{Conn: conn}
	
	// Ping should fail because the health check returns NOT_SERVING
	err = client.Ping()
	if err == nil {
		t.Fatal("expected Ping() to fail when health check returns NOT_SERVING")
	}
}
