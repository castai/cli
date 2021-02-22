package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

type AuthorizedKeyConfig struct {
	PrivateKey    []byte
	User          string
	Addr          string
	AuthorizedKey []byte
}

// AddAuthorizedKey copies public key to remove machine the same as ssh-copy-id command.
func AddAuthorizedKey(ctx context.Context, cfg AuthorizedKeyConfig) error {
	signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return err
	}
	conn, err := ssh.Dial("tcp", cfg.Addr, &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	out, err := runCommand(conn, "cat ~/.ssh/authorized_keys", nil)
	if err != nil {
		return fmt.Errorf("catting authorized_keys: %s: %w", out, err)
	}
	// Check if key is not already added.
	if strings.Contains(out, strings.TrimSpace(string(cfg.AuthorizedKey))) {
		return nil
	}

	out, err = runCommand(conn, "cat >> ~/.ssh/authorized_keys", bytes.NewBuffer(cfg.AuthorizedKey))
	if err != nil {
		return fmt.Errorf("adding new key: %s: %w", out, err)
	}
	return nil
}

func runCommand(conn *ssh.Client, cmd string, in io.Reader) (string, error) {
	sess, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()

	out := &bytes.Buffer{}
	sess.Stdout = out
	sess.Stdin = in
	if err := sess.Run(cmd); err != nil {
		return "", err
	}
	return out.String(), nil
}
