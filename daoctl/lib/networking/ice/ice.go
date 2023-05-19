package ice

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	//"github.com/davecgh/go-spew/spew"

	"github.com/go-logr/logr"
	"github.com/pion/ice/v2"
	"github.com/workbenchapp/worknet/daoctl/lib/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// So the connection flow is:
// new client sends their auth info to SvenServer_auth
// server side starts its agent, sends its auth to SvenClient_auth
// both SvenClient Dials SvenServer's Accept, which trigger the candidate exchange, and the attempts to connect.

// if i'm right, this will allow the server to talk to multiple clients
func ListenForICEConnectionRequest(
	ctx context.Context,
	nodeName string,
	localProxyAddr string,
) {
	tracer := telemetry.TracerFromContext(ctx)
	remoteAuth := pull(ctx, nodeName+"_auth")
	log := logr.FromContextOrDiscard(ctx)

	for {
		log.Info("Waiting for client auth info from signal server", "name", nodeName+"_auth")
		select {
		case <-ctx.Done():
			return
		case authValues := <-remoteAuth:
			ctx, span := tracer.Start(
				ctx,
				"iceConnectionRequest",
				trace.WithNewRoot(),
			)
			//spew.Dump(authValues)
			remoteUfrag := authValues["ufrag"]
			remotePwd := authValues["pwd"]
			sessionId := authValues["sessionid"]
			otherNodeName := authValues["nodename"]
			span.SetAttributes(
				attribute.String("remoteUfrag", remoteUfrag),
				attribute.String("otherNodeName", otherNodeName),
				attribute.String("localProxyAddr", localProxyAddr),
				attribute.String("nodeName", nodeName),
			)

			// TODO: I don't think i need to track these
			serverListenConn := &Conection{}

			go func() {
				defer span.End()
				if err := serverListenConn.iceConnectionRequest(
					ctx,
					nodeName,
					otherNodeName,
					true,
					remoteUfrag,
					remotePwd,
					sessionId,
					localProxyAddr,
				); err != nil {
					// OnConnectionStateChange sometimes cancels the context
					// because the state is change to failed (seems to be 30s
					// timeout). This is a fairly normal state of operations,
					// so even though it's technically an error, it's not quite
					// worth logging.
					if !strings.Contains(err.Error(), "connecting canceled") {
						log.Error(err, "iceConnectionRequest failed")
					} else {
						log.V(1).Info("iceConnectionRequest context cancelled")
					}
					span.SetAttributes(attribute.String("error", err.Error()))
				}
			}()
		}
	}
}

// Client request to connect to NAT's server, and then proxy it's connection to a local port
func MakeNewICEConnectionRequest(
	ctx context.Context,
	nodeName string,
	serverNodeName string,
	localProxyAddr string, // is isServer, then the tcp server that we're making available on the client, if !isServer, then the local client tcp port you can use to talk to the remote service
) (conn *Conection) {
	tracer := telemetry.TracerFromContext(ctx)
	ctx, span := tracer.Start(
		ctx,
		"MakeNewICEConnectionRequest",
		trace.WithNewRoot(),
	)
	defer span.End()
	if connList == nil {
		connList = make(map[string]*Conection)
	}
	connName := nodeName + "_" + serverNodeName
	span.SetAttributes(
		attribute.String("nodeName", nodeName),
		attribute.String("serverNodeName", serverNodeName),
	)
	log := logr.FromContextOrDiscard(ctx).WithValues("node", nodeName, "server", serverNodeName)

	//USE connName in in log.Withfield(connconnName), and then pass that into the isConnectionRequest, so the label gets added ...

	conn, ok := connList[connName]
	if ok {
		if conn.Busy() {
			return conn
		}
		if conn != nil && conn.Cancel != nil {
			log.Info("CANCELLING to start new ICE client attempt")
			conn.Cancel()
		}
	}

	conn = &Conection{}
	connList[nodeName+"_"+serverNodeName] = conn
	go conn.iceConnectionRequest(
		ctx,
		nodeName,
		serverNodeName,
		false,
		"",
		"",
		time.Now().String(),
		localProxyAddr,
	)
	return conn
}

var connList map[string]*Conection

type Conection struct {
	Cancel                  func()
	iceConn                 *ice.Conn
	iceConnState            ice.ConnectionState
	nodeName, otherNodeName string
}

// if we're still doing something, don't start a new client connection
func (c *Conection) Busy() bool {
	if c == nil {
		return false
	}
	if c.iceConn == nil {
		return false
	}
	if c.iceConnState == ice.ConnectionStateFailed || c.iceConnState == ice.ConnectionStateDisconnected || c.iceConnState == ice.ConnectionStateClosed {
		return false
	}
	return true
}

type Status struct {
	nodeName, otherNodeName string
	Status                  string
}

func GetConnectionStates() map[string]Status {
	status := make(map[string]Status)
	for name, conn := range connList {
		status[name] = Status{
			Status: conn.iceConnState.String(),
		}
	}
	return status
}

