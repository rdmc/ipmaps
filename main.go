package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	//SCESubsriberHeader Header for  SCAS BB sm p3subsdb subscriber file
	SCESubsriberHeader = "# SCE Subscribers @%s\n# CSV line format: subscriber-id, domain, mappings, package-id\n"
	// SCESubscribersLine = "CMMAC,,CPEIPS, PACkAGE"
	SCESubscribersLine = "%s,,%s,%d\n"
)

var (
	// flags
	configFile = flag.String("cf", defaultConfigFile, "json configuration file")
	outputFile = flag.String("o", "", "Output file name")
	verbose    = flag.Bool("v", false, "verbose")

	// TemplatePackage global var, LDAP template to SCE Package Mappings
	TemplatePackage TemplatePackageMap

	// CMTemplate global var, holds template for each CM MAC from LDAP
	CMTemplate CMTemplateMap

	//CMCPE holds  CPE ip address for each CM
	CMCPE CMCPEMap
)

func main() {
	var err error
	startTime := time.Now()

	fmt.Println("ipmaps. Generates subscriber/ip mappings for the Cisco SCA BB, SM p3subsdb.")

	log.SetFlags(0)
	log.SetPrefix("ipmaps: ")

	fmt.Print("readConfig: ")
	err = readConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("ok\nreadTemplateToPackage: ")
	TemplatePackage, err = readTemplateToPackage()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("ok\ngetMACTemplates: ")
	CMTemplate, err = getCMTemplates()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("ok\ngetCMCPELeases: ")
	CMCPE, err = getCMCPELeases()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ok\nwrite output file %q:", Conf.OutputFile)
	of, err := os.Create(Conf.OutputFile)
	if err != nil {
		panic(err)
	}
	defer of.Close()
	w := bufio.NewWriter(of)

	// subscriber file header
	fmt.Fprintf(w, SCESubsriberHeader, time.Now().Format(time.RFC3339))

	// subscriber file lines
	noPackTmpl := 0
	for cm, cpes := range CMCPE {
		template, ok := CMTemplate[cm]
		if !ok { // This should NEVER hapen
			log.Printf("WTF: cm mac %q not found in ldap\n", cm)
		}
		pack, ok := TemplatePackage[template]
		if !ok {
			log.Printf("WARNING: cm mac %q no package for template %d\n", cm, template)
			pack = TemplatePackage[0] // 0 = default
			noPackTmpl++
		}

		fmt.Fprintf(w, SCESubscribersLine, cm, cpes, pack)
	}
	w.Flush()
	fmt.Println("ok")

	runtime := time.Since(startTime)

	// Write stats
	fmt.Printf("stats: TemplatePackages=%d, CMTemplate=%d, CMCPE=%d, noPackTmpl=%d, run in %f secs.\n",
		len(TemplatePackage), len(CMTemplate), len(CMCPE), noPackTmpl, runtime.Seconds())

	// bye, bye
	fmt.Println("That's all Folks!!")
}
