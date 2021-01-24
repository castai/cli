package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/castai/cast-cli/pkg/client"
)

func TestCommands(t *testing.T) {
	api := client.NewMock()
	root := NewRootCmd(logrus.New(), api)

	t.Run("cluster list", func(t *testing.T) {
		out, err := executeCommand(root, "cluster", "list")
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` ID  NAME            STATUS  CLOUDS  REGION                      
 c1  test-cluster-1  ready   aws      Europe Central (Frankfurt)`
		require.Equal(t, expected+" \n", out)
	})

	t.Run("cluster create from configuration", func(t *testing.T) {
		out, err := executeCommand(
			root,
			"cluster",
			"create",
			"--name", "test",
			"--region", "eu-central",
			"--credentials", "cred1",
			"--vpn", "wiregaurd_cross_location_mesh",
			"--configuration", "ha",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("cluster create from nodes", func(t *testing.T) {
		out, err := executeCommand(
			root,
			"cluster",
			"create",
			"--name", "test",
			"--region", "eu-central",
			"--credentials", "cred1",
			"--vpn", "wiregaurd_cross_location_mesh",
			"--node", `"aws,master,medium"`,
			"--node", `"aws,worker,small"`,
			"--wait", "--progress",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("cluster get by cluster name", func(t *testing.T) {
		out, err := executeCommand(
			root,
			"cluster", "get", "test-cluster-1",
		)
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` ID  NAME            STATUS  CLOUDS  REGION                      
 c1  test-cluster-1  ready   aws      Europe Central (Frankfurt)`
		require.Equal(t, expected+" \n", out)
	})

	t.Run("cluster get-kubeconfig", func(t *testing.T) {
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
		out, err := executeCommand(
			root,
			"cluster", "delete", "test-cluster-1",
		)
		require.NoError(t, err)
		fmt.Println(out)
	})

	t.Run("region list", func(t *testing.T) {
		out, err := executeCommand(root, "region", "list")
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` NAME        DISPLAYNAME                 CLOUDS           
 eu-central  Europe Central (Frankfurt)  aws gcp azure do`
		require.Equal(t, expected+" \n", out)
	})

	t.Run("credentials list", func(t *testing.T) {
		out, err := executeCommand(root, "credentials", "list")
		require.NoError(t, err)
		fmt.Println(out)
		expected := ` ID     NAME   CLOUD  USEDBY 
 cred1  aws    aws         0 
 cred2  gcp    gcp         0 
 cred3  azure  azure       0`
		require.Equal(t, expected+" \n", out)
	})
}

func TestPa(t *testing.T) {
	p := path.Join("/cast/config")
	fmt.Println(path.Dir(p))
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
