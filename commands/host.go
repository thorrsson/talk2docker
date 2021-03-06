package commands

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/gopass"
	"github.com/spf13/cobra"
	api "github.com/yungsang/dockerclient"
	"github.com/yungsang/tablewriter"
	"github.com/yungsang/talk2docker/client"
)

var cmdHost = &cobra.Command{
	Use:   "host [command]",
	Short: "Manage hosts",
	Long:  appName + " host - Manage hosts",
	Run: func(ctx *cobra.Command, args []string) {
		ctx.Usage()
	},
}

var cmdListHosts = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List hosts",
	Long:    appName + " host list - List hosts",
	Run:     listHosts,
}

var cmdSwitchHost = &cobra.Command{
	Use:     "switch <NAME>",
	Aliases: []string{"sw"},
	Short:   "Switch the default host",
	Long:    appName + " host switch - Switch the default host",
	Run:     switchHost,
}

var cmdGetHostInfo = &cobra.Command{
	Use:   "info [NAME]",
	Short: "Get the host information",
	Long:  appName + " host info - Get the host information",
	Run:   getHostInfo,
}

var cmdLogin = &cobra.Command{
	Use:   "login [NAME]",
	Short: "Log in to a Docker registry server through the host",
	Long:  appName + " host login - Log in to a Docker registry server through the host",
	Run:   login,
}

var cmdLogout = &cobra.Command{
	Use:   "logout [NAME]",
	Short: "Log out from a Docker registry server through the host",
	Long:  appName + " host logout - Log out from a Docker registry server through the host",
	Run:   logout,
}

var cmdAddHost = &cobra.Command{
	Use:   "add <NAME> <URL> [DESCRIPTION]",
	Short: "Add a new host into the config file",
	Long:  appName + " host add - Add a new host into the config",
	Run:   addHost,
}

var cmdRemoveHost = &cobra.Command{
	Use:     "remove <NAME>",
	Aliases: []string{"rm", "delete", "del"},
	Short:   "Remove a host from the config file",
	Long:    appName + " host remove - Remove a host from the config file",
	Run:     removeHost,
}

var cmdHosts = &cobra.Command{
	Use:   "hosts",
	Short: "Shortcut to list hosts",
	Long:  appName + " hosts - List hosts",
	Run:   listHosts,
}

func init() {
	cmdListHosts.Flags().BoolVarP(&boolQuiet, "quiet", "q", false, "Only display numeric IDs")
	cmdListHosts.Flags().BoolVarP(&boolNoHeader, "no-header", "n", false, "Omit the header")

	cmdHosts.Flags().BoolVarP(&boolQuiet, "quiet", "q", false, "Only display host names")
	cmdHosts.Flags().BoolVarP(&boolNoHeader, "no-header", "n", false, "Omit the header")

	cmdGetHostInfo.Flags().BoolVarP(&boolNoHeader, "no-header", "n", false, "Omit the header")

	cmdHost.AddCommand(cmdListHosts)
	cmdHost.AddCommand(cmdSwitchHost)
	cmdHost.AddCommand(cmdGetHostInfo)
	cmdHost.AddCommand(cmdLogin)
	cmdHost.AddCommand(cmdLogout)
	cmdHost.AddCommand(cmdAddHost)
	cmdHost.AddCommand(cmdRemoveHost)
}

