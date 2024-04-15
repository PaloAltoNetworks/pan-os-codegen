package normalized

import (
	"fmt"
	"strings"
	"text/template"
)

type HashProfile struct {
	// The following are static styles of naming the encrypted value.

	// This just saves the param value itself, useful when there are no other encrypted
	// values in a schema.  Every other encryption naming style is for when there
	// exist multiple encrypted values in a schema.
	Solo *bool `json:"solo" yaml:"solo"`

	// This will pair the param name with the name of a parent name, useful when
	// there are multiple encrypted values in a schema, but they can be differentiated
	// by the class name that contains them.
	//
	// Note that this cannot be used if you cross a ref: border.
	WithParentName *int `json:"with_parent_name" yaml:"with_parent_name"`

	// The following are dynamic styles of naming the encrypted value.

	// This will pair the param name with the value saved in an adjacent param, useful
	// when an adjacent param is a required param and we can rely on it being unique.
	//
	// The value specified should be the param's internal name.
	WithParamValue *string `json:"with_param_value" yaml:"with_param_value"`
}

func (o *HashProfile) IsEncrypted() bool {
	if o.Solo != nil && *o.Solo {
		return true
	} else if o.WithParentName != nil {
		return true
	} else if o.WithParamValue != nil && *o.WithParamValue != "" {
		return true
	}

	return false
}

func (o *HashProfile) IsPlainText() bool {
	if o.Solo != nil && *o.Solo {
		return false
	} else if o.WithParentName != nil {
		return false
	} else if o.WithParamValue != nil && *o.WithParamValue != "" {
		return false
	}

	return true
}

func (o *HashProfile) IsStatic() bool {
	if o.Solo != nil && *o.Solo {
		return true
	} else if o.WithParentName != nil {
		return true
	}

	return false
}

func (o *HashProfile) GetEncryptionKey(i Item, src, varName string, srcIsTfType bool, keyType byte) (string, error) {
	if keyType != 'e' && keyType != 'p' {
		return "", fmt.Errorf("keyType should be 'e' for encrypted or 'p' for plaintext")
	}

	keyTypeMap := map[byte]string{
		'e': "encrypted",
		'p': "plaintext",
	}

	fm := template.FuncMap{
		"Source":                func() string { return src },
		"VarName":               func() string { return varName },
		"SourceIsTerraformType": func() bool { return srcIsTfType },
		"KeyType":               func() string { return keyTypeMap[keyType] },
		"IsSolo":                func() bool { return o.Solo != nil && *o.Solo },
		"IsWithParentName":      func() bool { return o.WithParentName != nil },
		"GetParent": func() (Item, error) {
			if o.WithParentName == nil {
				return nil, fmt.Errorf("with_parent_name is not configured")
			}

			count := *o.WithParentName
			if count < 0 {
				return nil, fmt.Errorf("with_parent_name value is negative")
			} else if count == 0 {
				return nil, fmt.Errorf("with_parent_name value should be 1 or more")
			}

			p := i
			for j := 0; j < count; j++ {
				if p == nil {
					return nil, fmt.Errorf("%s parent num %d is nil", i.GetInternalName(), j+1)
				}
				p = p.GetParent()
			}

			if p == nil {
				return nil, fmt.Errorf("went to parent of nil, too far")
			}
			return p, nil
		},
		"IsWithParamValue": func() bool { return o.WithParamValue != nil && *o.WithParamValue != "" },
		"GetOtherParam": func() (Item, error) {
			if o.WithParamValue == nil || *o.WithParamValue == "" {
				return nil, fmt.Errorf("this is not a with_param_value hash")
			}

			parent := i.GetParent()
			if parent == nil {
				return nil, fmt.Errorf("item.Parent is nil")
			}
			obj, ok := parent.(*Object)
			if !ok {
				return nil, fmt.Errorf("item.Parent is not an object")
			}
			param, ok := obj.Params[*o.WithParamValue]
			if !ok {
				return nil, fmt.Errorf("param %q not present", *o.WithParamValue)
			}

			// Right now, all of the pairings are strings.  If the paired param
			// is not a string, we'll need to also import either fmt or strconv,
			// which means that info needs to be passed up the chain.  I'm going
			// to leave this as an exercise for the future if the paired variable
			// is not a string to meet the release deadline.
			if _, ok := param.(*String); !ok {
				return nil, fmt.Errorf("TODO: allow a non-string paired var")
			}

			return param, nil
		},
		"WithParamValueSuffix": func(p Item) (string, error) {
			if !srcIsTfType {
				return "", fmt.Errorf("source is not terraform type")
			}

			switch p.(type) {
			case *Bool:
				return "ValueBool", nil
			case *Int:
				return "ValueInt64", nil
			case *Float:
				return "ValueFloat64", nil
			case *String:
				return "ValueString", nil
			}

			return "", fmt.Errorf("unsupported param pairing type: %T", p)
		},
		"Unsupported": func() error { return fmt.Errorf("unknown hash style") },
	}

	t := template.Must(
		template.New(
			"encrypted-key",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $item := . }}
{{- if IsSolo }}
    {{ VarName }} := "solo | {{ KeyType }} | {{ $item.GetInternalName }}"
{{- else if IsWithParentName }}
{{- $parent := GetParent }}
    {{ VarName }} := "with_parent_name | {{ KeyType }} | {{ $parent.GetInternalName }} | {{ $item.GetInternalName }}"
{{- else if IsWithParamValue }}
{{- $param := GetOtherParam }}
{{- if SourceIsTerraformType }}
    {{ VarName }} := "with_param_value | {{ KeyType }} | {{ $param.GetInternalName }} | " + {{ Source }}.{{ $param.GetCamelCaseName }}.{{ WithParamValueSuffix $param }}() + " | {{ $item.GetInternalName }}"
{{- else if $param.IsRequired }}
    {{ VarName }} := "with_param_value | {{ KeyType }} | {{ $param.GetInternalName }} | " + {{ Source }}.{{ $param.GetCamelCaseName }} + " | {{ $item.GetInternalName }}"
{{- else }}
    {{ VarName }} := "with_param_value | {{ KeyType }} | {{ $param.GetInternalName }} | "
    if {{ Source }}.{{ $param.GetCamelCaseName }} != nil {
        // NOTE: this is where the fmt or strconv will need to happen.
        {{ VarName }} += *{{ Source }}.{{ $param.GetCamelCaseName }}
    }
    {{ VarName }} += " | {{ $item.GetInternalName }}"
{{- end }}
{{- else }}
{{ Unsupported }}
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, i)

	return b.String(), err
}

