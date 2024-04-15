package terraform

const (
	EncryptedValuesUnderscoreName = "encrypted_values"
	EncryptedValuesCamelCaseName  = "EncryptedValues"
	EncryptedValueSchema          = `            "encrypted_values": rsschema.MapAttribute{
                Description: "(Internal use) Encrypted values returned from the API.",
                Computed: true,
                Sensitive: true,
                ElementType: types.StringType,
            },`
)
