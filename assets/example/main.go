package main

import (
	"context"
	"encoding/xml"
	"log"

	"github.com/PaloAltoNetworks/pango"
	"github.com/PaloAltoNetworks/pango/device/services/dns"
	"github.com/PaloAltoNetworks/pango/device/services/ntp"
	"github.com/PaloAltoNetworks/pango/objects/address"
	"github.com/PaloAltoNetworks/pango/objects/service"
	"github.com/PaloAltoNetworks/pango/objects/tag"
	"github.com/PaloAltoNetworks/pango/policies/rules/security"
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

	// SECURITY POLICY RULE - ADD
	securityPolicyRuleName := "codegen_rule"
	securityPolicyRuleAction := "allow"
	securityPolicyRuleSourceZones := []string{"any"}
	securityPolicyRuleDestinationZones := []string{"any"}
	securityPolicyRuleEntry := security.Entry{
		Name:            securityPolicyRuleName,
		Action:          &securityPolicyRuleAction,
		SourceZone:      securityPolicyRuleSourceZones,
		DestinationZone: securityPolicyRuleDestinationZones,
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
	log.Printf("Security policy rule '%s:%s' created", *securityPolicyRuleReply.Uuid, securityPolicyRuleReply.Name)

	// SECURITY POLICY RULE - DELETE
	err = securityPolicyRuleApi.Delete(ctx, securityPolicyRuleLocation, securityPolicyRuleName)
	if err != nil {
		log.Printf("Failed to delete security policy rule: %s", err)
		return
	}
	log.Printf("Security policy rule '%s' deleted", securityPolicyRuleName)

	// TAG - CREATE
	tagName := "codegen_color"
	tagColor := tag.ColorAzureBlue
	tagObject := tag.Entry{
		Name:  tagName,
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
	err = tagApi.Delete(ctx, tagLocation, tagName)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Printf("Tag '%s' deleted", tagName)

	// ADDRESS - CREATE
	addressValue := "12.13.14.25"
	addressName := "codegen_address_test1"
	addressObject := address.Entry{
		Name:      addressName,
		IpNetmask: &addressValue,
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
	err = addressApi.Delete(ctx, addressLocation, addressName)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Printf("Address '%s' deleted", addressName)

	// ADDRESS - LIST
	addresses, err := addressApi.List(ctx, addressLocation, "get", "name starts-with 'wu'", "'")
	if err != nil {
		log.Printf("Failed to list object: %s", err)
	} else {
		for index, item := range addresses {
			log.Printf("Address %d: '%s'", index, item.Name)
		}
	}

	// SERVICE - ADD
	serviceName := "codegen_service_test1"
	servicePort := 8642
	serviceDescription := "test description"
	tcpOverride := ""
	serviceObject := service.Entry{
		Name:        serviceName,
		Description: &serviceDescription,
		Protocol: &service.SpecProtocol{
			Tcp: &service.SpecProtocolTcp{
				DestinationPort: &servicePort,
				Override: &service.SpecProtocolTcpOverride{
					No: &tcpOverride,
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
	serviceDescription = "changed description"
	serviceObject.Description = &serviceDescription

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, serviceObject, serviceName)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' updated", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - UPDATE 2
	servicePort = 1234
	serviceObject.Protocol.Tcp.DestinationPort = &servicePort

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, serviceObject, serviceName)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' updated", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - RENAME
	newServiceName := "codegen_service_test2"
	serviceObject.Name = newServiceName

	serviceReply, err = serviceApi.Update(ctx, serviceLocation, serviceObject, serviceName)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' renamed", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERViCE - LIST
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
	serviceDescription = "some text changed now"
	serviceReply.Description = &serviceDescription

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

	// NTP
	ntpAddress := "11.12.13.14"
	ntpConfig := ntp.Config{
		NtpServers: &ntp.SpecNtpServers{
			PrimaryNtpServer: &ntp.SpecNtpServersPrimaryNtpServer{
				NtpServerAddress: &ntpAddress,
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

	// DNS
	primaryDnsAddress := "8.8.8.8"
	secondaryDnsAddress := "4.4.4.4"
	refreshTime := 27
	dnsConfig := dns.Config{
		DnsSetting: &dns.SpecDnsSetting{
			Servers: &dns.SpecDnsSettingServers{
				Primary:   &primaryDnsAddress,
				Secondary: &secondaryDnsAddress,
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
