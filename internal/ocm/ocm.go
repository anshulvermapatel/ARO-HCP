package ocm

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ClusterServiceClientSpec interface {
	GetConn() *sdk.Connection
	AddProperties(builder *cmv1.ClusterBuilder) *cmv1.ClusterBuilder
	GetCSCluster(ctx context.Context, internalID InternalID) (*cmv1.Cluster, error)
	PostCSCluster(ctx context.Context, cluster *cmv1.Cluster) (*cmv1.Cluster, error)
	UpdateCSCluster(ctx context.Context, internalID InternalID, cluster *cmv1.Cluster) (*cmv1.Cluster, error)
	DeleteCSCluster(ctx context.Context, internalID InternalID) error
	GetCSNodePool(ctx context.Context, internalID InternalID) (*cmv1.NodePool, error)
	PostCSNodePool(ctx context.Context, clusterInternalID InternalID, nodePool *cmv1.NodePool) (*cmv1.NodePool, error)
	UpdateCSNodePool(ctx context.Context, internalID InternalID, nodePool *cmv1.NodePool) (*cmv1.NodePool, error)
	DeleteCSNodePool(ctx context.Context, internalID InternalID) error
}

// Get the default set of properties for the Cluster Service
func getDefaultAdditionalProperities() map[string]string {
	// additionalProperties should be empty in production, it is configurable for development to pin to specific
	// provision shards or instruct CS to skip the full provisioning/deprovisioning flow.
	additionalProperties := map[string]string{
		// Enable the ARO HCP provisioner during development. For now, if not set a cluster will not progress past the
		// installing state in CS.
		"provisioner_hostedcluster_step_enabled": "true",
		// Enable the provisioning of ACM's ManagedCluster CR associated to the ARO-HCP
		// cluster during ARO-HCP Cluster provisioning. For now, if not set a cluster will not progress past the
		// installing state in CS.
		"provisioner_managedcluster_step_enabled": "true",

		// Enable the provisioning and deprovisioning of ARO-HCP Node Pools. For now, if not set the provisioning
		// and deprovisioning of day 2 ARO-HCP Node Pools will not be performed on the Management Cluster.
		"np_provisioner_provision_enabled":   "true",
		"np_provisioner_deprovision_enabled": "true",
	}
	return additionalProperties
}

type ClusterServiceClient struct {
	// Conn is an ocm-sdk-go connection to Cluster Service
	Conn *sdk.Connection

	// ProvisionShardID sets the provision_shard_id property for all cluster requests to Cluster Service, which pins all
	// cluster requests to Cluster Service to a specific shard during testing
	ProvisionShardID *string

	// ProvisionerNoOpProvision sets the provisioner_noop_provision property for all cluster requests to Cluster
	// Service, which short-circuits the full provision flow during testing
	ProvisionerNoOpProvision bool

	// ProvisionerNoOpDeprovision sets the provisioner_noop_deprovision property for all cluster requests to Cluster
	// Service, which short-circuits the full deprovision flow during testing
	ProvisionerNoOpDeprovision bool
}

func (csc *ClusterServiceClient) GetConn() *sdk.Connection { return csc.Conn }

// AddProperties injects the some addtional properties into the CSCluster Object.
func (csc *ClusterServiceClient) AddProperties(builder *cmv1.ClusterBuilder) *cmv1.ClusterBuilder {
	additionalProperties := getDefaultAdditionalProperities()
	if csc.ProvisionShardID != nil {
		additionalProperties["provision_shard_id"] = *csc.ProvisionShardID
	}
	if csc.ProvisionerNoOpProvision {
		additionalProperties["provisioner_noop_provision"] = "true"
	}
	if csc.ProvisionerNoOpDeprovision {
		additionalProperties["provisioner_noop_deprovision"] = "true"
	}
	return builder.Properties(additionalProperties)
}

// GetCSCluster creates and sends a GET request to fetch a cluster from Clusters Service
func (csc *ClusterServiceClient) GetCSCluster(ctx context.Context, internalID InternalID) (*cmv1.Cluster, error) {
	client, ok := internalID.GetClusterClient(csc.Conn)
	if !ok {
		return nil, fmt.Errorf("OCM path is not a cluster: %s", internalID)
	}
	clusterGetResponse, err := client.Get().SendContext(ctx)
	if err != nil {
		return nil, err
	}
	cluster, ok := clusterGetResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return cluster, nil
}

