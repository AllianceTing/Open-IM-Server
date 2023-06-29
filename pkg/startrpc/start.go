package startrpc

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/config"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/log"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/mw"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/network"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/prome"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/discoveryregistry"
	openKeeper "github.com/OpenIMSDK/Open-IM-Server/pkg/discoveryregistry/zookeeper"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/utils"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Start(rpcPort int, rpcRegisterName string, prometheusPort int, rpcFn func(client discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error, options ...grpc.ServerOption) error {
	fmt.Println("start", rpcRegisterName, "server, port: ", rpcPort, "prometheusPort:", prometheusPort, ", OpenIM version: ", config.Version)
	listener, err := net.Listen("tcp", net.JoinHostPort(config.Config.ListenIP, strconv.Itoa(rpcPort)))
	if err != nil {
		return err
	}
	defer listener.Close()
	client, err := openKeeper.NewClient(config.Config.Zookeeper.ZkAddr, config.Config.Zookeeper.Schema,
		openKeeper.WithFreq(time.Hour), openKeeper.WithUserNameAndPassword(config.Config.Zookeeper.UserName,
			config.Config.Zookeeper.Password), openKeeper.WithRoundRobin(), openKeeper.WithTimeout(10), openKeeper.WithLogger(log.NewZkLogger()))
	if err != nil {
		return utils.Wrap1(err)
	}
	if client.CreateRpcRootNodes(config.GetServiceNames()); err != nil {
		return err
	}
	defer client.CloseZK()
	client.AddOption(mw.GrpcClient(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	registerIP, err := network.GetRpcRegisterIP(config.Config.RpcRegisterIP)
	if err != nil {
		return err
	}
	// ctx 中间件
	if config.Config.Prometheus.Enable {
		prome.NewGrpcRequestCounter()
		prome.NewGrpcRequestFailedCounter()
		prome.NewGrpcRequestSuccessCounter()
		unaryInterceptor := mw.InterceptChain(grpcPrometheus.UnaryServerInterceptor, mw.RpcServerInterceptor)
		options = append(options, []grpc.ServerOption{
			grpc.StreamInterceptor(grpcPrometheus.StreamServerInterceptor),
			grpc.UnaryInterceptor(unaryInterceptor),
		}...)
	} else {
		options = append(options, mw.GrpcServer())
	}
	srv := grpc.NewServer(options...)
	defer srv.GracefulStop()
	err = rpcFn(client, srv)
	if err != nil {
		return utils.Wrap1(err)
	}
	err = client.Register(rpcRegisterName, registerIP, rpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return utils.Wrap1(err)
	}
	go func() {
		if config.Config.Prometheus.Enable && prometheusPort != 0 {
			if err := prome.StartPrometheusSrv(prometheusPort); err != nil {
				panic(err.Error())
			}
		}
	}()
	return utils.Wrap1(srv.Serve(listener))
}
