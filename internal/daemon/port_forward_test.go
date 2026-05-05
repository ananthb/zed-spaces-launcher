package daemon

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestValidatePortForwardKey(t *testing.T) {
	if err := validatePortForwardKey(portForwardKey{Codespace: "cs", RemotePort: 1455, LocalPort: 1455}); err != nil {
		t.Fatalf("valid key: %v", err)
	}
	for _, key := range []portForwardKey{
		{RemotePort: 1455, LocalPort: 1455},
		{Codespace: "cs", RemotePort: 0, LocalPort: 1455},
		{Codespace: "cs", RemotePort: 1455, LocalPort: 70000},
	} {
		if err := validatePortForwardKey(key); err == nil {
			t.Fatalf("validatePortForwardKey(%#v) expected error", key)
		}
	}
}

func TestFriendlyPortForwardMessageClassifiesBindErrors(t *testing.T) {
	key := portForwardKey{Codespace: "expert-spoon", RemotePort: 1455, LocalPort: 1455}
	msg := friendlyPortForwardMessage(key, errors.New("exit status 1"), "listen tcp 127.0.0.1:1455: bind: address already in use")
	if !strings.Contains(msg, "localhost port 1455 is already in use") {
		t.Fatalf("msg = %q", msg)
	}
	if !strings.Contains(msg, "codespace forward") {
		t.Fatalf("msg should mention codespace forwards: %q", msg)
	}
}

func TestPortForwardManagerRejectsManagedLocalPortConflict(t *testing.T) {
	manager := newPortForwardManager()
	manager.forwards[portForwardKey{Codespace: "one", RemotePort: 1455, LocalPort: 1455}] = &managedPortForward{}

	err := manager.Start("two", 1455, 1455)
	if err == nil {
		t.Fatal("expected local port conflict")
	}
	if !strings.Contains(err.Error(), "already forwarded to one:1455") {
		t.Fatalf("err = %v", err)
	}
}

func TestEnsureLocalPortAvailableDetectsBoundPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	err = ensureLocalPortAvailable(port)
	if err == nil {
		t.Fatal("expected bound port error")
	}
	if !strings.Contains(err.Error(), "localhost port "+strconv.Itoa(port)+" is already in use") {
		t.Fatalf("err = %v", err)
	}
}
