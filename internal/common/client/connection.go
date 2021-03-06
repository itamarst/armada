package client

import (
	"strings"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/G-Research/armada/internal/common"
	"github.com/G-Research/armada/internal/common/client/auth/kerberos"
	"github.com/G-Research/armada/internal/common/client/auth/oidc"
)

type ApiConnectionDetails struct {
	ArmadaUrl                   string
	BasicAuth                   common.LoginCredentials
	OpenIdAuth                  oidc.PKCEDetails
	OpenIdPasswordAuth          oidc.ClientPasswordDetails
	OpenIdClientCredentialsAuth oidc.ClientCredentialsDetails
	KerberosAuth                kerberos.ClientConfig
}

func CreateApiConnection(config *ApiConnectionDetails, additionalDialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(300 * time.Millisecond)),
		grpc_retry.WithMax(3),
	}

	defaultCallOptions := grpc.WithDefaultCallOptions(grpc.WaitForReady(true))
	unuaryInterceptors := grpc.WithChainUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...))
	streamInterceptors := grpc.WithChainStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...))

	dialOpts := append(additionalDialOptions,
		defaultCallOptions,
		unuaryInterceptors,
		streamInterceptors,
		transportCredentials(config.ArmadaUrl))

	creds, err := perRpcCredentials(config)
	if err != nil {
		return nil, err
	}
	if creds != nil {
		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(creds))
	}

	return grpc.Dial(config.ArmadaUrl, dialOpts...)
}

func perRpcCredentials(config *ApiConnectionDetails) (credentials.PerRPCCredentials, error) {
	if config.BasicAuth.Username != "" {
		return &config.BasicAuth, nil

	} else if config.OpenIdAuth.ProviderUrl != "" {
		return oidc.AuthenticatePkce(config.OpenIdAuth)

	} else if config.OpenIdPasswordAuth.ProviderUrl != "" {
		return oidc.AuthenticateWithPassword(config.OpenIdPasswordAuth)

	} else if config.OpenIdClientCredentialsAuth.ProviderUrl != "" {
		return oidc.AuthenticateWithClientCredentials(config.OpenIdClientCredentialsAuth)

	} else if config.KerberosAuth.Enabled {
		return kerberos.NewSPNEGOCredentials(config.ArmadaUrl, config.KerberosAuth)
	}
	return nil, nil
}

func transportCredentials(url string) grpc.DialOption {
	if !strings.Contains(url, "localhost") {
		return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}
	return grpc.WithInsecure()
}