func listHosts(ctx *cobra.Command, args []string) {
	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if boolQuiet {
		for _, host := range config.Hosts {
			fmt.Println(host.Name)
		}
		return
	}

	var items [][]string
	for _, host := range config.Hosts {
		out := []string{
			FormatBool(host.Name == config.Default, "*", ""),
			host.Name,
			host.URL,
			FormatNonBreakingString(host.Description),
			FormatBool(host.TLS, "YES", ""),
		}
		items = append(items, out)
	}

	header := []string{
		"",
		"Name",
		"URL",
		"Description",
		"TLS",
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

func switchHost(ctx *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Needs an argument <NAME> to switch")
		ctx.Usage()
		return
	}

	name := args[0]

	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	host, err := config.GetHost(name)
	if err != nil {
		log.Fatal(err)
	}

	config.Default = host.Name

	err = config.SaveConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	listHosts(ctx, args)
}

func getHostInfo(ctx *cobra.Command, args []string) {
	if len(args) > 0 {
		hostName = args[0]
	}

	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	docker, err := client.GetDockerClient(configPath, hostName)
	if err != nil {
		log.Fatal(err)
	}

	info, err := docker.Info()
	if err != nil {
		log.Fatal(err)
	}

	var items [][]string

	host, err := config.GetHost(hostName)
	if err != nil {
		log.Fatal(err)
	}

	items = append(items, []string{
		"Host", host.Name,
	})
	items = append(items, []string{
		"URL", host.URL,
	})
	items = append(items, []string{
		"Description", FormatNonBreakingString(host.Description),
	})
	items = append(items, []string{
		"TLS", FormatBool(host.TLS, "Supported", "No"),
	})
	if host.TLS {
		items = append(items, []string{
			FormatNonBreakingString("  CA Certificate file"), FormatNonBreakingString(host.TLSCaCert),
		})
		items = append(items, []string{
			FormatNonBreakingString("  Certificate file"), FormatNonBreakingString(host.TLSCert),
		})
		items = append(items, []string{
			FormatNonBreakingString("  Key file"), FormatNonBreakingString(host.TLSKey),
		})
		items = append(items, []string{
			FormatNonBreakingString("  Verify"), FormatBool(host.TLSVerify, "Required", "No"),
		})
	}

	items = append(items, []string{
		"Containers", strconv.Itoa(info.Containers),
	})
	items = append(items, []string{
		"Images", strconv.Itoa(info.Images),
	})
	items = append(items, []string{
		"Storage Driver", info.Driver,
	})
	for _, pair := range info.DriverStatus {
		items = append(items, []string{
			FormatNonBreakingString(fmt.Sprintf("  %s", pair[0])), FormatNonBreakingString(fmt.Sprintf("%s", pair[1])),
		})
	}
	items = append(items, []string{
		"Execution Driver", info.ExecutionDriver,
	})
	items = append(items, []string{
		"Kernel Version", info.KernelVersion,
	})
	items = append(items, []string{
		"Operating System", FormatNonBreakingString(info.OperatingSystem),
	})
	items = append(items, []string{
		"CPUs", strconv.Itoa(info.NCPU),
	})
	items = append(items, []string{
		"Total Memory", fmt.Sprintf("%s GB", FormatFloat(float64(info.MemTotal)/1000000000)),
	})

	items = append(items, []string{
		"Index Server Address", info.IndexServerAddress,
	})

	items = append(items, []string{
		"Memory Limit", FormatBool(info.MemoryLimit != 0, "Supported", "No"),
	})
	items = append(items, []string{
		"Swap Limit", FormatBool(info.SwapLimit != 0, "Supported", "No"),
	})
	items = append(items, []string{
		"IPv4 Forwarding", FormatBool(info.IPv4Forwarding != 0, "Enabled", "Disabled"),
	})

	items = append(items, []string{
		"ID", info.ID,
	})
	items = append(items, []string{
		"Name", info.Name,
	})
	items = append(items, []string{
		"Labels", FormatNonBreakingString(strings.Join(info.Labels, ", ")),
	})

	items = append(items, []string{
		"Debug Mode", FormatBool(info.Debug != 0, "Yes", "No"),
	})
	if info.Debug != 0 {
		items = append(items, []string{
			FormatNonBreakingString("  Events Listeners"), strconv.Itoa(info.NEventsListener),
		})
		items = append(items, []string{
			FormatNonBreakingString("  Fds"), strconv.Itoa(info.NFd),
		})
		items = append(items, []string{
			FormatNonBreakingString("  Goroutines"), strconv.Itoa(info.NGoroutines),
		})

		items = append(items, []string{
			FormatNonBreakingString("  Init Path"), info.InitPath,
		})
		items = append(items, []string{
			FormatNonBreakingString("  Init SHA1"), info.InitSha1,
		})
		items = append(items, []string{
			FormatNonBreakingString("  Docker Root Dir"), info.DockerRootDir,
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	if boolNoHeader {
		table.SetBorder(false)
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(items)
	table.Render()
}

func login(ctx *cobra.Command, args []string) {
	if len(args) > 0 {
		hostName = args[0]
	}

	docker, err := client.GetDockerClient(configPath, hostName)
	if err != nil {
		log.Fatal(err)
	}

	info, err := docker.Info()
	if err != nil {
		log.Fatal(err)
	}

	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	server, _ := config.GetIndexServer(info.IndexServerAddress)

	authConfig := api.AuthConfig{
		ServerAddress: server.URL,
	}

	promptDefault := func(prompt string, configDefault string) {
		if configDefault == "" {
			fmt.Printf("%s: ", prompt)
		} else {
			fmt.Printf("%s (%s): ", prompt, configDefault)
		}
	}

	readInput := func() string {
		reader := bufio.NewReader(os.Stdin)
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		return string(line)
	}

	promptDefault("Username", server.Username)
	authConfig.Username = readInput()
	if authConfig.Username == "" {
		authConfig.Username = server.Username
	}

	authConfig.Password, err = gopass.GetPass("Password: ")

	promptDefault("Email", server.Email)
	authConfig.Email = readInput()
	if authConfig.Email == "" {
		authConfig.Email = server.Email
	}

	err = docker.Auth(&authConfig)
	if err != nil {
		log.Fatal(err)
	}

	server.Username = authConfig.Username
	server.Email = authConfig.Email
	server.Auth = server.Encode(authConfig.Username, authConfig.Password)

	config.SetIndexServer(server)

	err = config.SaveConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Login Succeeded!")
}

func logout(ctx *cobra.Command, args []string) {
	if len(args) > 0 {
		hostName = args[0]
	}

	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	docker, err := client.GetDockerClient(configPath, hostName)
	if err != nil {
		log.Fatal(err)
	}

	info, err := docker.Info()
	if err != nil {
		log.Fatal(err)
	}

	server, notFound := config.GetIndexServer(info.IndexServerAddress)
	if (notFound != nil) || (server.Auth == "") {
		log.Fatal("Not logged in")
	}

	config.LogoutIndexServer(info.IndexServerAddress)

	err = config.SaveConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Logout Succeeded!")
}

func addHost(ctx *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("Needs two arguments <NAME> and <URL> at least")
		ctx.Usage()
		return
	}

	var (
		name = args[0]
		url  = args[1]
		desc = ""
	)

	if len(args) > 2 {
		desc = strings.Join(args[2:], " ")
	}

	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = config.GetHost(name)
	if err == nil {
		log.Fatal(fmt.Sprintf("\"%s\" already exists", name))
	}

	newHost := client.Host{
		Name:        name,
		URL:         url,
		Description: desc,
	}

	config.Default = newHost.Name
	config.Hosts = append(config.Hosts, newHost)

	err = config.SaveConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	listHosts(ctx, args)
}

func removeHost(ctx *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Needs an argument <NAME> to remove")
		ctx.Usage()
		return
	}

	name := args[0]

	config, err := client.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if config.Default == name {
		log.Fatal("You can't remove the default host.")
	}

	_, err = config.GetHost(name)
	if err != nil {
		log.Fatal(err)
	}

	hosts := []client.Host{}

	for _, host := range config.Hosts {
		if host.Name != name {
			hosts = append(hosts, host)
		}
	}

	config.Hosts = hosts

	err = config.SaveConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	listHosts(ctx, args)
}
