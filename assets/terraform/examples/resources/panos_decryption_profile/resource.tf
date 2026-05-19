resource "panos_decryption_profile" "example" {
	location = {
		shared = {}
	}
	name        = "my-decryption-profile"
	description = "Example decryption profile"

	ssl_forward_proxy = {
		block_expired_certificate  = true
		block_untrusted_issuer     = true
		block_unknown_cert         = true
		block_unsupported_version  = true
		block_unsupported_cipher   = true
		restrict_cert_exts         = true
	}

	ssl_protocol_settings = {
		min_version        = "tls1-2"
		max_version        = "max"
		keyxchg_algo_rsa   = true
		keyxchg_algo_ecdhe = true
		enc_algo_aes_128_gcm = true
		enc_algo_aes_256_gcm = true
		auth_algo_sha256   = true
		auth_algo_sha384   = true
	}
}
