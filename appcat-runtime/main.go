package main

import (
	"flag"
	"fmt"
	"os"

	function "github.com/crossplane/function-sdk-go"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	addr := flag.String("addr", ":9443", "gRPC listen address")
	tlsDirFlag := flag.String("tls-dir", "", "Directory containing tls.crt, tls.key, ca.crt (defaults to TLS_SERVER_CERTS_DIR)")
	proxyEndpoint := flag.String("proxy", "", "Proxy endpoint for debugging (e.g., '127.0.0.1:9443'). If set, all requests are forwarded to this endpoint.")
	insecure := flag.Bool("insecure", false, "Run in insecure mode without TLS (for local debugging only)")
	flag.Parse()

	// Get TLS directory from flag or environment
	tlsDir := *tlsDirFlag
	if tlsDir == "" {
		tlsDir = os.Getenv("TLS_SERVER_CERTS_DIR")
	}

	// Validate TLS configuration unless in insecure mode
	if !*insecure && tlsDir == "" {
		panic("TLS server cert directory not set; set --tls-dir or TLS_SERVER_CERTS_DIR, or use --insecure for local debugging")
	}

	// Health server allows Crossplane to probe readiness
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("function-appcat-poc", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Create and register manager with proxy endpoint
	mgr := NewManager(zap.New(), *proxyEndpoint)

	// Build server options
	opts := []function.ServeOption{
		function.Listen("tcp", *addr),
		function.Insecure(*insecure),
		function.WithHealthServer(healthSrv),
	}
	if !*insecure {
		opts = append(opts, function.MTLSCertificates(tlsDir))
	}

	// Log startup configuration
	if *insecure {
		fmt.Printf("Starting gRPC server on %s (INSECURE MODE)\n", *addr)
	} else {
		fmt.Printf("Starting gRPC server on %s (mTLS: %s)\n", *addr, tlsDir)
	}
	if *proxyEndpoint != "" {
		fmt.Printf("PROXY MODE: Forwarding to %s\n", *proxyEndpoint)
	}

	if err := function.Serve(mgr, opts...); err != nil {
		panic(fmt.Errorf("serve: %w", err))
	}
}
