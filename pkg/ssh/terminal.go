package ssh

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type TerminalConfig struct {
	PrivateKey []byte
	User       string
	Addr       string
}

func Terminal(ctx context.Context, cfg TerminalConfig) error {
	// Create SSH connection.
	signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("parsing private key: %w", err)
	}

	conn, err := ssh.Dial("tcp", cfg.Addr, &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// TODO: Implement server host key validation.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}

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
