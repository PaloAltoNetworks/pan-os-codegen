package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"

	"github.com/PaloAltoNetworks/pango"
	"github.com/PaloAltoNetworks/pango/device/services/dns"
	"github.com/PaloAltoNetworks/pango/device/services/ntp"
	"github.com/PaloAltoNetworks/pango/network/profiles/interface_management"
	"github.com/PaloAltoNetworks/pango/objects/address"
	"github.com/PaloAltoNetworks/pango/objects/service"
	"github.com/PaloAltoNetworks/pango/objects/tag"
	"github.com/PaloAltoNetworks/pango/policies/rules/security"
	"github.com/PaloAltoNetworks/pango/rule"
	"github.com/PaloAltoNetworks/pango/util"
)

func main() {
	var err error
	ctx := context.Background()

	// FW
	c := &pango.XmlApiClient{
		CheckEnvironment:      true,
		SkipVerifyCertificate: true,
	}
	if err = c.Setup(); err != nil {
		log.Printf("Failed to setup client: %s", err)
		return
	}
	log.Printf("Setup client %s (%s)", c.Hostname, c.Username)

	if err = c.Initialize(ctx); err != nil {
		log.Printf("Failed to initialize client: %s", err)
		return
	}

	// CHECKS
	checkInterfaceMgmtProfil(c, ctx)
	checkSecurityPolicyRules(c, ctx)
	checkSecurityPolicyRulesMove(c, ctx)
	checkTag(c, ctx)
	checkAddress(c, ctx)
	checkService(c, ctx)
	checkNtp(c, ctx)
	checkDns(c, ctx)
}

func checkInterfaceMgmtProfil(c *pango.XmlApiClient, ctx context.Context) {
	entry := interface_management.Entry{
		Name:        "codegen_mgmt_profile",
		Http:        util.Bool(true),
		Ping:        util.Bool(true),
		PermittedIp: []string{"1.1.1.1", "2.2.2.2"},
	}
	location := interface_management.Location{
		Device: &interface_management.DeviceLocation{
			NgfwDevice: "localhost.localdomain",
		},
	}
	api := interface_management.NewService(c)

	reply, err := api.Create(ctx, location, entry)
	if err != nil {
		log.Printf("Failed to create interface management profile: %s", err)
		return
	}
	log.Printf("Interface management profile %s created\n", reply.Name)
}

