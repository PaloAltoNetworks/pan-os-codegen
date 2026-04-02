package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/PaloAltoNetworks/pango"
	"github.com/PaloAltoNetworks/pango/objects/address"
	"github.com/PaloAltoNetworks/pango/panorama/devicegroup"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

// This example demonstrates partial field read operations with device groups and addresses.
// It creates a device group with multiple fields populated, then creates 2 addresses
// within that device group. Finally, it demonstrates how to read the device group
// metadata without fetching all the nested configuration (templates, devices, etc.)

func main() {
	ctx := context.Background()

	// Connect to Panorama
	hostname := os.Getenv("PANOS_HOSTNAME")
	username := os.Getenv("PANOS_USERNAME")
	password := os.Getenv("PANOS_PASSWORD")

	if hostname == "" || username == "" || password == "" {
		log.Fatal("Please set PANOS_HOSTNAME, PANOS_USERNAME, and PANOS_PASSWORD environment variables")
	}

	useCredentials := true
	client := &pango.Client{
		Hostname:              hostname,
		Username:              username,
		Password:              password,
		SkipVerifyCertificate: true,
		UseCredentials:        &useCredentials,
		CheckEnvironment:      true,
		Logging: pango.LoggingInfo{
			LogCategories: pango.LogCategorySend | pango.LogCategoryReceive | pango.LogCategoryPango | pango.LogCategorySensitive,
			LogLevel:      slog.LevelDebug,
		},
	}

	if err := client.Setup(); err != nil {
		log.Fatalf("Failed to connect to Panorama: %v", err)
	}

	fmt.Printf("Connected to Panorama: %s\n", hostname)
	fmt.Printf("PAN-OS Version: %s\n\n", client.Versioning())

	// Service instances
	dgSvc := devicegroup.NewService(client)
	addrSvc := address.NewService(client)

	// Test device group name
	dgName := fmt.Sprintf("test-dg-%d", time.Now().Unix())

	// Step 1: Create device group with multiple fields
	fmt.Println("=== Step 1: Create Device Group with Multiple Fields ===")

	newDG := &devicegroup.Entry{
		Name:              dgName,
		Description:       util.String("Test device group for selective read operations"),
		AuthorizationCode: util.String("AUTH123456"),
	}

	dgLoc := devicegroup.NewPanoramaLocation()
	createdDG, err := dgSvc.Create(ctx, *dgLoc, newDG)
	if err != nil {
		log.Fatalf("Failed to create device group: %v", err)
	}
	fmt.Printf("Created device group: %s\n", createdDG.Name)
	if createdDG.Description != nil {
		fmt.Printf("  Description: %s\n", *createdDG.Description)
	}
	if createdDG.AuthorizationCode != nil {
		fmt.Printf("  Auth Code: %s\n", *createdDG.AuthorizationCode)
	}

	// Step 2: Create 2 addresses in the device group using MultiConfig
	fmt.Println("\n=== Step 2: Create 2 Addresses in Device Group (using MultiConfig) ===")

	addrLoc := address.NewDeviceGroupLocation()
	addrLoc.DeviceGroup.DeviceGroup = dgName

	// Build multi-config request with 2 address creations
	addrXpathPrefix, err := addrLoc.XpathPrefix(client.Versioning())
	if err != nil {
		log.Fatalf("Failed to get address XPath prefix: %v", err)
	}
	addrXpathBase := util.AsXpath(append(addrXpathPrefix, "address"))

	multiConfigOps := ""
	for i := 1; i <= 2; i++ {
		addrName := fmt.Sprintf("test-addr-%d", i)
		addrDesc := fmt.Sprintf("Test address number %d", i)
		addrIp := fmt.Sprintf("10.1.%d.0/24", i)

		// Build entry XML for this address
		entryXML := fmt.Sprintf(`<entry name="%s"><description>%s</description><ip-netmask>%s</ip-netmask></entry>`,
			addrName, addrDesc, addrIp)

		// Add to multi-config operations
		multiConfigOps += fmt.Sprintf(`<set xpath="%s">%s</set>`, addrXpathBase, entryXML)
	}

	// Send multi-config request
	multiConfigXML := fmt.Sprintf("<multi-configure-request>%s</multi-configure-request>", multiConfigOps)

	cmd := &xmlapi.Config{
		Action:  "multi-config",
		Element: multiConfigXML,
		Target:  client.GetTarget(),
	}

	startTime := time.Now()
	_, _, err = client.Communicate(ctx, cmd, false, nil)
	multiConfigDuration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Failed to create addresses via MultiConfig: %v", err)
	}

	fmt.Printf("Created 2 addresses via MultiConfig in %v\n", multiConfigDuration)
	for i := 1; i <= 2; i++ {
		fmt.Printf("  Created address: test-addr-%d (10.1.%d.0/24)\n", i, i)
	}

	// Step 3: Demonstrate PAN-OS Limitation with Multi-Entry Partial Field Reads
	fmt.Println("\n=== Step 3: Multi-Entry Partial Field Read Limitation ===")
	fmt.Println("NOTE: This demonstrates why multi-entry partial field reads don't work with PAN-OS")
	fmt.Println("XPath: .../address/entry[@name='test-addr-1' or @name='test-addr-2']/*[name()='description']")
	fmt.Println("Problem: PAN-OS returns unwrapped fields without entry context")
	fmt.Println("Result: Cannot associate which field belongs to which entry")

	// Build XPath for multiple entries using util.AsEntryXpath with multiple names
	addrMultiXpath := append(addrXpathPrefix, "address", util.AsEntryXpath("test-addr-1", "test-addr-2"))

	// Read with partial field selection - demonstrates the limitation
	startTime = time.Now()

	// Get marshaller for partial field read
	vn := client.Versioning()
	marshaller, err := address.Versioning(vn)
	if err != nil {
		log.Fatalf("Failed to get marshaller: %v", err)
	}

	// Build XPath with field selection
	xpathWithFields := util.AsXpath(addrMultiXpath) + "/*[name()='description']"

	fmt.Printf("XPath being sent: %s\n", xpathWithFields)

	cmd = &xmlapi.Config{
		Action: "get",
		Xpath:  xpathWithFields,
		Target: client.GetTarget(),
	}

	// Create normalizer container
	container := marshaller.NewNormalizer()

	_, _, err = client.Communicate(ctx, cmd, true, container)
	multiReadDuration := time.Since(startTime)

	if err != nil {
		log.Printf("Failed to read multiple addresses (partial field): %v", err)
	} else {
		// Try to normalize - this will fail because PAN-OS returns unwrapped fields
		// Response looks like: <description>val1</description><description>val2</description>
		// No <entry> wrappers, so we can't tell which description belongs to which address
		multiAddrs, err := container.Normalize()

		if err != nil {
			log.Printf("Failed to normalize addresses: %v", err)
		} else {
			fmt.Printf("Multi-Entry Partial Field Read completed in %v\n", multiReadDuration)
			fmt.Printf("Found %d addresses (EXPECTED: 0 due to PAN-OS limitation)\n", len(multiAddrs))
			if len(multiAddrs) == 0 {
				fmt.Printf("  ✓ Confirmed: Multi-entry partial field reads don't work\n")
				fmt.Printf("  Reason: PAN-OS returns unwrapped fields without entry association\n")
			} else {
				fmt.Printf("  ⚠ Unexpected: Got %d addresses (this shouldn't happen)\n", len(multiAddrs))
			}
		}
	}

	// Compare with reading both individually
	fmt.Println("\n  Comparing with individual reads:")
	startTime = time.Now()
	addr1, _ := addrSvc.ReadWithOptions(ctx, *addrLoc, "test-addr-1", "get", address.WithFields("description"))
	addr2, _ := addrSvc.ReadWithOptions(ctx, *addrLoc, "test-addr-2", "get", address.WithFields("description"))
	individualReadDuration := time.Since(startTime)

	fmt.Printf("  Individual reads (2 requests) completed in %v\n", individualReadDuration)
	if addr1 != nil && addr2 != nil {
		fmt.Printf("  Both addresses read successfully\n")

		if individualReadDuration > multiReadDuration {
			timeImprovement := float64(individualReadDuration-multiReadDuration) / float64(individualReadDuration) * 100
			fmt.Printf("  Time improvement with multi-entry read: %.1f%%\n", timeImprovement)
		}
	}

	// Step 4: Standard Read (fetches all fields)
	// IMPORTANT: This reads the entire device group XML, including ALL nested configuration
	// like the 2 addresses we created. Even though addresses don't appear in the Entry struct,
	// they are transferred in the XML payload and stored in Misc[]
	fmt.Println("\n=== Step 4: Standard Read (All Fields) ===")
	fmt.Println("NOTE: This fetches the ENTIRE device group XML including all embedded addresses!")
	fmt.Println("XPath used: /config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='" + dgName + "']")

	// Construct XPath manually using ReadWithXpath to avoid name formatting issues
	dgXpathPrefix, err := dgLoc.XpathPrefix(client.Versioning())
	if err != nil {
		log.Fatalf("Failed to get device group XPath prefix: %v", err)
	}
	dgXpath := append(dgXpathPrefix, "device-group", util.AsEntryXpath(dgName))

	startTime = time.Now()
	fullDG, err := dgSvc.ReadWithXpath(ctx, util.AsXpath(dgXpath), "get")
	fullReadDuration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Failed to read device group (full): %v", err)
	}

	// Calculate approximate XML payload size from Misc field
	fullMiscSize := 0
	for _, misc := range fullDG.Misc {
		// Each Misc element represents nested XML that wasn't unmarshalled
		fullMiscSize += len(fmt.Sprintf("%v", misc))
	}

	fmt.Printf("Standard Read completed in %v\n", fullReadDuration)
	fmt.Printf("Device Group: %s\n", fullDG.Name)
	if fullDG.Description != nil {
		fmt.Printf("  Description: %s\n", *fullDG.Description)
	}
	if fullDG.AuthorizationCode != nil {
		fmt.Printf("  Auth Code: %s\n", *fullDG.AuthorizationCode)
	}
	fmt.Printf("  Templates: %v\n", fullDG.Templates)
	fmt.Printf("  Devices: %d\n", len(fullDG.Devices))
	fmt.Printf("  Misc XML elements: %d (includes all nested addresses!)\n", len(fullDG.Misc))
	fmt.Printf("  Approximate Misc data size: %d bytes\n", fullMiscSize)

	// Step 4: Partial Field Read (only description field)
	// CRITICAL: This uses XPath wildcard predicates to fetch ONLY the description field
	// XPath: /config/.../device-group/entry[@name='test-dg']/*[name()='description']
	// PAN-OS will NOT return the nested addresses at all - they're filtered at the server!
	fmt.Println("\n=== Step 4: Partial Field Read (Description Only) ===")
	fmt.Println("NOTE: XPath wildcard filters at the SERVER - addresses are NOT transferred!")
	fmt.Println("NOTE: Only top-level fields can be selected, not nested objects")
	fmt.Println("XPath used: /config/.../device-group/entry[@name='" + dgName + "']/*[name()='description']")

	startTime = time.Now()
	selectiveDG, err := dgSvc.ReadWithOptions(ctx, *dgLoc, dgName, "get",
		devicegroup.WithFields("description"))
	selectiveReadDuration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Failed to read device group (partial field): %v", err)
	}

	// Calculate approximate XML payload size from Misc field
	selectiveMiscSize := 0
	for _, misc := range selectiveDG.Misc {
		selectiveMiscSize += len(fmt.Sprintf("%v", misc))
	}

	fmt.Printf("Partial Field Read completed in %v\n", selectiveReadDuration)
	fmt.Printf("Device Group: %s\n", selectiveDG.Name)
	if selectiveDG.Description != nil {
		fmt.Printf("  Description: %s (SELECTED)\n", *selectiveDG.Description)
	}
	fmt.Printf("  Auth Code populated: %t (not selected, should be nil)\n", selectiveDG.AuthorizationCode != nil)
	fmt.Printf("  Templates populated: %t (not selected, should be empty)\n", len(selectiveDG.Templates) > 0)
	fmt.Printf("  Devices populated: %t (not selected, should be empty)\n", len(selectiveDG.Devices) > 0)
	fmt.Printf("  Misc XML elements: %d (should be MUCH less than full read!)\n", len(selectiveDG.Misc))
	fmt.Printf("  Approximate Misc data size: %d bytes\n", selectiveMiscSize)

	// Step 5: Partial Field Read (multiple fields)
	fmt.Println("\n=== Step 5: Partial Field Read (Description + Auth Code) ===")

	startTime = time.Now()
	multiSelectDG, err := dgSvc.ReadWithOptions(ctx, *dgLoc, dgName, "get",
		devicegroup.WithFields("description", "authorization-code"))
	multiSelectDuration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Failed to read device group (multi-select): %v", err)
	}

	fmt.Printf("Multi-field Partial Field Read completed in %v\n", multiSelectDuration)
	fmt.Printf("Device Group: %s\n", multiSelectDG.Name)
	if multiSelectDG.Description != nil {
		fmt.Printf("  Description: %s (SELECTED)\n", *multiSelectDG.Description)
	}
	if multiSelectDG.AuthorizationCode != nil {
		fmt.Printf("  Auth Code: %s (SELECTED)\n", *multiSelectDG.AuthorizationCode)
	}
	fmt.Printf("  Templates populated: %t (not selected, should be empty)\n", len(multiSelectDG.Templates) > 0)
	fmt.Printf("  Devices populated: %t (not selected, should be empty)\n", len(multiSelectDG.Devices) > 0)

	// Step 6: Partial Field Read (exclude fields)
	fmt.Println("\n=== Step 6: Partial Field Read (Exclude Authorization Code) ===")

	excludeDG, err := dgSvc.ReadWithOptions(ctx, *dgLoc, dgName, "get",
		devicegroup.WithoutFields("authorization-code"))

	if err != nil {
		log.Fatalf("Failed to read device group (exclude): %v", err)
	}

	fmt.Printf("Device Group: %s\n", excludeDG.Name)
	if excludeDG.Description != nil {
		fmt.Printf("  Description: %s (included)\n", *excludeDG.Description)
	}
	fmt.Printf("  Auth Code populated: %t (excluded, should be nil)\n", excludeDG.AuthorizationCode != nil)

	// Step 7: Performance Comparison
	fmt.Println("\n=== Step 7: Performance Comparison ===")
	fmt.Println("\n*** CRITICAL INSIGHT ***")
	fmt.Println("The Standard Read fetched the ENTIRE device group XML including all addresses,")
	fmt.Println("even though the addresses ended up in Misc[] and aren't exposed in the Entry struct.")
	fmt.Println("The Partial Field Read used XPath wildcards to filter at the SERVER, so addresses")
	fmt.Println("were NEVER transferred over the network!")
	fmt.Println()
	fmt.Printf("Standard Read:              %v\n", fullReadDuration)
	fmt.Printf("Partial Field Read (1 fld): %v\n", selectiveReadDuration)
	fmt.Printf("Partial Field Read (2 fld): %v\n", multiSelectDuration)
	fmt.Println()
	fmt.Printf("Misc XML in Standard Read:      %d bytes (includes addresses)\n", fullMiscSize)
	fmt.Printf("Misc XML in Partial Field Read: %d bytes (NO addresses)\n", selectiveMiscSize)

	if fullMiscSize > 0 && selectiveMiscSize >= 0 {
		payloadReduction := float64(fullMiscSize-selectiveMiscSize) / float64(fullMiscSize) * 100
		fmt.Printf("\nPayload reduction: %.1f%%\n", payloadReduction)
	}

	if fullReadDuration > selectiveReadDuration {
		timeImprovement := float64(fullReadDuration-selectiveReadDuration) / float64(fullReadDuration) * 100
		fmt.Printf("Time improvement: %.1f%%\n", timeImprovement)
	}

	// Step 8: Demonstrate reading addresses with partial field selection
	fmt.Println("\n=== Step 8: Partial Field Read on Addresses ===")

	// Read first address with all fields
	fullAddr, err := addrSvc.Read(ctx, *addrLoc, "test-addr-1", "get")
	if err != nil {
		log.Printf("Failed to read address (full): %v", err)
	} else {
		fmt.Printf("Standard Address Read:\n")
		fmt.Printf("  Name: %s\n", fullAddr.Name)
		if fullAddr.Description != nil {
			fmt.Printf("  Description: %s\n", *fullAddr.Description)
		}
		if fullAddr.IpNetmask != nil {
			fmt.Printf("  IP Netmask: %s\n", *fullAddr.IpNetmask)
		}
	}

	// Read same address with only description
	selectiveAddr, err := addrSvc.ReadWithOptions(ctx, *addrLoc, "test-addr-1", "get",
		address.WithFields("description"))
	if err != nil {
		log.Printf("Failed to read address (partial field): %v", err)
	} else {
		fmt.Printf("\nPartial Field Address Read (description only):\n")
		fmt.Printf("  Name: %s\n", selectiveAddr.Name)
		if selectiveAddr.Description != nil {
			fmt.Printf("  Description: %s (SELECTED)\n", *selectiveAddr.Description)
		}
		fmt.Printf("  IP Netmask populated: %t (not selected, should be nil)\n", selectiveAddr.IpNetmask != nil)
	}

	// Cleanup
	fmt.Println("\n=== Cleanup ===")

	// Delete addresses
	fmt.Println("Deleting addresses...")
	for i := 1; i <= 2; i++ {
		addrName := fmt.Sprintf("test-addr-%d", i)
		if err := addrSvc.Delete(ctx, *addrLoc, addrName); err != nil {
			log.Printf("Failed to delete address %s: %v", addrName, err)
		}
	}
	fmt.Println("  Deleted 2 addresses")

	// Delete device group
	fmt.Println("Deleting device group...")
	if err := dgSvc.Delete(ctx, *dgLoc, dgName); err != nil {
		log.Printf("Failed to delete device group: %v", err)
	} else {
		fmt.Printf("  Deleted device group: %s\n", dgName)
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("\n*** KEY TAKEAWAYS ***")
	fmt.Println()
	fmt.Println("1. STANDARD READ PROBLEM:")
	fmt.Println("   - Fetches entire device group XML including ALL nested config (addresses, etc.)")
	fmt.Println("   - Nested config goes into Misc[] field (not exposed in Entry struct)")
	fmt.Println("   - Must transfer ALL data over network even though you don't use it")
	fmt.Println("   - Even with just 2 addresses, you can see the Misc[] field contains embedded XML")
	fmt.Println()
	fmt.Println("2. PARTIAL FIELD READ SOLUTION:")
	fmt.Println("   - Uses XPath wildcard predicates to filter at the SERVER")
	fmt.Println("   - XPath: /config/.../entry[@name='dg']/*[name()='description']")
	fmt.Println("   - PAN-OS never returns nested addresses - filtered before transmission")
	fmt.Println("   - Massive payload reduction (especially with nested objects)")
	fmt.Println("   - Only works with top-level fields, not nested objects")
	fmt.Println()
	fmt.Println("3. PERFORMANCE BENEFITS:")
	fmt.Println("   - Reduced network transfer (payload reduction visible even with 2 addresses)")
	fmt.Println("   - Faster API responses")
	fmt.Println("   - Lower memory usage")
	fmt.Println("   - Critical for large Panorama configurations with hundreds of device groups")
	fmt.Println()
	fmt.Println("4. USAGE:")
	fmt.Println("   - Use ReadWithOptions() with WithFields() to select specific fields")
	fmt.Println("   - Use WithoutFields() to exclude specific fields")
	fmt.Println("   - Works for any entry-based resource (addresses, device groups, etc.)")
}