// GetCSClusterStatus creates and sends a GET request to fetch a cluster's status from Clusters Service
func (csc *ClusterServiceClient) GetCSClusterStatus(ctx context.Context, internalID InternalID) (*cmv1.ClusterStatus, error) {
	client, ok := internalID.GetClusterClient(csc.Conn)
	if !ok {
		return nil, fmt.Errorf("OCM path is not a cluster: %s", internalID)
	}
	clusterStatusGetResponse, err := client.Status().Get().SendContext(ctx)
	if err != nil {
		return nil, err
	}
	status, ok := clusterStatusGetResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return status, nil
}

// PostCSCluster creates and sends a POST request to create a cluster in Clusters Service
func (csc *ClusterServiceClient) PostCSCluster(ctx context.Context, cluster *cmv1.Cluster) (*cmv1.Cluster, error) {
	clustersAddResponse, err := csc.Conn.ClustersMgmt().V1().Clusters().Add().Body(cluster).SendContext(ctx)
	if err != nil {
		return nil, err
	}
	cluster, ok := clustersAddResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return cluster, nil
}

// UpdateCSCluster sends a PATCH request to update a cluster in Clusters Service
func (csc *ClusterServiceClient) UpdateCSCluster(ctx context.Context, internalID InternalID, cluster *cmv1.Cluster) (*cmv1.Cluster, error) {
	client, ok := internalID.GetClusterClient(csc.Conn)
	if !ok {
		return nil, fmt.Errorf("OCM path is not a cluster: %s", internalID)
	}
	clusterUpdateResponse, err := client.Update().Body(cluster).SendContext(ctx)
	if err != nil {
		return nil, err
	}
	cluster, ok = clusterUpdateResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return cluster, nil
}

// DeleteCSCluster creates and sends a DELETE request to delete a cluster from Clusters Service
func (csc *ClusterServiceClient) DeleteCSCluster(ctx context.Context, internalID InternalID) error {
	client, ok := internalID.GetClusterClient(csc.Conn)
	if !ok {
		return fmt.Errorf("OCM path is not a cluster: %s", internalID)
	}
	_, err := client.Delete().SendContext(ctx)
	return err
}

// GetCSNodePool creates and sends a GET request to fetch a node pool from Clusters Service
func (csc *ClusterServiceClient) GetCSNodePool(ctx context.Context, internalID InternalID) (*cmv1.NodePool, error) {
	client, ok := internalID.GetNodePoolClient(csc.Conn)
	if !ok {
		return nil, fmt.Errorf("OCM path is not a node pool: %s", internalID)
	}
	nodePoolGetResponse, err := client.Get().SendContext(ctx)
	if err != nil {
		return nil, err
	}
	nodePool, ok := nodePoolGetResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return nodePool, nil
}

// PostCSNodePool creates and sends a POST request to create a node pool in Clusters Service
func (csc *ClusterServiceClient) PostCSNodePool(ctx context.Context, clusterInternalID InternalID, nodePool *cmv1.NodePool) (*cmv1.NodePool, error) {
	client, ok := clusterInternalID.GetClusterClient(csc.Conn)
	if !ok {
		return nil, fmt.Errorf("OCM path is not a cluster: %s", clusterInternalID)
	}
	nodePoolsAddResponse, err := client.NodePools().Add().Body(nodePool).SendContext(ctx)
	if err != nil {
		return nil, err
	}
	nodePool, ok = nodePoolsAddResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return nodePool, nil
}

// UpdateCSNodePool sends a PATCH request to update a node pool in Clusters Service
func (csc *ClusterServiceClient) UpdateCSNodePool(ctx context.Context, internalID InternalID, nodePool *cmv1.NodePool) (*cmv1.NodePool, error) {
	client, ok := internalID.GetNodePoolClient(csc.Conn)
	if !ok {
		return nil, fmt.Errorf("OCM path is not a node pool: %s", internalID)
	}
	nodePoolUpdateResponse, err := client.Update().Body(nodePool).SendContext(ctx)
	if err != nil {
		return nil, err
	}
	nodePool, ok = nodePoolUpdateResponse.GetBody()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return nodePool, nil
}

// DeleteCSNodePool creates and sends a DELETE request to delete a node pool from Clusters Service
func (csc *ClusterServiceClient) DeleteCSNodePool(ctx context.Context, internalID InternalID) error {
	client, ok := internalID.GetNodePoolClient(csc.Conn)
	if !ok {
		return fmt.Errorf("OCM path is not a node pool: %s", internalID)
	}
	_, err := client.Delete().SendContext(ctx)
	return err
}

