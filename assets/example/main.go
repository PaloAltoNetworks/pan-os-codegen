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
)

func main() {
	var err error
	x := context.Background()

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

	if err = c.Initialize(x); err != nil {
		log.Printf("Failed to initialize client: %s", err)
		return
	}

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
	tagReply, err := tagApi.Create(x, tagLocation, tagObject)
	if err != nil {
		log.Printf("Failed to create object: %s", err)
		return
	}
	log.Printf("Tag '%s' created", tagReply.Name)

	// TAG - DELETE
	err = tagApi.Delete(x, tagLocation, tagName)
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
	addressReply, err := addressApi.Create(x, addressLocation, addressObject)
	if err != nil {
		log.Printf("Failed to create object: %s", err)
		return
	}
	log.Printf("Address '%s=%s' created", addressReply.Name, *addressReply.IpNetmask)

	// ADDRESS - DELETE
	err = addressApi.Delete(x, addressLocation, addressName)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Printf("Address '%s' deleted", addressName)

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
	serviceReply, err := serviceApi.Create(x, serviceLocation, serviceObject)
	if err != nil {
		log.Printf("Failed to create object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' created", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - UPDATE 1
	serviceDescription = "changed description"
	serviceObject.Description = &serviceDescription

	serviceReply, err = serviceApi.Update(x, serviceLocation, serviceObject, serviceName)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' updated", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - UPDATE 2
	servicePort = 1234
	serviceObject.Protocol.Tcp.DestinationPort = &servicePort

	serviceReply, err = serviceApi.Update(x, serviceLocation, serviceObject, serviceName)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' updated", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - RENAME
	newServiceName := "codegen_service_test2"
	serviceObject.Name = newServiceName

	serviceReply, err = serviceApi.Update(x, serviceLocation, serviceObject, serviceName)
	if err != nil {
		log.Printf("Failed to update object: %s", err)
		return
	}
	log.Printf("Service '%s=%d' renamed", serviceReply.Name, *serviceReply.Protocol.Tcp.DestinationPort)

	// SERVICE - DELETE
	err = serviceApi.Delete(x, serviceLocation, newServiceName)
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
	serviceReply, err = serviceApi.Read(x, serviceLocation, "test", "get")
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

	serviceReply, err = serviceApi.Update(x, serviceLocation, *serviceReply, "test")
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
	ntpReply, err := ntpApi.Create(x, ntpLocation, ntpConfig)
	if err != nil {
		log.Printf("Failed to create NTP: %s", err)
		return
	}
	log.Printf("NTP '%s' created", *ntpReply.NtpServers.PrimaryNtpServer.NtpServerAddress)

	// NTP - DELETE
	err = ntpApi.Delete(x, ntpLocation, ntpConfig)
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
	dnsReply, err := dnsApi.Create(x, dnsLocation, dnsConfig)
	if err != nil {
		log.Printf("Failed to create DNS: %s", err)
		return
	}
	log.Printf("DNS '%s, %s' created", *dnsReply.DnsSetting.Servers.Primary, *dnsReply.DnsSetting.Servers.Secondary)

	// DNS - DELETE
	err = dnsApi.Delete(x, dnsLocation, dnsConfig)
	if err != nil {
		log.Printf("Failed to delete object: %s", err)
		return
	}
	log.Print("DNS deleted")
}
