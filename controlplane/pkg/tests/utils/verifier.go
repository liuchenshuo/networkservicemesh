package utils

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/crossconnect"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/model"
)

// Verifier is a common verifier interface to be used in tests
type Verifier interface {
	Verify(t *testing.T)
}

// ModelVerifier is a Verifier to check model.Model state
type ModelVerifier struct {
	model     model.Model
	verifiers []Verifier
}

// NewModelVerifier is a constructor for ModelVerifier
func NewModelVerifier(model model.Model) *ModelVerifier {
	return &ModelVerifier{
		model:     model,
		verifiers: []Verifier{},
	}
}

// EndpointNotExists is a builder method to add check if model.Endpoint with
//   Endpoint.NetworkServiceEndpoint.Name == name
// doesn't exist in v.model
func (v *ModelVerifier) EndpointNotExists(name string) *ModelVerifier {
	v.verifiers = append(v.verifiers, &endpointVerifier{
		exists: false,
		name:   name,

		model: v.model,
	})

	return v
}

// EndpointExists is a builder method to add check if model.Endpoint with
//   Endpoint.NetworkServiceEndpoint.Name == name
//   Endpoint.NetworkServiceManager.Name == nsm
// exists in v.model
func (v *ModelVerifier) EndpointExists(name, nsm string) *ModelVerifier {
	v.verifiers = append(v.verifiers, &endpointVerifier{
		exists: true,
		name:   name,
		nsm:    nsm,

		model: v.model,
	})

	return v
}

// ClientConnectionNotExists is a builder method to add check if model.ClientConnection with
//   ConnectionID == connectionID
// doesn't exist in v.model
func (v *ModelVerifier) ClientConnectionNotExists(connectionID string) *ModelVerifier {
	v.verifiers = append(v.verifiers, &clientConnectionVerifier{
		exists:       false,
		connectionID: connectionID,

		model: v.model,
	})

	return v
}

// ClientConnectionExists is a builder method to add check if model.ClientConnection with
//   ConnectionID == connectionID
//   Xcon.Source.Id = srcID
//   Xcon.Destination.Id = dst.ID
//   RemoteNsm.Name = remoteNSM
//   Endpoint.NetworkServiceEndpoint.Name = nse
//   Dataplane.RegisteredName = dataplane
// exists in v.model
func (v *ModelVerifier) ClientConnectionExists(connectionID, srcID, dstID, remoteNSM, nse, dataplane string) *ModelVerifier {
	v.verifiers = append(v.verifiers, &clientConnectionVerifier{
		exists:       true,
		connectionID: connectionID,
		srcID:        srcID,
		dstID:        dstID,
		remoteNSM:    remoteNSM,
		nse:          nse,
		dataplane:    dataplane,

		model: v.model,
	})

	return v
}

// DataplaneNotExists is a builder method to add check if model.Dataplane with
//   RegisteredName = name
// doesn't exist in v.model
func (v *ModelVerifier) DataplaneNotExists(name string) *ModelVerifier {
	v.verifiers = append(v.verifiers, &dataplaneVerifier{
		exists: false,
		name:   name,

		model: v.model,
	})

	return v
}

// DataplaneExists is a builder method to add check if model.Dataplane with
//   RegisteredName = name
// exists in v.model
func (v *ModelVerifier) DataplaneExists(name string) *ModelVerifier {
	v.verifiers = append(v.verifiers, &dataplaneVerifier{
		exists: true,
		name:   name,

		model: v.model,
	})

	return v
}

// Verify invokes all stored checks
func (v *ModelVerifier) Verify(t *testing.T) {
	for _, verifier := range v.verifiers {
		verifier.Verify(t)
	}
}

type endpointVerifier struct {
	exists bool
	name   string
	nsm    string

	model model.Model
}

func (v *endpointVerifier) Verify(t *testing.T) {
	g := NewWithT(t)

	nse := v.model.GetEndpoint(v.name)
	if !v.exists {
		g.Expect(nse).To(BeNil())
		return
	}

	g.Expect(nse).NotTo(BeNil())

	g.Expect(nse.Endpoint.GetNetworkServiceManager().GetName()).To(Equal(v.nsm))
}

type clientConnectionVerifier struct {
	exists       bool
	connectionID string
	srcID        string
	dstID        string
	remoteNSM    string
	nse          string
	dataplane    string

	model model.Model
}

func (v *clientConnectionVerifier) Verify(t *testing.T) {
	g := NewWithT(t)

	connection := v.model.GetClientConnection(v.connectionID)
	if !v.exists {
		g.Expect(connection).To(BeNil())
		return
	}

	g.Expect(connection).NotTo(BeNil())

	v.verifyXcon(connection.Xcon, t)
	g.Expect(connection.RemoteNsm.GetName()).To(Equal(v.remoteNSM))
	g.Expect(connection.Endpoint.GetNetworkServiceEndpoint().GetName()).To(Equal(v.nse))
	g.Expect(connection.DataplaneRegisteredName).To(Equal(v.dataplane))
}

func (v *clientConnectionVerifier) verifyXcon(xcon *crossconnect.CrossConnect, t *testing.T) {
	g := NewWithT(t)

	if source := xcon.GetLocalSource(); source != nil {
		g.Expect(source.GetId()).To(Equal(v.srcID))
	} else if source := xcon.GetRemoteSource(); source != nil {
		g.Expect(source.GetId()).To(Equal(v.srcID))
	} else {
		t.Fatalf("Expected xcon.Source not to be nil")
	}

	if destination := xcon.GetLocalDestination(); destination != nil {
		g.Expect(destination.GetId()).To(Equal(v.dstID))
	} else if destination := xcon.GetRemoteDestination(); destination != nil {
		g.Expect(destination.GetId()).To(Equal(v.dstID))
	} else {
		t.Fatalf("Expected xcon.Destination not to be nil")
	}
}

type dataplaneVerifier struct {
	exists bool
	name   string

	model model.Model
}

func (v *dataplaneVerifier) Verify(t *testing.T) {
	g := NewWithT(t)

	dataplane := v.model.GetDataplane(v.name)
	if !v.exists {
		g.Expect(dataplane).To(BeNil())
		return
	}

	g.Expect(dataplane).NotTo(BeNil())
}