func checkSecurityPolicyRules(c *pango.XmlApiClient, ctx context.Context) {
	// SECURITY POLICY RULE - ADD
	securityPolicyRuleEntry := security.Entry{
		Name:               "codegen_rule",
		Description:        util.String("initial description"),
		Action:             util.String("allow"),
		SourceZone:         []string{"any"},
		SourceAddress:      []string{"any"},
		DestinationZone:    []string{"any"},
		DestinationAddress: []string{"any"},
		Application:        []string{"any"},
		Service:            []string{"application-default"},
	}

	securityPolicyRuleLocation := security.Location{
		Vsys: &security.VsysLocation{
			NgfwDevice: "localhost.localdomain",
			Rulebase:   "post-rulebase",
			Vsys:       "vsys1",
		},
	}

	securityPolicyRuleApi := security.NewService(c)
	securityPolicyRuleReply, err := securityPolicyRuleApi.Create(ctx, securityPolicyRuleLocation, securityPolicyRuleEntry)
	if err != nil {
		log.Printf("Failed to create security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s:%s' with description '%s' created", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, *securityPolicyRuleReply.Description)

	// SECURITY POLICY RULE - READ
	securityPolicyRuleReply, err = securityPolicyRuleApi.Read(ctx, securityPolicyRuleLocation, securityPolicyRuleReply.Name, "get")
	if err != nil {
		log.Printf("Failed to update security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s:%s' with description '%s' read", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, *securityPolicyRuleReply.Description)

	// SECURITY POLICY RULE - UPDATE
	securityPolicyRuleEntry.Description = util.String("changed description")
	securityPolicyRuleReply, err = securityPolicyRuleApi.Update(ctx, securityPolicyRuleLocation, securityPolicyRuleEntry, securityPolicyRuleReply.Name)
	if err != nil {
		log.Printf("Failed to update security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s:%s' with description '%s' updated", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, *securityPolicyRuleReply.Description)

	// SECURITY POLICY RULE - READ BY ID
	securityPolicyRuleReply, err = securityPolicyRuleApi.ReadById(ctx, securityPolicyRuleLocation, *securityPolicyRuleReply.Uuid, "get")
	if err != nil {
		log.Printf("Failed to update security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s:%s' with description '%s' read by id", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, *securityPolicyRuleReply.Description)

	// SECURITY POLICY RULE - UPDATE 2
	securityPolicyRuleEntry.Description = util.String("changed by id description")
	securityPolicyRuleReply, err = securityPolicyRuleApi.UpdateById(ctx, securityPolicyRuleLocation, securityPolicyRuleEntry, *securityPolicyRuleReply.Uuid)
	if err != nil {
		log.Printf("Failed to update security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s:%s' with description '%s' updated", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, *securityPolicyRuleReply.Description)

	// SECURITY POLICY RULE - HIT COUNT
	hitCount, err := securityPolicyRuleApi.HitCount(ctx, securityPolicyRuleLocation, "test-policy")
	if err != nil {
		log.Printf("Failed to get hit count for security policy rule: %s", err)
		return
	}
	if len(hitCount) > 0 {
		log.Printf("Security policy rule '%d' hit count", hitCount[0].HitCount)
	} else {
		log.Printf("Security policy rule not found")
	}

	// SECURITY POLICY RULE - AUDIT COMMENT
	err = securityPolicyRuleApi.SetAuditComment(ctx, securityPolicyRuleLocation, securityPolicyRuleReply.Name, "another audit comment")
	if err != nil {
		log.Printf("Failed to set audit comment for security policy rule: %s", err)
		return
	}

	comment, err := securityPolicyRuleApi.CurrentAuditComment(ctx, securityPolicyRuleLocation, securityPolicyRuleEntry.Name)
	if err != nil {
		log.Printf("Failed to get audit comment for security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s:%s' current comment: '%s'", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, comment)

	comments, err := securityPolicyRuleApi.AuditCommentHistory(ctx, securityPolicyRuleLocation, securityPolicyRuleEntry.Name, "forward", 10, 0)
	if err != nil {
		log.Printf("Failed to get audit comments for security policy rule: %s", err)
		return
	}
	for _, comment := range comments {
		log.Printf("Security policy rule '%s:%s' comment history: '%s:%s'", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name, comment.Time, comment.Comment)
	}

	// SECURITY POLICY RULE - DELETE
	err = securityPolicyRuleApi.DeleteById(ctx, securityPolicyRuleLocation, *securityPolicyRuleReply.Uuid)
	if err != nil {
		log.Printf("Failed to delete security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s' deleted", securityPolicyRuleReply.Name)

	// SECURITY POLICY RULE - FORCE ERROR WHILE DELETE
	err = securityPolicyRuleApi.Delete(ctx, securityPolicyRuleLocation, securityPolicyRuleReply.Name)
	if err != nil {
		log.Printf("Failed to delete security policy rule: %s", err)
	} else {
		log.Printf("Security policy rule '%s' deleted", securityPolicyRuleReply.Name)
	}
}

func checkSecurityPolicyRulesMove(c *pango.XmlApiClient, ctx context.Context) {
	// SECURITY POLICY RULE - MOVE GROUP
	securityPolicyRuleLocation := security.Location{
		Vsys: &security.VsysLocation{
			NgfwDevice: "localhost.localdomain",
			Rulebase:   "post-rulebase",
			Vsys:       "vsys1",
		},
	}

	securityPolicyRuleApi := security.NewService(c)

	securityPolicyRulesNames := make([]string, 10)
	var securityPolicyRulesEntries []security.Entry
	for i := 0; i < 10; i++ {
		securityPolicyRulesNames[i] = fmt.Sprintf("codegen_rule%d", i)
		securityPolicyRuleItem := security.Entry{
			Name:               securityPolicyRulesNames[i],
			Description:        util.String("initial description"),
			Action:             util.String("allow"),
			SourceZone:         []string{"any"},
			SourceAddress:      []string{"any"},
			DestinationZone:    []string{"any"},
			DestinationAddress: []string{"any"},
			Application:        []string{"any"},
			Service:            []string{"application-default"},
		}
		securityPolicyRulesEntries = append(securityPolicyRulesEntries, securityPolicyRuleItem)
		securityPolicyRuleItemReply, err := securityPolicyRuleApi.Create(ctx, securityPolicyRuleLocation, securityPolicyRuleItem)
		if err != nil {
			log.Printf("Failed to create security policy rule: %s", err)
			return
		}
		log.Printf("Security policy rule '%s:%s' with description '%s' created", *securityPolicyRuleItemReply.Uuid, securityPolicyRuleItemReply.Name, *securityPolicyRuleItemReply.Description)
	}
	rulePositionBefore7 := rule.Position{
		First:           nil,
		Last:            nil,
		SomewhereBefore: nil,
		DirectlyBefore:  util.String("codegen_rule7"),
		SomewhereAfter:  nil,
		DirectlyAfter:   nil,
	}
	rulePositionBottom := rule.Position{
		First:           nil,
		Last:            util.Bool(true),
		SomewhereBefore: nil,
		DirectlyBefore:  nil,
		SomewhereAfter:  nil,
		DirectlyAfter:   nil,
	}
	var securityPolicyRulesEntriesToMove []security.Entry
	securityPolicyRulesEntriesToMove = append(securityPolicyRulesEntriesToMove, securityPolicyRulesEntries[3])
	securityPolicyRulesEntriesToMove = append(securityPolicyRulesEntriesToMove, securityPolicyRulesEntries[5])
	for _, securityPolicyRuleItemToMove := range securityPolicyRulesEntriesToMove {
		log.Printf("Security policy rule '%s' is going to be moved", securityPolicyRuleItemToMove.Name)
	}
	err := securityPolicyRuleApi.MoveGroup(ctx, securityPolicyRuleLocation, rulePositionBefore7, securityPolicyRulesEntriesToMove)
	if err != nil {
		log.Printf("Failed to move security policy rules %v: %s", securityPolicyRulesEntriesToMove, err)
		return
	}
	securityPolicyRulesEntriesToMove = []security.Entry{securityPolicyRulesEntries[1]}
	for _, securityPolicyRuleItemToMove := range securityPolicyRulesEntriesToMove {
		log.Printf("Security policy rule '%s' is going to be moved", securityPolicyRuleItemToMove.Name)
	}
	err = securityPolicyRuleApi.MoveGroup(ctx, securityPolicyRuleLocation, rulePositionBottom, securityPolicyRulesEntriesToMove)
	if err != nil {
		log.Printf("Failed to move security policy rules %v: %s", securityPolicyRulesEntriesToMove, err)
		return
	}
	err = securityPolicyRuleApi.Delete(ctx, securityPolicyRuleLocation, securityPolicyRulesNames...)
	if err != nil {
		log.Printf("Failed to delete security policy rules %s: %s", securityPolicyRulesNames, err)
		return
	}
}

func checkTag(c *pango.XmlApiClient, ctx context.Context) {
	// TAG - CREATE
	tagColor := tag.ColorAzureBlue
	tagObject := tag.Entry{
		Name:  "codegen_color",
		Color: &tagColor,
	}

	tagLocation := tag.Location{
		Shared: true,
	}

	tagApi := tag.NewService(c)
	tagReply, err := tagApi.Create(ctx, tagLocation, tagObject)
	if err != nil {
		log.Printf("Failed to create object: %s", err)
		return
	}
	log.Printf("Tag '%s' created", tagReply.Name)

	// TAG - DELETE
	err = tagApi.Delete(ctx, tagLocation, tagReply.Name)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Printf("Tag '%s' deleted", tagReply.Name)
}

func checkAddress(c *pango.XmlApiClient, ctx context.Context) {
	// ADDRESS - CREATE
	addressObject := address.Entry{
		Name:      "codegen_address_test1",
		IpNetmask: util.String("12.13.14.25"),
	}

	addressLocation := address.Location{
		Shared: true,
	}

	addressApi := address.NewService(c)
	addressReply, err := addressApi.Create(ctx, addressLocation, addressObject)
	if err != nil {
		log.Printf("Failed to create object: %s", err)
		return
	}
	log.Printf("Address '%s=%s' created", addressReply.Name, *addressReply.IpNetmask)

	// ADDRESS - DELETE
	err = addressApi.Delete(ctx, addressLocation, addressReply.Name)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Printf("Address '%s' deleted", addressReply.Name)

	// ADDRESS - LIST
	addresses, err := addressApi.List(ctx, addressLocation, "get", "name starts-with 'wu'", "'")
	if err != nil {
		log.Printf("Failed to list object: %s", err)
	} else {
		for index, item := range addresses {
			log.Printf("Address %d: '%s'", index, item.Name)
		}
	}
}

func checkService(c *pango.XmlApiClient, ctx context.Context) {
	// SERVICE - ADD
	servicePort := 8642
	serviceObject := service.Entry{
		Name:        "codegen_service_test1",
		Description: util.String("test description"),
		Protocol: &service.SpecProtocol{
			Tcp: &service.SpecProtocolTcp{
				DestinationPort: &servicePort,
				Override: &service.SpecProtocolTcpOverride{
					No: util.String(""),
				},
			},
		},
	}

	serviceLocation := service.Location{
		Shared: false,
		Vsys: &service.VsysLocation{
			NgfwDevice: "localhost.localdomain",
			Vsys:       "vsys1",
		},
	}

	serviceApi := service.NewService(c)
	serviceReply, err := serviceApi.Create(ctx, serviceLocation, serviceObject)
	if err != nil {
		log.Printf("Failed to create object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' created", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - UPDATE 1
	serviceObject.Description = util.String("changed description")

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, serviceObject, serviceReply.Name)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' updated", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - UPDATE 2
	servicePort = 1234
	serviceObject.Protocol.Tcp.DestinationPort = &servicePort

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, serviceObject, serviceReply.Name)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' updated", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - RENAME
	newServiceName := "codegen_service_test2"
	serviceObject.Name = newServiceName

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, serviceObject, serviceReply.Name)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' renamed", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - LIST
	//services, err := serviceApi.List(ctx, serviceLocation, "get", "name starts-with 'test'", "'")
	services, err := serviceApi.List(ctx, serviceLocation, "get", "", "")
	if err != nil {
		log.Printf("Failed to list object: %s", err)
	} else {
		for index, item := range services {
			log.Printf("Service %d: '%s'", index, item.Name)
		}
	}

	// SERVICE - DELETE
	err = serviceApi.Delete(ctx, serviceLocation, newServiceName)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Printf("Service '%s' deleted", newServiceName)

	// SERVICE - READ
	serviceLocation = service.Location{
		Shared: false,
		Vsys: &service.VsysLocation{
			NgfwDevice: "localhost.localdomain",
			Vsys:       "vsys1",
		},
	}

	serviceApi = service.NewService(c)
	serviceReply, err = serviceApi.Read(ctx, serviceLocation, "test", "get")
	if err != nil {
		log.Printf("Failed to read object: %s", err)
		return
	}
	readDescription := ""
	if serviceReply.Description != nil {
		readDescription = *serviceReply.Description
	}
	keys := make([]string, 0, len(serviceReply.Misc))
	xmls := make([]string, 0, len(serviceReply.Misc))
	for key := range serviceReply.Misc {
		keys = append(keys, key)
		data, _ := xml.Marshal(serviceReply.Misc[key])
		xmls = append(xmls, string(data))
	}
	log.Printf("Service '%s=%d, description: %s misc XML: %s, misc keys: %s' read",
		serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort, readDescription, xmls, keys)

	// SERVICE - UPDATE 3
	serviceReply.Description = util.String("some text changed now")

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, *serviceReply, "test")
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	readDescription = ""
	if serviceReply.Description != nil {
		readDescription = *serviceReply.Description
	}
	keys = make([]string, 0, len(serviceReply.Misc))
	xmls = make([]string, 0, len(serviceReply.Misc))
	for key := range serviceReply.Misc {
		keys = append(keys, key)
		data, _ := xml.Marshal(serviceReply.Misc[key])
		xmls = append(xmls, string(data))
	}
	log.Printf("Service '%s=%d, description: %s misc XML: %s, misc keys: %s' update",
		serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort, readDescription, xmls, keys)
}

func checkNtp(c *pango.XmlApiClient, ctx context.Context) {
	// NTP - ADD
	ntpConfig := ntp.Config{
		NtpServers: &ntp.SpecNtpServers{
			PrimaryNtpServer: &ntp.SpecNtpServersPrimaryNtpServer{
				NtpServerAddress: util.String("11.12.13.14"),
			},
		},
	}

	ntpLocation := ntp.Location{
		System: &ntp.SystemLocation{
			NgfwDevice: "localhost.localdomain",
		},
	}

	ntpApi := ntp.NewService(c)
	ntpReply, err := ntpApi.Create(ctx, ntpLocation, ntpConfig)
	if err != nil {
		log.Printf("Failed to create NTP: %s", err)
		return
	}
	log.Printf("NTP '%s' created", *ntpReply.NtpServers.PrimaryNtpServer.NtpServerAddress)

	// NTP - DELETE
	err = ntpApi.Delete(ctx, ntpLocation, ntpConfig)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Print("NTP deleted")
	return
}

func checkDns(c *pango.XmlApiClient, ctx context.Context) {
	// DNS - ADD
	refreshTime := 27
	dnsConfig := dns.Config{
		DnsSetting: &dns.SpecDnsSetting{
			Servers: &dns.SpecDnsSettingServers{
				Primary:   util.String("8.8.8.8"),
				Secondary: util.String("4.4.4.4"),
			},
		},
		FqdnRefreshTime: &refreshTime,
	}

	dnsLocation := dns.Location{
		System: &dns.SystemLocation{
			NgfwDevice: "localhost.localdomain",
		},
	}

	dnsApi := dns.NewService(c)
	dnsReply, err := dnsApi.Create(ctx, dnsLocation, dnsConfig)
	if err != nil {
		log.Printf("Failed to create DNS: %s", err)
		return
	}
	log.Printf("DNS '%s, %s' created", *dnsReply.DnsSetting.Servers.Primary, *dnsReply.DnsSetting.Servers.Secondary)

	// DNS - DELETE
	err = dnsApi.Delete(ctx, dnsLocation, dnsConfig)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Print("DNS deleted")
}
