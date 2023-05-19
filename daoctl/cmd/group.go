package cmd

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path"
	"sort"
	"text/tabwriter"

	gagliardetto "github.com/gagliardetto/solana-go"
	"github.com/go-logr/logr"
	"github.com/kardianos/service"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
	"github.com/workbenchapp/worknet/daoctl/lib/solana"
)

type GroupListCmd struct{}

type GroupSelectCmd struct {
	Name string `arg:"default" help:"name of group to select"`
}

type GroupJoinCmd struct {
	Name string `arg:"" cmd:"" default:"default" help:"The work group name"`
}

type GroupCmd struct {
	Join   GroupJoinCmd   `cmd:"" help:"Create a device key to join a new work group"`
	List   GroupListCmd   `cmd:"" help:"List smart wallet owners" default:"1"`
	Select GroupSelectCmd `cmd:"" help:"Select a group as the active group"`
}

type SortedGroup struct {
	*options.WorknetConfig
	Active, Name, FormattedPorts string
}

func (r *GroupListCmd) Run(gOpts *options.GlobalOptions) error {
	// settings mostly cribbed from docker
	tw := tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0)
	fmt.Fprintln(tw, "ACTIVE\tNAME\tKEYFILE\tPORTS")

	defer func() {
		tw.Flush()
	}()

	agentConfig, err := options.Config()
	if err != nil {
		return err
	}

	sortedGroups := []SortedGroup{}

	for groupName, group := range agentConfig.Worknets {
		activeMark := ""
		if groupName == agentConfig.ActiveNet {
			activeMark = "*"
		}
		portFmt := ""
		for _, port := range group.Ports {
			portFmt += fmt.Sprintf("%s:%d/%s ", port.Name, port.TargetPort, port.Protocol)
		}
		sortedGroups = append(sortedGroups, SortedGroup{
			Active:         activeMark,
			Name:           groupName,
			FormattedPorts: portFmt,
			WorknetConfig:  group,
		})
	}

	sort.Slice(sortedGroups, func(i, j int) bool {
		return sortedGroups[i].Name < sortedGroups[j].Name
	})

	for _, group := range sortedGroups {
		fmt.Fprintln(tw, group.Active+"\t"+group.Name+"\t"+group.KeyFile+"\t"+group.FormattedPorts)
	}

	return nil
}

func (r *GroupSelectCmd) Run(gOpts *options.GlobalOptions) error {
	log := logr.FromContextOrDiscard(gOpts.Ctx)

	agentConfig, err := options.Config()
	if err != nil {
		return err
	}

	if _, ok := agentConfig.Worknets[r.Name]; !ok {
		return fmt.Errorf("cannot select group %q, it is not present in config", r.Name)
	}

	agentConfig.ActiveNet = r.Name

	if err := options.WriteAgentConfig(agentConfig); err != nil {
		return err
	}

	log.Info("Restarting service to apply changes")

	s, err := common()
	if err != nil {
		return err
	}
	if err := service.Control(s, "stop"); err != nil {
		return fmt.Errorf("stop service failed: %s", err)
	}
	if err := service.Control(s, "start"); err != nil {
		return fmt.Errorf("start service failed: %s", err)
	}

	return nil
}

func (r *GroupJoinCmd) Run(gOpts *options.GlobalOptions) error {
	agentConfig, err := options.Config()
	if err != nil {
		return err
	}

	keyFile := fmt.Sprintf("%s-key.json", r.Name)
	key := gagliardetto.NewWallet()
	configDir, err := options.GetConfigDir("WorkNet")
	if err != nil {
		return err
	}

	if err := solana.WritePrivateKeyAsJSON(
		path.Join(configDir, keyFile),
		ed25519.PrivateKey(key.PrivateKey),
	); err != nil {
		return err
	}

	agentConfig.Worknets[r.Name] = &options.WorknetConfig{
		KeyFile: keyFile,
		Ports:   []options.Publisher{},
	}

	fmt.Printf("Config for new workgroup %q created, add the following device key to the work group:\n", r.Name)
	fmt.Println(key.PublicKey().String())

	return options.WriteAgentConfig(agentConfig)
}
