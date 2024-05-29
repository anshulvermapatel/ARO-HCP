package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/cobra"

	"github.com/Azure/ARO-HCP/frontend/pkg/config"
	"github.com/Azure/ARO-HCP/frontend/pkg/database"
	"github.com/Azure/ARO-HCP/frontend/pkg/frontend"
)

type FrontendOpts struct {
	clustersServiceURL string
	insecure           bool

	region string
	port   int

	databaseName string
	databaseURL  string
}

func NewRootCmd() *cobra.Command {
	opts := &FrontendOpts{}
	rootCmd := &cobra.Command{
		Use:   "aro-hcp-frontend",
		Args:  cobra.NoArgs,
		Short: "Serve the ARO HCP Frontend",
		Long: `Serve the ARO HCP Frontend

	This command runs the ARO HCP Frontend. It communicates with Clusters Service and a CosmosDB

	# Run ARO HCP Frontend locally to connect to a local Clusters Service at http://localhost:8000
	./aro-hcp-frontend --database-name ${DB_NAME} --database-url ${DB_URL} --region ${REGION} \
		--clusters-service-url "http://localhost:8000"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.Run()
		},
	}

	rootCmd.Flags().StringVar(&opts.databaseName, "database-name", os.Getenv("DB_NAME"), "database name")
	rootCmd.Flags().StringVar(&opts.databaseURL, "database-url", os.Getenv("DB_URL"), "database url")
	rootCmd.Flags().StringVar(&opts.region, "region", os.Getenv("REGION"), "Azure region")
	rootCmd.Flags().IntVar(&opts.port, "port", 8443, "port to listen on")

	rootCmd.Flags().StringVar(&opts.clustersServiceURL, "clusters-service-url", "https://api.openshift.com", "URL of the OCM API gateway.")
	rootCmd.Flags().BoolVar(&opts.insecure, "insecure", false, "Skip validating TLS for clusters-service.")

	return rootCmd
}

func (opts *FrontendOpts) Run() error {
	version := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				version = setting.Value
				break
			}
		}
	}

	logger := config.DefaultLogger()
	logger.Info(fmt.Sprintf("%s (%s) started", frontend.ProgramName, version))

	// Init prometheus emitter
	prometheusEmitter := frontend.NewPrometheusEmitter()

	// Configure database configuration and client
	dbConfig := database.NewDatabaseConfig(opts.databaseName, opts.databaseURL)
	dbClient, err := database.NewDatabaseClient(dbConfig)
	if err != nil {
		return fmt.Errorf("creating the database client failed: %v", err)
	}

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", opts.port))
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Application running in region: %s", opts.region))

	// Initialize Clusters Service Client
	conn, err := sdk.NewUnauthenticatedConnectionBuilder().
		URL(opts.clustersServiceURL).
		Insecure(opts.insecure).
		Build()
	if err != nil {
		return err
	}

	f := frontend.NewFrontend(logger, listener, prometheusEmitter, dbClient, opts.region, conn)

	// Verify the Async DB is available and accessible
	logger.Info("Testing DB Access")
	result, err := dbClient.DBConnectionTest(context.Background())
	if err != nil {
		logger.Error(fmt.Sprintf("Database test failed to fetch properties: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Database check completed - %s", result))
	}

	stop := make(chan struct{})
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go f.Run(context.Background(), stop)

	sig := <-signalChannel
	logger.Info(fmt.Sprintf("caught %s signal", sig))
	close(stop)

	f.Join()
	logger.Info(fmt.Sprintf("%s (%s) stopped", frontend.ProgramName, version))

	return nil
}