package cmd

import (
	"fmt"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	dockercli "github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"github.com/kardianos/service"
	serviceimpl "github.com/workbenchapp/worknet/daoctl/cmd/service"
	"github.com/workbenchapp/worknet/daoctl/lib/options"
)

type ExposeCmd struct {
	Port     int    `help:"Port to expose"`
	Protocol string `help:"Protocol type to expose (tcp, http)" default:"tcp"`
	Driver   string `help:"Expose all local ports from Docker or Kubernetes instead of a specific port" default:"none"`
	Name     string `help:"Name of the service"`

	// TODO: Might be nice to time limit the exposure, e.g., to 24 hours.
	// "Hey, can you take a look at this?" to your teammate or a support tech
	// type of use case.
	TTL uint64 `default:"24" help:"Time limit for how long the port(s) should be exposed in hours"`
}

func (r *ExposeCmd) Run(gOpts *options.GlobalOptions) error {
	ctx := gOpts.Ctx
	log := logr.FromContextOrDiscard(ctx)

	if !serviceimpl.Admin() {
		return fmt.Errorf("exposing ports requires Admin (%s)", serviceimpl.HelpAdmin())
	}

	if r.Name == "" {
		return fmt.Errorf("--name is required")
	}

	agentConfig, err := options.Config()
	if err != nil {
		return fmt.Errorf("couldn't load/create proxy config: %s", err)
	}

	activeNet, err := agentConfig.Active()
	if err != nil {
		return fmt.Errorf("couldn't load active network: %s", err)
	}

	if r.Port == 0 && r.Driver == "none" {
		return fmt.Errorf("either --port or --driver must be specified")
	}

	validProtocol := false
	for _, exposeType := range []string{"tcp", "http"} {
		if r.Protocol == exposeType {
			validProtocol = true
		}
	}
	if !validProtocol {
		return fmt.Errorf("invalid protocol type: %q (supported: tcp, http)", r.Protocol)
	}

	if r.Port != 0 {
		log.Info("Generating config to expose port", "port", r.Port, "protocol", r.Protocol)

		// check if the port is already exposed
		for _, port := range activeNet.Ports {
			if port.PublishedPort == r.Port {
				return fmt.Errorf("port %d is already exposed", r.Port)
			}
		}

		activeNet.Ports = append(activeNet.Ports, options.Publisher{
			TargetPort:    r.Port,
			PublishedPort: r.Port,
			URL:           "0.0.0.0",
			Protocol:      r.Protocol,
			Name:          r.Name,
		})
	} else {
		validDriver := false
		for _, driverType := range []string{"docker", "kubernetes"} {
			if driverType == r.Driver {
				validDriver = true
			}
		}
		if !validDriver {
			return fmt.Errorf("invalid driver type: %q (supported: docker, kubernetes)", r.Driver)
		}

		switch r.Driver {
		case "docker":
			docker, err := dockercli.NewClientWithOpts(dockercli.FromEnv)
			if err != nil {
				return fmt.Errorf("unable to create docker client: %s", err)
			}

			containers, err := docker.ContainerList(
				ctx,
				dockertypes.ContainerListOptions{},
			)
			if err != nil {
				return fmt.Errorf("couldn't list containers: %s", err)
			}

			for _, container := range containers {
				for _, port := range container.Ports {
					if port.PublicPort == 0 || port.Type != "tcp" {
						continue
					}

					activeNet.Ports = append(activeNet.Ports, options.Publisher{
						TargetPort:    int(port.PublicPort),
						PublishedPort: int(port.PublicPort),
						// remove leading slash
						Name:     "docker-" + strings.Replace(container.Names[0][1:], "_", "-", -1),
						URL:      "0.0.0.0",
						Protocol: "tcp",
					})
				}
			}
		case "kubernetes":
			return fmt.Errorf("kubernetes driver not yet implemented")
		}
	}

	agentConfig.Worknets[agentConfig.ActiveNet] = activeNet

	if err := options.WriteAgentConfig(agentConfig); err != nil {
		return fmt.Errorf("couldn't write proxy config: %s", err)
	}

	log.Info("Restarting service to apply changes")

	// TODO: "restart" seemed to cause a bug for me on OSX at one point where
	// `daoctl restart`` never exited.
	//
	// TODO2: start and stop seem to also have this issue, so this might be
	// just as risky. For now, workaround is to spam kill -9 as needed.
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
