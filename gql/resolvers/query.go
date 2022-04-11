package resolvers

import (
	"bytes"
	"context"
	"github.com/Masterminds/sprig/v3"
	"github.com/kfsoftware/hlf-cp-manager/gql/models"
	"github.com/kfsoftware/hlf-operator/kubectl-hlf/cmd/helpers"
	"github.com/pkg/errors"
	"text/template"
)

const tmplGoConfig = `
name: hlf-network
version: 1.0.0
client:
  organization: "{{ .Organization }}"
{{- if not .Organizations }}
organizations: {}
{{- else }}
organizations:
  {{ range $mspID, $org := .Organizations }}
  {{$mspID}}:
    mspid: {{$mspID}}
    cryptoPath: /tmp/cryptopath
    users: {}
{{- if not $org.Peers }}
    peers: []
{{- else }}
    peers:
      {{- range $peer := $org.Peers }}
      - {{ $peer.Name }}
 	  {{- end }}
{{- end }}
{{- if not $org.OrdererNodes }}
    orderers: []
{{- else }}
    orderers:
      {{- range $orderer := $org.OrdererNodes }}
      - {{ $orderer.Name }}
 	  {{- end }}

    {{- end }}
{{- end }}
{{- end }}

{{- if not .Orderers }}
orderers: []
{{- else }}
orderers:
{{- range $orderer := .Orderers }}
  {{$orderer.Name}}:
{{if $.Internal }}
    url: grpcs://{{ $orderer.PrivateURL }}
{{ else }}
    url: grpcs://{{ $orderer.PublicURL }}
{{ end }}
    grpcOptions:
      allow-insecure: false
    tlsCACerts:
      pem: |
{{ or $orderer.Status.TlsCACert $orderer.Status.TlsCert | indent 8 }}
{{- end }}
{{- end }}

{{- if not .Peers }}
peers: []
{{- else }}
peers:
  {{- range $peer := .Peers }}
  {{$peer.Name}}:
{{if $.Internal }}
    url: grpcs://{{ $peer.PrivateURL }}
{{ else }}
    url: grpcs://{{ $peer.PublicURL }}
{{ end }}
    grpcOptions:
      hostnameOverride: ""
      ssl-target-name-override: ""
      allow-insecure: false
    tlsCACerts:
      pem: |
{{ $peer.Status.TlsCACert | indent 8 }}
{{- end }}
{{- end }}

{{- if not .CertAuths }}
certificateAuthorities: []
{{- else }}
certificateAuthorities:
{{- range $ca := .CertAuths }}
  
  {{ $ca.Name }}:
{{if $.Internal }}
    url: https://{{ $ca.PrivateURL }}
{{ else }}
    url: https://{{ $ca.PublicURL }}
{{ end }}
{{if $ca.EnrollID }}
    registrar:
        enrollId: {{ $ca.EnrollID }}
        enrollSecret: {{ $ca.EnrollPWD }}
{{ end }}
    caName: ca
    tlsCACerts:
      pem: 
       - |
{{ $ca.Status.TlsCert | indent 12 }}

{{- end }}
{{- end }}

channels:
  _default:
{{- if not .Orderers }}
    orderers: []
{{- else }}
    orderers:
{{- range $orderer := .Orderers }}
      - {{$orderer.Name}}
{{- end }}
{{- end }}
{{- if not .Peers }}
    peers: {}
{{- else }}
    peers:
{{- range $peer := .Peers }}
       {{$peer.Name}}:
        discover: true
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true
{{- end }}
{{- end }}

`

func (r *queryResolver) NetworkConfig(ctx context.Context, mspID string) (*models.NetworkConfig, error) {
	tmpl, err := template.New("networkConfig").Funcs(sprig.HermeticTxtFuncMap()).Parse(tmplGoConfig)
	if err != nil {
		return nil, err
	}
	dc1 := r.DCS["dc1"]
	kubeClientset := dc1.KubeClient
	hlfClientSet := dc1.HLFClient
	var buf bytes.Buffer
	certAuths, err := helpers.GetClusterCAs(kubeClientset, hlfClientSet, "")
	if err != nil {
		return nil, err
	}
	ordOrgs, _, err := helpers.GetClusterOrderers(kubeClientset, hlfClientSet, "")
	if err != nil {
		return nil, err
	}
	ordererNodes, err := helpers.GetClusterOrdererNodes(kubeClientset, hlfClientSet, "")
	if err != nil {
		return nil, err
	}
	peerOrgs, clusterPeers, err := helpers.GetClusterPeers(kubeClientset, hlfClientSet, "")
	if err != nil {
		return nil, err
	}
	orgMap := map[string]*helpers.Organization{}
	orgFound := false
	for _, v := range ordOrgs {
		orgMap[v.MspID] = v
		if v.MspID == mspID {
			orgFound = true
		}
	}
	for _, v := range peerOrgs {
		orgMap[v.MspID] = v
		if v.MspID == mspID {
			orgFound = true
		}
	}
	if !orgFound {
		return nil, errors.Errorf("organization %s not found", mspID)
	}
	var peers []*helpers.ClusterPeer
	for _, peer := range clusterPeers {
		peers = append(peers, peer)
	}
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Peers":         peers,
		"Orderers":      ordererNodes,
		"Organizations": orgMap,
		"CertAuths":     certAuths,
		"Organization":  mspID,
		"Internal":      false,
	})
	if err != nil {
		return nil, err
	}
	return &models.NetworkConfig{
		Nc: buf.String(),
	}, nil
}
