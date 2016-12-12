package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/urfave/cli"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

type listOptions struct {
	// id of the container
	id string
	// podID of the container
	podID string
	// state of the container
	state string
	// quiet is for listing just container IDs
	quiet bool
	// labels are selectors for the container
	labels map[string]string
}

var listFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "quiet",
		Usage: "list only container IDs",
	},
	cli.StringFlag{
		Name:  "id",
		Value: "",
		Usage: "filter by container id",
	},
	cli.StringFlag{
		Name:  "pod",
		Value: "",
		Usage: "filter by container pod id",
	},
	cli.StringFlag{
		Name:  "state",
		Value: "",
		Usage: "filter by container state",
	},
	cli.StringSliceFlag{
		Name:  "label",
		Usage: "filter by key=value label",
	},
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "list running pods",
	Flags: listFlags,
	Action: func(context *cli.Context) error {
		// Set up a connection to the server.
		conn, err := getClientConnection(context)
		if err != nil {
			return fmt.Errorf("failed to connect: %v", err)
		}
		defer conn.Close()
		client := runtime.NewRuntimeServiceClient(conn)
		opts := listOptions{
			id:     context.String("id"),
			podID:  context.String("pod"),
			state:  context.String("state"),
			quiet:  context.Bool("quiet"),
			labels: make(map[string]string),
		}

		for _, l := range context.StringSlice("label") {
			pair := strings.Split(l, "=")
			if len(pair) != 2 {
				return fmt.Errorf("incorrectly specified label: %v", l)
			}
			opts.labels[pair[0]] = pair[1]
		}

		err = ListContainers(client, opts)
		if err != nil {
			return fmt.Errorf("listing containers failed: %v", err)
		}
		return nil
	},
}

// ListContainers sends a ListContainerRequest to the server, and parses
// the returned ListContainerResponse.
func ListContainers(client runtime.RuntimeServiceClient, opts listOptions) error {
	filter := &runtime.ContainerFilter{}
	if opts.id != "" {
		filter.Id = &opts.id
	}
	if opts.podID != "" {
		filter.PodSandboxId = &opts.podID
	}
	if opts.state != "" {
		st := runtime.ContainerState_CONTAINER_UNKNOWN
		switch opts.state {
		case "created":
			st = runtime.ContainerState_CONTAINER_CREATED
			filter.State = &st
		case "running":
			st = runtime.ContainerState_CONTAINER_RUNNING
			filter.State = &st
		case "stopped":
			st = runtime.ContainerState_CONTAINER_EXITED
			filter.State = &st
		default:
			log.Fatalf("--state should be one of created, running or stopped")
		}
	}
	if opts.labels != nil {
		filter.LabelSelector = opts.labels
	}
	r, err := client.ListContainers(context.Background(), &runtime.ListContainersRequest{
		Filter: filter,
	})
	if err != nil {
		return err
	}
	for _, c := range r.GetContainers() {
		if opts.quiet {
			fmt.Println(*c.Id)
			continue
		}
		fmt.Printf("ID: %s\n", *c.Id)
		fmt.Printf("Pod: %s\n", *c.PodSandboxId)
		if c.Metadata != nil {
			if c.Metadata.Name != nil {
				fmt.Printf("Name: %s\n", *c.Metadata.Name)
			}
			if c.Metadata.Attempt != nil {
				fmt.Printf("Attempt: %v\n", *c.Metadata.Attempt)
			}
		}
		if c.State != nil {
			fmt.Printf("Status: %s\n", *c.State)
		}
		if c.CreatedAt != nil {
			ctm := time.Unix(0, *c.CreatedAt)
			fmt.Printf("Created: %v\n", ctm)
		}
		if c.Labels != nil {
			fmt.Println("Labels:")
			for _, k := range getSortedKeys(c.Labels) {
				fmt.Printf("\t%s -> %s\n", k, c.Labels[k])
			}
		}
		if c.Annotations != nil {
			fmt.Println("Annotations:")
			for _, k := range getSortedKeys(c.Annotations) {
				fmt.Printf("\t%s -> %s\n", k, c.Annotations[k])
			}
		}
		fmt.Println()
	}
	return nil
}

func getClientConnection(context *cli.Context) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(context.GlobalString("connect"), grpc.WithInsecure(), grpc.WithTimeout(context.GlobalDuration("timeout")),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return conn, nil
}

func getSortedKeys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