func (conn *Conection) iceConnectionRequest(
	ctx context.Context,
	nodeName, otherNodeName string,
	isServer bool,
	otherNodeUfrag, otherNodePwd, sessionId string,
	localProxyAddr string, // is isServer, then the tcp server that we're making available on the client, if !isServer, then the local client tcp port you can use to talk to the remote service
) (err error) {
	tracer := telemetry.TracerFromContext(ctx)
	parentCtx, span := tracer.Start(
		ctx,
		"iceConnectionRequest",
	)
	span.SetAttributes(
		attribute.String("nodeName", nodeName),
		attribute.String("otherNodeName", otherNodeName),
	)
	defer span.End()
	log := logr.FromContextOrDiscard(parentCtx).WithName("iceConnectionRequest").WithValues("node", nodeName, "sessionId", sessionId)

	log.Info("Start")

	ctx, conn.Cancel = context.WithCancel(parentCtx)

	go func() {
		<-ctx.Done()
		log.Info("CancelService()")
		conn.Cancel()
		log.Info("CancelService() done")
	}()

	iceAgent, err := ice.NewAgent(&ice.AgentConfig{
		NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4},
	})
	if err != nil {
		return err
	}
	defer iceAgent.Close()
	// When ICE Connection state has change print to stdout
	if err = iceAgent.OnConnectionStateChange(func(c ice.ConnectionState) {
		span.AddEvent(
			"OnConnectionStateChange",
			trace.WithAttributes(attribute.String("conn.state", c.String())),
		)
		conn.iceConnState = c
		log.V(2).Info("ICE Connection State has changed", "state", c.String())
		if c == ice.ConnectionStateFailed || c == ice.ConnectionStateDisconnected || c == ice.ConnectionStateClosed {
			conn.Cancel()
			return
		}
	}); err != nil {
		return err
	}

	clientAuthCtx, CancelClientAuth := context.WithCancel(ctx) // TODO: yeah, this is bad
	var remoteAuth <-chan SignalValues
	if !isServer {
		remoteAuth = pull(clientAuthCtx, nodeName+"_auth")
	}
	candidatesCtx, CancelCandidatesPull := context.WithCancel(ctx)
	remoteCandidates := pull(candidatesCtx, nodeName)

	go func() {
		_, span := tracer.Start(parentCtx, "pullRemoteCandidate")
		defer span.End()

		// TODO: this needs a timeout so we can also cancel the context
		candidate := <-remoteCandidates
		log.V(2).Info("receiving", "candidate", candidate["name"])
		//spew.Dump(candidate)
		count, err := strconv.Atoi(candidate["count"])
		if err != nil {
			return
		}
		for i := 0; i < count; i++ {
			cName := fmt.Sprintf("candidate%d", i)
			cString, ok := candidate[cName]
			if !ok {
				continue
			}

			c, err := ice.UnmarshalCandidate(cString)
			if err != nil {
				log.V(2).Error(err, "Skipping", "candidate", cName)
				continue
			}

			span.SetAttributes(attribute.String("candidate", c.String()))

			if err := iceAgent.AddRemoteCandidate(c); err != nil {
				log.V(2).Error(err, "Skipping", "candidate", cName)
				continue
			}
		}
		log.V(2).Info("exiting candidates pull")
		CancelCandidatesPull()
	}()

	// Get the local auth details and send to remote peer
	localUfrag, localPwd, err := iceAgent.GetLocalUserCredentials()
	if err != nil {
		return err
	}

	// spew.Dump(localUfrag)
	// spew.Dump(localPwd)

	// TODO: except that it means if the server isn't listening when the client starts, we all just do nothing
	// need a timeout on the client...
	if !isServer {
		// I'd say only the Client kicks things off here
		// the server will likely respond instantly-ish with its auth
		log.V(2).Info("Sending client side auth info to", "othernode", otherNodeName+"_auth")
		push(ctx, otherNodeName+"_auth", SignalValues{
			"ufrag":     localUfrag,
			"pwd":       localPwd,
			"sessionid": sessionId,
			"nodename":  nodeName,
		})
	}

	// When we have gathered a new ICE Candidate send it to the remote peer
	count := 0
	signal := SignalValues{
		"name": "candidate",
	}
	if err = iceAgent.OnCandidate(func(c ice.Candidate) {
		span.AddEvent("OnCandidate")
		if c == nil {
			// Last candidate gathered
			log.V(2).Info("Sending Candidates", "count", count)
			signal["count"] = fmt.Sprintf("%d", count)
			push(ctx, otherNodeName, signal)
			return
		} else {
			span.AddEvent("OnCandidate",
				trace.WithAttributes(attribute.String("candidate", c.String())))
		}
		cName := fmt.Sprintf("candidate%d", count)
		log.V(2).Info("Gathered", "candidate", cName)
		signal[cName] = c.Marshal()
		count = count + 1
	}); err != nil {
		return err
	}

	remoteUfrag := otherNodeUfrag // set by server
	remotePwd := otherNodePwd
	if !isServer {
		log.V(2).Info("wait for pull info", "name", nodeName+"_auth")

		// TODO: this needs a timeout so we can also cancel it.
		authValues := <-remoteAuth
		CancelClientAuth()
		//spew.Dump(authValues)
		remoteUfrag = authValues["ufrag"]
		remotePwd = authValues["pwd"]
		remoteSessionId := authValues["sessionid"]
		remoteNodeName := authValues["nodename"]

		log.V(2).Info("Received auth info", "sessionID", sessionId, "remote node", remoteNodeName, "remote sessionId", remoteSessionId)
	} else {
		log.V(2).Info("push info", "othernode", otherNodeName+"_auth")
		push(ctx, otherNodeName+"_auth", SignalValues{
			"ufrag": localUfrag,
			"pwd":   localPwd,
		})
	}

	// spew.Dump(remoteUfrag)
	// spew.Dump(remotePwd)

	log.V(2).Info("GatherCandidates")
	if err = iceAgent.GatherCandidates(); err != nil {
		return err
	}

	// These are the blocking calls that everything above is a goroutine for negotiation (which is why we need the ufrag*pwd before the candidates)
	// Start the ICE Agent. One side must be controlled, and the other must be controlling
	if isServer {
		log.V(2).Info("Server listening for connection")
		conn.iceConn, err = iceAgent.Accept(ctx, remoteUfrag, remotePwd)
	} else {
		log.V(2).Info("Contact Server")
		conn.iceConn, err = iceAgent.Dial(ctx, remoteUfrag, remotePwd)
	}
	if err != nil {
		return err
	}
	log.V(2).Info("Tunnel Connected")

	// connect / listen to something
	remoteAddr, err := net.ResolveUDPAddr("udp", localProxyAddr)
	if err != nil {
		return err
	}
	var something *net.UDPConn
	if isServer {
		something, err = net.DialUDP("udp", nil, remoteAddr)
		if err != nil {
			return err
		}
	} else {
		// TODO: this is a use once - listen - if we extract both the Dial and Listen, we could also re-use the tunnel..
		// BUT, I think that would require the server side to have a sidechannel or mux-conn to Dial again?
		log.V(2).Info("Waiting for connection to udp", "localProxyAddr", localProxyAddr)

		something, err = net.ListenUDP("udp", remoteAddr)
		if err != nil {
			return err
		}
	}

	defer something.Close()
	defer conn.iceConn.Close()
	if isServer {
		log.V(2).Info("-------------------- Start tunnel io.copy from tunnel to udp")
	} else {
		log.V(2).Info("-------------------- Start tunnel io.copy from udp to tunnel")
	}

	err = forwardConnection(ctx, conn.iceConn, something, remoteAddr)
	if err != nil {
		return err
	}
	log.V(2).Info("tunnel io.copy done")

	// TODO: this doesn't happen on the server because the conn.Read() has no timeout / poll...
	// And I'm guessing that means the wrong server agent is getting the pull() on either the auth or candidate info
	log.V(2).Info("Exit OK\n", "sessionId", sessionId)
	return err
}

