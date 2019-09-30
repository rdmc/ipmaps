package main

import (
	"fmt"
	"log"

	"github.com/rdmc/mac"

	ldap "gopkg.in/ldap.v3"

	"strconv"
)

const (
	searchBase        = "ou=cms,o=incognito,dc=cabotva,dc=net"
	filter            = "(objectClass=modemClass)"
	pageSize   uint32 = 1024 * 16 // 16K best results determined by experiments
)

// CMTemplateMap - CM mac -> Ldap template  mapping
type CMTemplateMap map[string]int

var (
	attributes = []string{"cn", "accessType"}
)

// is "getLDAPCMTemplates()" a better name then ?"getCMTemplates()" ???
func getCMTemplates() (CMTemplateMap, error) {
	var mt CMTemplateMap
	mt = make(map[string]int)

	// connect
	fmt.Print(" Connect..")

	// ldap.DefaultTimeout package-level variavel
	ldap.DefaultTimeout = Conf.Timeout

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", Conf.LDAPHost, Conf.LDAPPort))
	if err != nil {
		return nil, fmt.Errorf("getMACTemplates Dial error: %s", err)
	}
	defer l.Close()

	fmt.Print(" Bind..")

	l.SetTimeout(Conf.Timeout)
	// bind
	err = l.Bind(Conf.LDAPBindUser, Conf.LDAPBindPassword)
	if err != nil {
		return nil, fmt.Errorf("getMACTemplates Bind error: %s", err)
	}

	fmt.Print(" Search..")

	// search
	pagingControl := ldap.NewControlPaging(pageSize)
	controls := []ldap.Control{pagingControl}

	//packagesCnt := make(map[int]int)

	for {
		request := ldap.NewSearchRequest(searchBase, ldap.ScopeSingleLevel, ldap.DerefAlways,
			0, 0, false, filter, attributes, controls)
		sr, err := l.Search(request)
		if err != nil {
			return nil, fmt.Errorf("getMACTemplates Search error: %s", err)
		}

		for _, entry := range sr.Entries {

			m, err := mac.ParseMAC(entry.GetAttributeValue("cn"))
			if err != nil {
				log.Printf("getMACTemplates: Invalid MAC %q, %v", entry.GetAttributeValue("cn"), err)
				continue
			}

			p, err := strconv.Atoi(entry.GetAttributeValue("accessType"))
			if err != nil {
				log.Printf("getMACTemplates: Invalid package %q, %v", entry.GetAttributeValue("cn"), err)
				continue
			}

			if _, ok := mt[m.PlainString()]; ok {
				log.Printf("getMACTemplates: Duplicated MAC %q", m.PlainString())
				continue
			}
			mt[m.PlainString()] = p
			//packagesCnt[p]++

		}

		// loop guard for ldap search with pagging
		updatedControl := ldap.FindControl(sr.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pagingControl.SetCookie(ctrl.Cookie)
			continue // "for" loop
		}
		break // exit "for" loop
	}

	return mt, nil
}
