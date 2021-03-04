package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/castai/cli/pkg/client"
	"github.com/castai/cli/pkg/config"
	"github.com/castai/cli/pkg/ssh"
)

func newTestRootCmd() *cobra.Command {
	api := client.NewMock()
	log := logrus.New()
	return NewRootCmd(log, &config.Config{}, api, &mockTerminal{}, &mockIpify{})
}

func TestCommands(t *testing.T) {
	t.Run("cluster list", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(root, "cluster", "list")
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` ID  NAME            STATUS  CLOUDS  REGION                       AGE      
 c1  test-cluster-1  ready   aws      Europe Central (Frankfurt)  just now`
		require.Equal(t, expected+" \n", out)
	})

	t.Run("cluster create from configuration", func(t *testing.T) {
		root := newTestRootCmd()

		_, err := executeCommand(
			root,
			"cluster",
			"create",
			"--name", "test",
			"--region", "eu-central",
			"--credentials", "aws",
			"--vpn", "wireguard_cross_location_mesh",
			"--configuration", "ha",
		)
		require.NoError(t, err)
	})

	t.Run("cluster create from nodes", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(
			root,
			"cluster",
			"create",
			"--name", "test",
			"--region", "eu-central",
			"--credentials", "gcp",
			"--vpn", "wireguard_cross_location_mesh",
			"--node", `"aws,master,medium"`,
			"--node", `"aws,worker,small"`,
			"--wait",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("cluster get by cluster name", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(
			root,
			"cluster", "get", "test-cluster-1",
		)
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` ID  NAME            STATUS  CLOUDS  REGION                       AGE      
 c1  test-cluster-1  ready   aws      Europe Central (Frankfurt)  just now`
		require.Equal(t, expected+" \n", out)
	})

	t.Run("cluster get-kubeconfig", func(t *testing.T) {
		root := newTestRootCmd()

		configPath := path.Join(t.TempDir(), "config")
		out, err := executeCommand(
			root,
			"cluster", "get-kubeconfig", "test-cluster-1", "--path", configPath,
		)
		require.NoError(t, err)
		fmt.Println(out)
		_, err = os.Stat(configPath)
		require.NoError(t, err)
	})

	t.Run("cluster delete", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(
			root,
			"cluster", "delete", "test-cluster-1", "-y",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("node list", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(
			root,
			"node", "list", "-c", "test-cluster-1",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("node ssh", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(
			root,
			"node", "ssh", "node1", "-c", "test-cluster-1",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("region list", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(root, "region", "list")
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` NAME        DISPLAYNAME                 CLOUDS           
 eu-central  Europe Central (Frankfurt)  aws gcp azure do`
		require.Equal(t, expected+" \n", out)
	})

	t.Run("credentials list", func(t *testing.T) {
		root := newTestRootCmd()

		out, err := executeCommand(root, "credentials", "list")
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` ID     NAME   CLOUD  CLUSTERS 
 cred1  aws    aws           0 
 cred2  gcp    gcp           0 
 cred3  azure  azure         0`
		require.Equal(t, expected+" \n", out)
	})
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

type mockTerminal struct {
}

func (m *mockTerminal) Connect(ctx context.Context, cfg ssh.ConnectConfig) error {
	return nil
}

type mockIpify struct {
}

func (m *mockIpify) GetPublicIP(ctx context.Context) (string, error) {
	return "1.1.1.1", nil
}
