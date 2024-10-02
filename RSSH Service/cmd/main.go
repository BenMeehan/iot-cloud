// package main

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"net"

// 	"os"

// 	"golang.org/x/crypto/ssh"
// )

// func main() {
// 	// SSH server configuration
// 	config := &ssh.ServerConfig{
// 		PasswordCallback: func(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
// 			if string(password) == "ghost" {
// 				return nil, nil
// 			}
// 			return nil, fmt.Errorf("password rejected for %q", c.User())
// 		},
// 	}

// 	// Load private key (replace with your actual private key)
// 	val, err := os.ReadFile("/home/ghost/Desktop/iot-project/iot-cloud/RSSH Service/cmd/priv.txt")
// 	if err != nil {
// 		log.Fatalf("Val: %v", err)
// 	}
// 	privateKey, err := ssh.ParsePrivateKey(val)
// 	if err != nil {
// 		log.Fatalf("Failed to parse private key: %v", err)
// 	}
// 	config.AddHostKey(privateKey)

// 	// Start listening on port 22 for SSH connections
// 	listener, err := net.Listen("tcp", "0.0.0.0:22")
// 	if err != nil {
// 		log.Fatalf("Failed to listen on port 22: %v", err)
// 	}
// 	defer listener.Close()
// 	log.Println("SSH server listening on port 22")

// 	for {
// 		nConn, err := listener.Accept()
// 		if err != nil {
// 			log.Printf("Failed to accept incoming connection: %v", err)
// 			continue
// 		}
// 		// Handle the connection in a new goroutine
// 		go handleConnection(nConn, config)
// 	}
// }

// func handleConnection(nConn net.Conn, config *ssh.ServerConfig) {
// 	// Establish the SSH connection
// 	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
// 	if err != nil {
// 		log.Printf("Failed to handshake: %v", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// Log SSH connection details
// 	log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())

// 	// Handle requests
// 	go ssh.DiscardRequests(reqs)

// 	// Handle channel requests (port forwarding)
// 	for newChannel := range chans {
// 		// Only handle forwarded-tcpip channels
// 		fmt.Println(newChannel.ChannelType())
// 		if newChannel.ChannelType() == "forwarded-tcpip" || newChannel.ChannelType() == "session" {
// 			channel, requests, err := newChannel.Accept()
// 			if err != nil {
// 				log.Printf("Could not accept channel: %v", err)
// 				continue
// 			}
// 			log.Printf("Forwarded-tcpip channel opened")

// 			// Handle the channel in a separate goroutine
// 			go handleForwardedChannel(channel, requests)
// 		} else {
// 			// Reject any non-forwarded-tcpip channel requests
// 			newChannel.Reject(ssh.UnknownChannelType, "Channel type not supported")
// 		}
// 	}
// }

// func handleForwardedChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
// 	defer channel.Close()

// 	// Process requests on the channel
// 	go func() {
// 		for req := range requests {
// 			if req.WantReply {
// 				req.Reply(false, nil)
// 			}
// 		}
// 	}()

// 	// Forward the traffic from the SSH channel to the device and vice versa
// 	localConn, err := net.Dial("tcp", "localhost:8080") // Replace with the desired local service/port
// 	if err != nil {
// 		log.Printf("Failed to connect to the local service: %v", err)
// 		return
// 	}
// 	defer localConn.Close()

// 	// Use io.Copy to forward data between the SSH channel and the local connection
// 	go func() { _, _ = io.Copy(localConn, channel) }()
// 	_, _ = io.Copy(channel, localConn)
// }

////

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

func main() {
	// SSH server configuration
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if string(password) == "ghost" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	// Load private key (replace with your actual private key)
	val, err := os.ReadFile("/home/ghost/Desktop/iot-project/iot-cloud/RSSH Service/cmd/priv.txt")
	if err != nil {
		log.Fatalf("Val: %v", err)
	}
	privateKey, err := ssh.ParsePrivateKey(val)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	config.AddHostKey(privateKey)

	// Start listening on port 2222 for SSH connections (use 22 for production)
	listener, err := net.Listen("tcp", "0.0.0.0:2000")
	if err != nil {
		log.Fatalf("Failed to listen on port 2000: %v", err)
	}
	defer listener.Close()
	log.Println("SSH server listening on port 2000")

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}
		// Handle the connection in a new goroutine
		go handleConnection(nConn, config)
	}
}

func handleConnection(nConn net.Conn, config *ssh.ServerConfig) {
	// Establish the SSH connection
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer conn.Close()

	// Log SSH connection details
	log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())
	// log.Println(chans)
	// for newChannel := range chans {
	// 	log.Println(newChannel)
	// }

	// Handle requests
	go ssh.DiscardRequests(reqs)

	// Handle different channel requests: session and forwarded-tcpip
	for newChannel := range chans {
		log.Println(newChannel.ChannelType())
		switch newChannel.ChannelType() {
		case "session":
			log.Println("Connection channel: session")
			go handleSessionChannel(newChannel)
		case "forwarded-tcpip":
			log.Println("Connection channel: forwarded-tcpip")
			channel, requests, err := newChannel.Accept()
			if err != nil {
				log.Printf("Could not accept forwarded-tcpip channel: %v", err)
				continue
			}
			log.Printf("Forwarded-tcpip channel opened")
			go handleForwardedChannel(channel, requests)
		default:
			// Reject any non-session and non-forwarded-tcpip channel requests
			newChannel.Reject(ssh.UnknownChannelType, "Channel type not supported")
		}
	}
}

func handleSessionChannel(newChannel ssh.NewChannel) {
	// Accept the session channel request
	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept session channel: %v", err)
		return
	}
	defer channel.Close()

	// Handle the session's requests (exec, shell, etc.)
	for req := range requests {
		switch req.Type {
		case "shell":
			if req.WantReply {
				req.Reply(true, nil)
			}
			handleShell(channel)
		case "exec":
			if req.WantReply {
				req.Reply(true, nil)
			}
			handleExec(channel, req.Payload)
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func handleShell(channel ssh.Channel) {
	// Launch a bash shell
	cmd := exec.Command("/bin/bash")
	cmd.Env = os.Environ()
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel

	// Start the shell
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to start shell: %v", err)
		return
	}
}

func handleExec(channel ssh.Channel, payload []byte) {
	// Extract command from payload (RFC4254 Section 6.5)
	var execPayload struct {
		Command string
	}
	if err := ssh.Unmarshal(payload, &execPayload); err != nil {
		log.Printf("Failed to unmarshal exec payload: %v", err)
		channel.Stderr().Write([]byte("Invalid exec request\n"))
		return
	}

	// Execute the requested command
	cmd := exec.Command("sh", "-c", execPayload.Command)
	cmd.Stdout = channel
	cmd.Stderr = channel

	// Run the command
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		channel.Stderr().Write([]byte(fmt.Sprintf("Failed to execute command: %v\n", err)))
	}
}

func handleForwardedChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	// Process requests on the channel
	go func() {
		for req := range requests {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}()

	// Forward the traffic from the SSH channel to the device and vice versa
	localConn, err := net.Dial("tcp", "localhost:8080") // Replace with the desired local service/port
	if err != nil {
		log.Printf("Failed to connect to the local service: %v", err)
		return
	}
	defer localConn.Close()

	// Use io.Copy to forward data between the SSH channel and the local connection
	go func() { _, _ = io.Copy(localConn, channel) }()
	_, _ = io.Copy(channel, localConn)
}
