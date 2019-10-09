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
	// Program name
	programName = "ipmaps"
	// Program Version
	programVersion = "0.1.0" // Major.Minor.Patch
	// Program header
	programHeader = programName + " " + programVersion +
		". Generates subscriber/ip mappings for the Cisco SCA BB, SM p3subsdb."
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

	fmt.Println(programHeader)

	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("usage: " + programName + " [-v][-cf filename][-o filename]")
		flag.PrintDefaults()
		log.Fatal("Bye, bye.")
		// PROGRAM TERMINATE
	}

	log.SetFlags(0)
	log.SetPrefix(programName + ": ")

	verbosePrintf("readConfig: ")
	//fmt.Print("readConfig: ")
	err = readConfig()
	if err != nil {
		log.Fatal(err)
	}

	verbosePrintf("ok\nreadTemplateToPackage: ")
	//fmt.Print("ok\nreadTemplateToPackage: ")
	TemplatePackage, err = readTemplateToPackage()
	if err != nil {
		log.Fatal(err)
	}

	verbosePrintf("ok\ngetMACTemplates: ")
	//fmt.Print("ok\ngetMACTemplates: ")
	CMTemplate, err = getLDAPCMTemplates()
	if err != nil {
		log.Fatal(err)
	}

	verbosePrintf("ok\ngetCMCPELeases: ")
	//fmt.Print("ok\ngetCMCPELeases: ")
	CMCPE, err = getCMCPELeases()
	if err != nil {
		log.Fatal(err)
	}

	ofName := cfg.OutputFile
	if *outputFile != "" {
		ofName = *outputFile
	}

	verbosePrintf("ok\nwrite output file %q:", ofName)
	//fmt.Printf("ok\nwrite output file %q:", ofName)
	of, err := os.Create(ofName)
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
	verbosePrintf("ok\n")
	//fmt.Println("ok")

	runtime := time.Since(startTime)

	// Write stats
	fmt.Printf("stats: TemplatePackages=%d, CMTemplate=%d, CMCPE=%d, noPackTmpl=%d, run in %f secs.\n",
		len(TemplatePackage), len(CMTemplate), len(CMCPE), noPackTmpl, runtime.Seconds())

	// bye, bye
	fmt.Println("That's all Folks!!")
}

/* logging to a file

 	f, err := os.OpenFile("app.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
        if err != nil {
                log.Fatal(err)
        }
        defer f.Close()
        log.SetOutput(f)
	log.Println("Application started")
*/

func verbosePrintf(fmts string, args ...interface{}) {
	if *verbose { // TODO: Config.Verbose if "-v" not set
		//programCounter, file, line, _ := runtime.Caller(1)
		//fn := runtime.FuncForPC(programCounter)
		//prefix := fmt.Sprintf("[%s:%s %d] %s", file, fn.Name(), line, fmts)
		fmt.Printf(fmts, args...)
	}
}