type MockClusterServiceClient struct {
	clusters  map[InternalID](*cmv1.Cluster)
	nodePools map[InternalID](*cmv1.NodePool)
}

// NewCache initializes a new Cache to allow for simple tests without needing a real CosmosDB. For production, use
// NewCosmosDBConfig instead.
func NewMockClusterServiceClient() MockClusterServiceClient {
	return MockClusterServiceClient{
		clusters:  make(map[InternalID]*cmv1.Cluster),
		nodePools: make(map[InternalID]*cmv1.NodePool),
	}
}

func (mcsc *MockClusterServiceClient) GetConn() *sdk.Connection { panic("GetConn not implemented") }

func (csc *MockClusterServiceClient) AddProperties(builder *cmv1.ClusterBuilder) *cmv1.ClusterBuilder {
	additionalProperties := getDefaultAdditionalProperities()
	return builder.Properties(additionalProperties)
}

func (mcsc *MockClusterServiceClient) GetCSCluster(ctx context.Context, internalID InternalID) (*cmv1.Cluster, error) {
	cluster, ok := mcsc.clusters[internalID]

	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return cluster, nil
}

func (mcsc *MockClusterServiceClient) PostCSCluster(ctx context.Context, cluster *cmv1.Cluster) (*cmv1.Cluster, error) {
	href := GenerateClusterHREF(cluster.Name())
	// Adding the HREF to correspond with what the full client does when crating the body
	clusterBuilder := cmv1.NewCluster()
	enrichedCluster, err := clusterBuilder.Copy(cluster).HREF(href).Build()
	if err != nil {
		return nil, err
	}
	internalID, err := NewInternalID(href)
	if err != nil {
		return nil, err
	}
	mcsc.clusters[internalID] = enrichedCluster
	return enrichedCluster, nil
}

func (mcsc *MockClusterServiceClient) UpdateCSCluster(ctx context.Context, internalID InternalID, cluster *cmv1.Cluster) (*cmv1.Cluster, error) {

	_, ok := mcsc.clusters[internalID]
	if !ok {
		return nil, fmt.Errorf("Not Found")
	}

	mcsc.clusters[internalID] = cluster
	return cluster, nil

}

func (mcsc *MockClusterServiceClient) DeleteCSCluster(ctx context.Context, internalID InternalID) error {
	_, ok := mcsc.clusters[internalID]

	if !ok {
		return fmt.Errorf("Not Found")
	}
	delete(mcsc.clusters, internalID)
	return nil
}

func (mcsc *MockClusterServiceClient) GetCSNodePool(ctx context.Context, internalID InternalID) (*cmv1.NodePool, error) {
	nodePool, ok := mcsc.nodePools[internalID]

	if !ok {
		return nil, fmt.Errorf("empty response body")
	}
	return nodePool, nil

}

func (mcsc *MockClusterServiceClient) PostCSNodePool(ctx context.Context, clusterInternalID InternalID, nodePool *cmv1.NodePool) (*cmv1.NodePool, error) {
	href := GenerateNodePoolHREF(clusterInternalID.path, nodePool.ID())
	// Adding the HREF to correspond with what the full client does when crating the body
	npBuilder := cmv1.NewNodePool()
	enrichedNodePool, err := npBuilder.Copy(nodePool).HREF(href).Build()
	if err != nil {
		return nil, err
	}

	internalID, err := NewInternalID(href)
	if err != nil {
		return nil, err
	}
	mcsc.nodePools[internalID] = enrichedNodePool
	return enrichedNodePool, nil
}

func (mcsc *MockClusterServiceClient) UpdateCSNodePool(ctx context.Context, internalID InternalID, nodePool *cmv1.NodePool) (*cmv1.NodePool, error) {
	_, ok := mcsc.nodePools[internalID]
	if !ok {
		return nil, fmt.Errorf("Not Found")
	}
	mcsc.nodePools[internalID] = nodePool
	return nodePool, nil
}

func (mcsc *MockClusterServiceClient) DeleteCSNodePool(ctx context.Context, internalID InternalID) error {
	_, ok := mcsc.nodePools[internalID]

	if !ok {
		return fmt.Errorf("Not Found")
	}
	delete(mcsc.nodePools, internalID)
	return nil
}