// TODO: well shit. I don't think this disconnects or stops on the server side.
func forwardConnection(ctx context.Context, tunnel io.ReadWriteCloser, udp *net.UDPConn, defaultRemoteAddr *net.UDPAddr) error {
	log := logr.FromContextOrDiscard(ctx).WithName("forwardConnection")

	remoteAddr := defaultRemoteAddr
	var err error
	// This isn't tcp
	go func() error {
		buf := make([]byte, 1500)
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("forwardConnection(from, to) done")
			default:
				n, err := tunnel.Read(buf)
				if err != nil {
					log.V(2).Error(err, "read error(from, to)")
					continue
				}
				//log.V(2).Info("read forwardConnection(from, to) %d bytes\n", n)

				if udp.RemoteAddr() == nil {
					_, err = udp.WriteToUDP(buf[:n], remoteAddr)
				} else {
					_, err = udp.Write(buf[:n])
				}
				if err != nil {
					log.V(2).Error(err, "wrote error(from, to)")
					continue
				}
				//log.V(2).Info("wrote forwardConnection(from, to) %d bytes\n", wroteN)
			}
		}
	}()
	go func() error {
		buf := make([]byte, 1500)
		for {
			select {
			case <-ctx.Done():
				//log.V(2).Info("stopped proxying to remote peer due to closed connection")
				return fmt.Errorf("forwardConnection(to, from) done")
			default:
				var n int
				n, remoteAddr, err = udp.ReadFromUDP(buf)
				if err != nil {
					log.V(2).Error(err, "read error", "remoteAddr", remoteAddr.String())
					continue
				}
				//log.V(2).Info("read forwardConnection(to, from) %d\n", n)

				_, err = tunnel.Write(buf[:n])
				if err != nil {
					log.V(2).Error(err, "read error", "remoteAddr", remoteAddr.String())
					continue
				}
				//log.V(2).Info("wrote forwardConnection(to, from) %d\n", wroteN)
			}
		}
	}()

	<-ctx.Done()

	log.V(2).Info("DONE")
	return nil
}
