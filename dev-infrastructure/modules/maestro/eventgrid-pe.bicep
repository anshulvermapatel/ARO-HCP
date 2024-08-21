param eventgridName string

param eventgridId string

param vnetId string

param subnetId string

param location string = resourceGroup().location

var privateDnsZoneName = 'privatelink.ts.eventgrid.azure.net'

resource eventgridPrivateEndpoint 'Microsoft.Network/privateEndpoints@2024-01-01' = {
  name: '${eventgridName}-pe'
  location: location
  properties: {
    privateLinkServiceConnections: [
      {
        name: '${eventgridName}-pe'
        properties: {
          groupIds: [
            'topicSpace'
          ]
          privateLinkServiceId: eventgridId
        }
      }
    ]
    subnet: {
      id: subnetId
    }
  }
}

resource eventgridPrivateEndpointDnsZone 'Microsoft.Network/privateDnsZones@2020-06-01' = {
  name: privateDnsZoneName
  location: 'global'
  properties: {}
}

resource eventgridPrivateDnsZoneVnetLink 'Microsoft.Network/privateDnsZones/virtualNetworkLinks@2020-06-01' = {
  parent: eventgridPrivateEndpointDnsZone
  name: 'eventgrid'
  location: 'global'
  properties: {
    registrationEnabled: false
    virtualNetwork: {
      id: vnetId
    }
  }
}

resource eventgridEndpointDnsGroup 'Microsoft.Network/privateEndpoints/privateDnsZoneGroups@2023-09-01' = {
  parent: eventgridPrivateEndpoint
  name: '${eventgridName}-dns-group'
  properties: {
    privateDnsZoneConfigs: [
      {
        name: 'config1'
        properties: {
          privateDnsZoneId: eventgridPrivateEndpointDnsZone.id
        }
      }
    ]
  }
  dependsOn: [
    eventgridPrivateDnsZoneVnetLink
  ]
}
