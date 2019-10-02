// Read/Write the ipmaps configurations as a JSON file.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

const (
	configFilePermission = 0644 // Unix permission bits
	defaultConfigFile    = programName + "_config.json"
	defaultOutputFile    = "subscribers.csv"
	defaultLogFile       = programName + "_log"
	defaulIPCliCommand   = "/usr/local/bin/ipcmd"
	defaultWorkDir       = ""
)

var (
	// Conf global var tha holds program configuration
	Conf Config
)

// Config structure tha holds application configuration
type Config struct {
	REMARK string `json:"_remark_"` // JSON files dont have comments...
	//IPCli
	IPCliCmd     string `json:"ipcli_cmd"`
	IPCliUser    string `json:"ipcli_user"`
	IPCliPass    string `json:"ipcli_pass"`
	IPCliCluster string `json:"ipcli_cluster"`
	//LDAP
	LDAPHost         string `json:"ldap_host"`
	LDAPPort         uint16 `json:"ldap_port"`
	LDAPBase         string `json:"ldap_base"`
	LDAPBindUser     string `json:"ldap_user"`
	LDAPBindPassword string `json:"ldap_password"`
	// Div
	Timeout time.Duration `json:"timeout"`
	Verbose bool          `json:"verbose"`
	//Files
	TemplateToPackageFile string `json:"template_to_package_file"`
	WorkDir               string `json:"work_dir"`
	OutputFile            string `json:"output_file"`
	LogFile               string `json:"log_file"`
	//networks Leases
	NetworkLeases []ipRange `json:"network_leases"`
}

func readConfig() error {

	f, err := os.Open(*configFile)
	if err != nil {
		return fmt.Errorf("can't open config file, %s", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&Conf); err != nil {
		return fmt.Errorf("unable to decode JSON file, %v", err)
	}

	return nil

}

func createConfigFile() {
	fmt.Println("Creating config file.")

	// Bogus ips, name, passwords, etc...
	cfg := &Config{
		REMARK: "Configuration file for _ipmaps , DO NOT EDIT!!!",
		//IPCLI
		IPCliCmd:     defaulIPCliCommand,
		IPCliUser:    "xxx",
		IPCliPass:    "xxx",
		IPCliCluster: "10.1.1.1",
		//LDAP
		LDAPHost:         "192.168.0.1",
		LDAPPort:         389,
		LDAPBase:         "dc=xxx,dc=xx",
		LDAPBindUser:     "cn=xxx,dc=xxx,dc=xxx",
		LDAPBindPassword: "xxxxx",
		// Div
		Timeout: 10 * time.Second,
		Verbose: false,
		//Files
		TemplateToPackageFile: "package_maps.csv",
		WorkDir:               defaultWorkDir,
		OutputFile:            defaultOutputFile,
		LogFile:               defaultLogFile,
		// Network Leases
		NetworkLeases: []ipRange{
			ipRange{Start: net.IPv4(8, 0, 0, 0), End: net.IPv4(8, 0, 0, 255)}, // 3K
			ipRange{Start: net.IPv4(7, 0, 0, 0), End: net.IPv4(7, 0, 0, 255)}, // 8K
			ipRange{Start: net.IPv4(7, 0, 1, 0), End: net.IPv4(7, 0, 1, 255)}, // 8K
			ipRange{Start: net.IPv4(1, 0, 0, 0), End: net.IPv4(1, 0, 0, 255)}, // 8K
		},
	}

	cfgJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatal("Error encoding to JSON, ", err)
	}
	err = ioutil.WriteFile(defaultConfigFile, cfgJSON, configFilePermission)
	if err != nil {
		log.Fatal("Error creating config file, ", err)
	}
	fmt.Println("done.")

}

// For generation the first time ipmap config.file
// uncomment func main below, then run "go run config.go cpe_leases.go"
/*
func main() {
	createConfigFile()

	err := readConfig()
	if err != nil {
		log.Printf("Error loading config file: %v", err)
	}
	//fmt.Printf("config:\n%v\n", Conf)

}
*/
