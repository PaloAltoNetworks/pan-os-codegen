package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	//"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCertificateImport_1(t *testing.T) {
	t.Parallel()

	nameSuffix := acctest.RandStringFromCharSet(6, acctest.CharSetAlphaNum)
	prefix := fmt.Sprintf("test-acc-%s", nameSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: panosCertificateImportTmpl1,
				ConfigVariables: map[string]config.Variable{
					"prefix":      config.StringVariable(prefix),
					"certificate": config.StringVariable(certificateImportCertificatePem),
					"private_key": config.StringVariable(certificateImportPrivateKeyPem),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"panos_certificate_import.example",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("%s-cert", prefix)),
					),
				},
			},
		},
	})
}

const panosCertificateImportTmpl1 = `
variable "prefix" { type = string }
variable "certificate" { type = string }
variable "private_key" { type = string }

resource "panos_template" "example" {
  location = { panorama = {} }

  name = format("%s-tmpl", var.prefix)
}

resource "panos_certificate_import" "example" {
  location = { template = { name = panos_template.example.name } }

  name = format("%s-cert", var.prefix)

  local = {
    pem = {
      certificate = var.certificate
      private_key = var.private_key
    }
  }
}
`

const certificateImportCertificatePem = `
-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----
`

const certificateImportPrivateKeyPem = `
-----BEGIN PRIVATE KEY-----
MIIJQQIBADANBgkqhkiG9w0BAQEFAASCCSswggknAgEAAoICAQCmIPiZDDutIU3G
yQvtJXz4tYgnNBTL2c+Yah1aWCcjLJAkdCChKkAdwwiN1mBLQh975kPH+Ie8tCMI
gOMsxv/BE2MC1p4N7V4gXCj4R6kAI71xDI7uT8gEsjZ3XQ6YddQ6zmb+GkG5sxPO
tOYIhf3ivmGAhaDV5363BK+TAVjI2keX4ZCEWkAiBjacoDi+sUAo6ObRvBpqjr7O
j+yIJEWrT4zITR0beZBkhevq3f8RhL0hEwp9vlUPjbA+GX5Eyv2v+r1y881n+DK6
Kp+UWlWph+Q2mRC016/xE2RvQ48FUa9RTAwmvytjtCtqF+euk5JOYvzvUwjTG+sa
4YJbNftvcg7DCr6mqU9ce720fSTZDUhvyuhtOkcIY2eIjcwQakftcgkhTLxLH8K0
GYiDdJNGEOIEnTiUSP4OZlsQquyJHPWuJ+n6t2sDD8LvqJSaI6P0HLbemzq6lBc/
Ksjq9h3EtSAb7FU9KQLOUEieClHmY3XYqgP84+5yQa/NWco4N556sYqr6oHyWqbe
l4cpKhxiwKTdP8y1t3OTrAqS8G3qNl2wWzcaIJ4UDzg533uDR8Emy9QejumE/iAl
VAHV460BLjIxHrfHBvip6clQYxDHL8Jzhda8v4Rir5SMnJubpXTD836A/WDT7nrB
gciUUPjXajyE4D1Sh1b2k52Hn6GdYwIDAQABAoICACvVolLb3/kwQvHzRYLW8/E6
EQlrHBunxreMNF+MyBLnZMdBnwR3fgB8YErwqGrjMSSDnxnqMYKws1fAjnDXt08u
Ot9aWs0I91+pgaP1YJnpVEi6jBJEmd3nWijHtJy05oF3ycQ9kF8b6duObu4L0PBd
1KNRXx1h3lUTVvJ+lfs1YVOpkHTjzW1M32cXfbGPWoMQ5SqtK/k23hDp9/r6OynX
LSoC8u23d0qW7aeE2RM5x5+tAwUnzhDzDXBtUJx9RVAEZK2qt+W5n0TxDzdZWKYJ
dcWUQMy+5q1BNSyIknnQUmasnr4wjhXaSeROF3NAfAfT5bKOYdM2WCQ5IajIhyW0
4WyDA4d3XdssRfexeiq6N9cOG/PdN4lmd4bGWM+1dNc5z+H7wYlU7msk37wENm8z
nnRowoplSN7kWWh0pWkDIFQDl62p17vgUg0JI5zpxVDuoyT/2HzrmiMcYOOycGHf
fP6uLKy8fDvfBhq3ZhZQX5CJc9kkWUfU/+mI7uizIKHlt7bEmXyLLIekXdxIAYsm
0E0ZdWmT7501yCpvKVQLrOOvcQk/8lUK29syQ8MADah6cpCZRNqB0Y91YyrYIA4T
XPHL7zKIP8CCvNGghH0+b69IYC5nMkOw78trTPXjtY9N5psmYofdAa+MSxwUiHZY
rgQKKp9q1QgH671vvOWxAoIBAQDmKzy4fH3ifB7YfX6Ogw87C0EHAYkq2ji+o+54
uWTlxZi0t3DPN4smf1ig4aSsPyWp1F3iNJfWyctzp9Wco9vUOoTOk9LERAAVdlLW
kX4juNUa+cUTrVVU1ncZEQ6fyN6aq6Az2+En33ewZbqKWpLQpHCQIbY6vWZr3WC7
CxmGNbFzGSlauM1zyO7b5ugIO9zoCuwn858bdwEwxNfr+f8s0j3IPXOpuIi68qeP
P1rKMt/NioylQ3Rg33CmmCD2ZXDRkl6/SaBMUa5IWG7gQ/+22U6wq6P9BnlRqMmc
pSR3Q2L35y4RZ/iYXI+D669NKnLpjZXxs0JupD23Tp8Vho1TAoIBAQC4xdveJjO3
64iiJVWViv6R6gYi2RJV22dIMfDSSlNAZu7t7yYfJkTPuBYpWv5Q6Xqj43OnN0Ac
YXxGEKPcVFUxFXqy3nI+cLA8nYbL43FPYEcehEeQIal+IA+w9WSRMQvL6AjDo3TI
vgOIXs2Jtcr9zTnO9GFKYyIcfFpLfxrffVV2lhk7s7hxz04plV0i4uWFDhiYFLvH
7gSBGYHpwdJYDytNYLiT2EmPImdQJ3kQhWyK2CXLGeNc66buu/KRHFU8RQacYP2t
mSbs5mXcsFB+KOquZ4IV2e2c2JPmMMp5JZZQrdUysi+KpkvH3NCmUzeA5IN3UcbD
1315gGpXmp2xAoIBAA2Lw+ITqZD3vxT8pcMbYX0XF3ejFoCIIUjO+wzt1EtVirww
A5qeaTkVy5CEVx2wBbZuAix67ei9LZUb7o1uc1SVMRW7S28zlVGuCggIvgS6LwiM
ZJXY4KnCiXXXNCYhO0CdEyuaKDEhjLi78/OKixNuahWBdmkUln+IotW/PHxSkqP0
eiOVtrm2vKACgetiIokhg26Cfv2tzkshepevud3YbbxoKXN2oc1m1IewsdYukk9V
dRuQ0buVytpzH5WAuNgMpjjZy25SbFBjq/rU5arMNT5ei6Mri15L8bmfWnsOYze1
yldJ6C6HXAbmiwWelu6533Y/F4zNa7hrDx/EMHMCggEAXHg4pp57t4maYXtJr4NW
D3QNCheUg38/2vOTT8p+i3Z4EH9klqYyPbok7SFqsNeH1skXshGGdi2bYf0l5DgY
Qm47b5S/m9wNduhm81ap+E14ih8tKUaPal1lPOwyHi9rdepzqGT/Jw9g+ThoqIhg
RFAWpCnNHssp4ROipLHBoyM4SBaqHiS9I8fZmBn1+GWQ89uwFzwZFd9aRbmcOH4V
ZJiC1UCYXvUZKxbOmWCHx+rd/UZa85/LF0+fxU4uAM0rIvRwcIZhriU9Q8WyKJXc
Uqbre8i1Y3Yi4iHJMqQsUCCtb0bvsWVXQY4j0qwBh5uR5WF3IZm9XXlUhB/uGFV6
oQKCAQAoByzQQbCvUewtVIRdMjNxVCem40VwBZdSi5CgTRgog6l2PNhieqXhpWCh
OuytvDWpkM7j/hYHvaY/Jp8WlwwB/YlH5+Pl9P3zzyN2nC9VXzkxjGfGWEVIfm8h
12F0KCsq4LKugHhdIzs9ttf7NAjVpbDoIOffmKO089x5A3EbfcjC66QQ3Q5zsCSp
Ir68XVAUCDosXc9aqobkmmF/W6JcTI+/pLL5T8daG9PFnrMJGkMK3XUZzHTXa/5P
yaUmHJhtB4/P0n66qbCgG93nC6+SJRvi2kbZAEaCPvdj3Tf6tDz+r4sw1I1iFNb9
LFwzdpJ73GS5R3EZcLCxKicFiqv3
-----END PRIVATE KEY-----
`