func (o *HashProfile) StaticKey(i Item) (string, error) {
	if o.Solo != nil && *o.Solo {
		parts := []string{i.GetShortName(), "solo", i.GetInternalName()}
		return strings.Join(parts, "\n"), nil
	}

	if o.WithParentName != nil {
		count := *o.WithParentName
		if count < 0 {
			return "", fmt.Errorf("%s.hash_profile.with_parent_name value must be >= 0", i.GetInternalName())
		}

		p := i
		for j := 0; j < count; j++ {
			if p == nil {
				return "", fmt.Errorf("%s parent num %d is nil", i.GetInternalName(), j+1)
			}
			p = p.GetParent()
		}

		parts := []string{i.GetShortName(), "with_parent_name", p.GetInternalName(), i.GetInternalName()}
		return strings.Join(parts, "\n"), nil
	}

	return "", fmt.Errorf("%s.hash_profile is not a static type", i.GetInternalName())
}

func (o *HashProfile) IsDynamic() bool {
	if o.WithParamValue != nil && *o.WithParamValue != "" {
		return true
	}
	return false
}

/*
[#/components/schemas/external-dynamic-lists]
WithParentClassName 2 ("ip" + \n + password value)
password: false <nil> []string{"password", "username"} - []string{"", "type", "ip", "auth", "password"}

WithParentClassName 2 ("url" + \n + password value)
password: false <nil> []string{"password", "username"} - []string{"", "type", "url", "auth", "password"}

WithParentClassName 2 ("imsi" + \n + password value)
password: false <nil> []string{"username", "password"} - []string{"", "type", "imsi", "auth", "password"}

WithParentClassName 2 ("imei" + \n + password value)
password: false <nil> []string{"password", "username"} - []string{"", "type", "imei", "auth", "password"}

[#/components/schemas/remote-networks]
WithParentClassName 1 ("bgp_peer" + \n + secret value)
secret: false <nil> []string{"peer_ip_address", "local_ip_address", "secret"} - []string{"", "protocol", "bgp_peer", "secret"}

ref remote-networks-protocol-bgp ???
secret: false <nil> []string{"originate_default_route", "peer_ip_address", "summarize_mobile_user_routes", "do_not_export_routes", "enable", "secret", "peering_type", "peer_as", "local_ip_address"} - []string{"", "protocol", "bgp", "secret"}

ref remote-networks-protocol-bgp ???
secret: false <nil> []string{"peering_type", "enable", "peer_ip_address", "peer_as", "originate_default_route", "local_ip_address", "secret", "do_not_export_routes", "summarize_mobile_user_routes"} - []string{"", "ecmp_tunnels", "_inline", "protocol", "bgp", "secret"}

[#/components/schemas/remote-networks-protocol-bgp]
WithParam local_ip_address
secret: false <nil> []string{"local_ip_address", "peer_ip_address", "summarize_mobile_user_routes", "do_not_export_routes", "peer_as", "secret", "peering_type", "enable", "originate_default_route"} - []string{"", "secret"}

[#/components/schemas/tacacs-server-profiles]
WithParam "name" (name value + \n + secret value)
secret: true <nil> []string{"address", "port", "secret", "name"} - []string{"", "server", "_inline", "secret"}

[#/components/schemas/ldap-server-profiles]
Solo
bind_password: false <nil> []string{"retry_interval", "id", "server", "timelimit", "bind_timelimit", "ldap_type", "base", "verify_server_certificate", "bind_dn", "bind_password", "ssl"} - []string{"", "bind_password"}

[#/components/schemas/local-users]
Solo
password: true <nil> []string{"id", "name", "password", "disabled"} - []string{"", "password"}

[#/components/schemas/scep-profiles]
Solo
password: false <nil> []string{"otp_server_url", "username", "password"} - []string{"", "scep_challenge", "dynamic", "password"}

[#/components/schemas/radius-server-profiles]
WithParamValue "name"
secret: true <nil> []string{"ip_address", "port", "secret", "name"} - []string{"", "server", "_inline", "secret"}
*/
