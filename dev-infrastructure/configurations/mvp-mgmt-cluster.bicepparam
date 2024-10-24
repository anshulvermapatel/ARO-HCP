using '../templates/mgmt-cluster.bicep'

param kubernetesVersion = '1.30.4'
param vnetAddressPrefix = '10.132.0.0/14'
param subnetPrefix = '10.132.8.0/21'
param podSubnetPrefix = '10.132.64.0/18'
param aksClusterName = take('aro-hcp-mgmt-cluster-${uniqueString('mgmt-cluster')}', 63)
param aksKeyVaultName = 'aks-kv-aro-hcp-dev-mc-1'
param systemAgentMinCount = 2
param systemAgentMaxCount = 3
param systemAgentVMSize = 'Standard_D2s_v3'
param aksSystemOsDiskSizeGB = 32
param userAgentMinCount = 1
param userAgentMaxCount = 9
param userAgentVMSize = 'Standard_D4s_v3'
param aksUserOsDiskSizeGB = 100
param userAgentPoolAZCount = 3
param persist = true

param deployMaestroConsumer = true
param maestroKeyVaultName = 'maestro-kv-aro-hcp-dev'
param maestroEventGridNamespacesName = 'maestro-eventgrid-aro-hcp-dev'
param maestroCertDomain = 'selfsigned.maestro.keyvault.aro-dev.azure.com'

param baseDNSZoneName = 'hcp.osadev.cloud'
param regionalDNSSubdomain = 'westus3'

param acrPullResourceGroups = [regionalResourceGroup, 'global']

// These parameters are always overridden in the Makefile
param currentUserId = ''
param regionalResourceGroup = ''
