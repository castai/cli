package ssh

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

type ConnectConfig struct {
	PrivateKey []byte
	User       string
	Addr       string
}

type Terminal interface {
	Connect(ctx context.Context, cfg ConnectConfig) error
}

func NewTerminal(log logrus.FieldLogger) Terminal {
	return &terminal{log: log}
}

type terminal struct {
	log logrus.FieldLogger
}

func (t *terminal) Connect(ctx context.Context, cfg ConnectConfig) error {
	// Create SSH connection.
	conn, err := t.dial(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create SSH session.
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// Exit on context cancel.
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// Create interactive terminal.
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal make raw: %w", err)
	}
	defer term.Restore(fd, state)

	w, h, err := term.GetSize(fd)
	if err != nil {
		return fmt.Errorf("terminal get size: %w", err)
	}

	// Setup session shell.
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	if err := sess.RequestPty(termType, h, w, modes); err != nil {
		return fmt.Errorf("session xterm: %w", err)
	}

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	sess.Stdin = os.Stdin

	if err := sess.Shell(); err != nil {
		return fmt.Errorf("session shell: %w", err)
	}
	if err := sess.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 130:
				return nil
			}
		}
		return fmt.Errorf("ssh: %w", err)
	}
	return nil
}

func (t *terminal) dial(ctx context.Context, cfg ConnectConfig) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	knownhostsFilePath := filepath.Join(home, ".ssh", "known_hosts")
	if err := ensureKnownHostsFile(knownhostsFilePath); err != nil {
		return nil, err
	}

	var conn *ssh.Client
	connTimeout := time.After(2 * time.Minute)
	for {
		var connerr error
		conn, connerr = ssh.Dial("tcp", cfg.Addr, &ssh.ClientConfig{
			User: cfg.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: t.validateHostKey(knownhostsFilePath),
			Timeout:         15 * time.Second,
		})
		if connerr == nil {
			return conn, nil
		}
		select {
		case <-time.After(2 * time.Second):
			t.log.Debugf("ssh connection failed, reconnecting: %v", connerr)
		case <-connTimeout:
			return nil, fmt.Errorf("connection timeout: %w", connerr)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// validateHostKey tries to validate known hosts file and adds new entry if host is unknown.
func (t *terminal) validateHostKey(knownhostsFilePath string) ssh.HostKeyCallback {
	return func(addr string, remote net.Addr, key ssh.PublicKey) error {
		validate, err := knownhosts.New(knownhostsFilePath)
		if err != nil {
			return err
		}

		err = validate(addr, remote, key)
		var keyErr *knownhosts.KeyError
		// Add host key to know hosts only if key is unknown.
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			f, err := os.OpenFile(knownhostsFilePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer f.Close()
			hostname, _, err := net.SplitHostPort(addr)
			if err != nil {
				return err
			}
			key := fmt.Sprintf("%s %s %s\n", hostname, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))
			_, err = f.WriteString(key)
			if err != nil {
				return err
			}
			t.log.Warnf("added %s to known hosts in %s", hostname, knownhostsFilePath)
			return nil
		}
		return err
	}
}

func ensureKnownHostsFile(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := f.Chmod(0644); err != nil {
			return err
		}
		defer f.Close()
		return nil
	}
	return err
}
