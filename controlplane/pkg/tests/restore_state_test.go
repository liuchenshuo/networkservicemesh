package tests

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/crossconnect"
	"github.com/networkservicemesh/networkservicemesh/controlplane/pkg/apis/local/connection"
)

func TestRestoreConnectionState(t *testing.T) {
	g := NewWithT(t)

	storage := newSharedStorage()
	srv := newNSMDFullServer(Master, storage)
	defer srv.Stop()

	srv.addFakeDataplane("dp1", "tcp:some_address")

	g.Expect(srv.nsmServer.Manager().WaitForDataplane(1 * time.Millisecond).Error()).To(Equal("Failed to wait for NSMD stare restore... timeout 1ms happened"))

	xcons := []*crossconnect.CrossConnect{}
	srv.nsmServer.Manager().RestoreConnections(xcons, "dp1")
	g.Expect(srv.nsmServer.Manager().WaitForDataplane(1 * time.Second)).To(BeNil())
}

func TestRestoreConnectionStateWrongDst(t *testing.T) {
	g := NewWithT(t)

	storage := newSharedStorage()
	srv := newNSMDFullServer(Master, storage)
	defer srv.Stop()

	srv.addFakeDataplane("dp1", "tcp:some_address")
	srv.registerFakeEndpointWithName("ns1", "IP", Worker, "ep2")

	requestConnection := &connection.Connection{
		Id:             "1",
		NetworkService: "ns1",
	}

	dstConnection := &connection.Connection{
		Id: "2",
		Mechanism: &connection.Mechanism{
			Type: connection.MechanismType_KERNEL_INTERFACE,
			Parameters: map[string]string{
				connection.WorkspaceNSEName: "nse1",
			},
		},
		NetworkService: "ns1",
	}
	xcons := []*crossconnect.CrossConnect{
		&crossconnect.CrossConnect{
			Source: &crossconnect.CrossConnect_LocalSource{
				LocalSource: requestConnection,
			},
			Destination: &crossconnect.CrossConnect_LocalDestination{
				LocalDestination: dstConnection,
			},
			Id: "1",
		},
	}
	srv.nsmServer.Manager().RestoreConnections(xcons, "dp1")
	g.Expect(srv.nsmServer.Manager().WaitForDataplane(1 * time.Second)).To(BeNil())
	g.Expect(len(srv.testModel.GetAllClientConnections())).To(Equal(0))
}
