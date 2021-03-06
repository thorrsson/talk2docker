package commands

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	api "github.com/yungsang/dockerclient"
	"github.com/yungsang/tablewriter"
	"github.com/yungsang/talk2docker/client"
)

var (
	boolAll, boolLatest, boolQuiet, boolSize, boolNoHeader bool
)

var cmdPs = &cobra.Command{
	Use:   "ps",
	Short: "List containers",
	Long:  appName + " ps - List containers",
}

func init() {
	cmdPs.Flags().BoolVarP(&boolAll, "all", "a", false, "Show all containers. Only running containers are shown by default.")
	cmdPs.Flags().BoolVarP(&boolLatest, "latest", "l", false, "Show only the latest created container, include non-running ones.")
	cmdPs.Flags().BoolVarP(&boolQuiet, "quiet", "q", false, "Only display numeric IDs")
	cmdPs.Flags().BoolVarP(&boolSize, "size", "s", false, "Display sizes")
	cmdPs.Flags().BoolVarP(&boolNoHeader, "no-header", "n", false, "Omit the header")
	cmdPs.Run = listContainers
}

func listContainers(ctx *cobra.Command, args []string) {
	docker, err := client.GetDockerClient(configPath, hostName)
	if err != nil {
		log.Fatal(err)
	}

	filters := ""
	if boolLatest {
		filters += "&limit=1"
	}
	if boolSize {
		filters += "&size=1"
	}

	containers, err := docker.ListContainers(boolAll, boolSize, filters)
	if err != nil {
		log.Fatal(err)
	}

	if boolQuiet {
		for _, container := range containers {
			fmt.Println(Truncate(container.Id, 12))
		}
		return
	}

	trimNamePrefix := func(ss []string) []string {
		for i, s := range ss {
			ss[i] = strings.TrimPrefix(s, "/")
		}
		return ss
	}

	formatPorts := func(ports []api.Port) string {
		result := []string{}
		for _, p := range ports {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s",
				p.IP, p.PublicPort, p.PrivatePort, p.Type))
		}
		return strings.Join(result, ", ")
	}

	var items [][]string
	for _, container := range containers {
		out := []string{
			Truncate(container.Id, 12),
			strings.Join(trimNamePrefix(container.Names), ", "),
			container.Image,
			Truncate(container.Command, 20),
			FormatDateTime(time.Unix(container.Created, 0)),
			container.Status,
			formatPorts(container.Ports),
		}
		if boolSize {
			out = append(out, FormatFloat(float64(container.SizeRw)/1000000))
		}
		items = append(items, out)
	}

	header := []string{
		"ID",
		"Names",
		"Image",
		"Command",
		"Created at",
		"Status",
		"Ports",
	}
	if boolSize {
		header = append(header, "Size(MB)")
	}

	table := tablewriter.NewWriter(os.Stdout)
	if !boolNoHeader {
		table.SetHeader(header)
	} else {
		table.SetBorder(false)
	}
	table.AppendBulk(items)
	table.Render()
}
