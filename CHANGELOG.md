## [2.0.9](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.8...v2.0.9) (2026-02-10)

### Bug Fixes

- **terraform:** allow empty location objects when schema has no attributes ([#700](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/700)) ([82c2549](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/82c25493af56f4ddcd74ff47c41ccaaf56994ea5))

## [2.0.8](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.7...v2.0.8) (2026-01-20)

### Bug Fixes

- **codegen:** Generate correct encrypted values handling code for variants ([#688](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/688)) ([da168bd](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/da168bdb1bdcd35576943c5cc5e189c4088f8cd2))
- **terraform:** Fix the generate_import_id function to better handle inline objects ([#692](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/692)) ([c16178e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/c16178e5ffa28cded3df1fa448412d19302d0d23))
- **terraform:** Handle attributes transitioning from plaintext to encrypted ([#689](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/689)) ([a342eeb](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/a342eeb624f209b54cdf2591392a7c4e124fd4cf))
- **terraform:** Update virtual-router spec and mark secret as hashed, and remove default values from variants ([#696](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/696)) ([b5815c8](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b5815c8b933894d979658bd739ee345720b579b0))

### Features

- **codegen:** Add support for custom validation and action attributes ([#670](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/670)) ([feda3f3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/feda3f3e841148a91913545d02bf07947ccd196d))
- **specs:** Add spec and examples for panos_push_to_devices action ([#671](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/671)) ([d8ce4d0](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d8ce4d09801f2c9e7ac77124ef7b5399aca6dcd3))
- **specs:** Expand commit action with additional attributes ([#672](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/672)) ([41dbb71](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/41dbb71529a7ea8d16dbb07e3c9cb7cd1bf12c01))

## [2.0.7](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.6...v2.0.7) (2025-11-25)

### Bug Fixes

- **specs:** Add missing ngfw location for panos_ipsec_tunnel ([#613](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/613)) ([f0d5ee7](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f0d5ee7ab5feefed68e16b717ce9513c39686fb3))
- **specs:** Change schema for template stack devices ([#677](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/677)) ([3153e8b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3153e8b2956d1b136449c9eb7979a7327f0c9abb))

### Features

- **codegen:** Add support for terraform actions ([#609](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/609)) ([0987b78](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0987b78dc2ba59402906dcde2b69ed47da0f8c61))
- **specs:** Add spec, tests and examples for panos_bgp_address_family_routing_profile ([#623](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/623)) ([f797549](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f797549cf3cefca30c299cbc396d7f08a81a57d0))
- **specs:** Add spec, tests and examples for panos_bgp_dampening_routing_profile ([#625](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/625)) ([bb55a52](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/bb55a528870d9de10604837d0cee4e7f48e52e1e))
- **specs:** Add spec, tests and examples for panos_bgp_redistribution_routing_profile ([#633](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/633)) ([446a02b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/446a02bcc9a93bc2f00f90a63b158a32e69cb27e))
- **specs:** Add spec, tests and examples for panos_bgp_timer_routing_profile ([#624](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/624)) ([9190f6f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9190f6fb696ec93482c523f5101552c81a621578))
- **specs:** Add spec, tests and examples for panos_filters_access_list_routing_profile ([#637](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/637)) ([63740ee](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/63740ee1b89f0ebfbaaab60f90716a2704148758))
- **specs:** Add spec, tests and examples for panos_filters_as_path_access_list_routing_profile ([#627](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/627)) ([0ecbfbc](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0ecbfbcccf4fc33b538289d69c70fe8c8b117719))
- **specs:** Add spec, tests and examples for panos_filters_community_list_routing_profile ([#626](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/626)) ([b51aa26](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b51aa267abe0534a3e92a9d27da212fc051a90db))
- **specs:** Add spec, tests and examples for panos_filters_prefix_list_routing_profile ([#636](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/636)) ([964850b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/964850beea08ddce4e1bf359e314716f902b9d8b))
- **specs:** Add spec, tests and examples for panos_filters_route_map_redistribution_routing_profile ([#630](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/630)) ([3babadf](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3babadf8a4f261649dc5a98d108a89469d7c83cb))
- **specs:** Add test, specs and examples for panos_bgp_filtering_routing_profile ([#631](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/631)) ([abe69ea](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/abe69ea90644c3df2a1755f933c4f161edb299f3))
- **specs:** Add test, specs and examples for panos_filters_bgp_route_map_routing_profile ([#647](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/647)) ([36fdd6a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/36fdd6a4b161300e4be6d347849765e9b74ad2cc))
- **specs:** Create spec, tests and examples for panos_commit action ([#610](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/610)) ([eb6ebe7](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/eb6ebe75991662f8018ebd468beb303628a510c9))

## [2.0.6](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.5...v2.0.6) (2025-11-19)

### Bug Fixes

- **codegen:** Fix XML marshalling code for panos_ssl_decrypt resource ([#639](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/639)) ([05e05ba](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/05e05ba1aac3fd668391792f88085a271aaa91b5))
- **codegen:** Generate correct terraform code for int64 lists ([#548](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/548)) ([a795dac](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/a795dac927e7d69e05ff4748350fd7c7e82b8670))
- **codegen:** Make audit comment code generation optional based on location ([#559](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/559)) ([3717c25](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3717c259ff44eb3545e7b5e15b089a30f3d75a84))
- **codegen:** support multiple encrypted values within a single object ([#534](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/534)) ([0d411f5](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0d411f53dc21d0d0c44c5854dc0d9d43c70b33e6))
- **specs:** fix locations for ssl-tls-profile ([#550](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/550)) ([1a93eb0](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1a93eb0f078d397b345aa111c2f72af6dd4fa3a6))
- **specs:** Mark IKE Gateway pre shared key as hashed value ([#554](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/554)) ([7f10bc6](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/7f10bc6217597d1f5ad295e1d99e5bec77c988de))
- **terraform:** Handle position attribute within lifecycle.ignore_changes ([#552](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/552)) ([07739bf](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/07739bf739a5afb09ffb0910b1c61557b86b203a))
- **terraform:** handle update of security rules when some entries are missing ([#555](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/555)) ([b498cd5](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b498cd589a6639c6cbe0943682f8d4216f3a5b89))
- **terraform:** Make panos_addresses only manage addresses from the state ([#571](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/571)) ([5a48863](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/5a4886390ff5324d9252ad5f0ce68f6ddff89dd1))
- **terraform:** properly handle object renaming ([#570](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/570)) ([7282002](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/728200251d59a50df62bac5d03816db74052ea70))
- **terraform:** skip validation of position attributes when not known at validation ([#527](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/527)) ([bedcf3c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/bedcf3c5f1fe54646af874283d4825d69464af38))
- **tests:** use template stack location for template variable meant for stack ([#553](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/553)) ([02b8324](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/02b8324dc95b700d6425fc848cd9d54a249ee344))

### Features

- **codegen:** Add support for disabling of variant group validation ([#536](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/536)) ([ac6816e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ac6816e40655c145bd534c43a3911f2aa3be99d6))
- **codegen:** Configure codegen log level and hide most output behind debug level ([#538](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/538)) ([04d0a1a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/04d0a1a82af63b3692234e02f45f2c4bf4f1176d))
- **codegen:** pass sdk client to custom hashing functions ([#546](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/546)) ([b715637](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b7156375e8bdae4c6ab92eac5836d9f6abc8626f))
- **codegen:** Support overriding of terraform parameter optional value ([#547](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/547)) ([4581bea](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/4581bea49417c85730a4b0d828aaceca1f4c1487))
- **codegen:** Use generic custom code dispatcher ([#608](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/608)) ([63e904e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/63e904e4bcfa1cd8c208d17ded3b323322f64a3c))
- **codegen:** use terraform types for structures and lists ([#531](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/531)) ([a25d30c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/a25d30cba2c02747799a362a5582048f075570df))
- **gosdk:** Enable optional Basic access authentication via headers ([#551](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/551)) ([5ea1d63](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/5ea1d634f9f21cf11e1b609246d4c044e556a943))
- **specs:** Add spec, tests and example for certificate import ([#495](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/495)) ([76c1d8a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/76c1d8a7a4f0834688396720ae101cc84caa6c21))
- **specs:** Add spec, tests and examples for authentication profiles ([#481](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/481)) ([0dcbfb1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0dcbfb1f5986d600702abc9431f6354b778d536c))
- **specs:** Add spec, tests and examples for globalprotect-gateway ([#549](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/549)) ([ab8055c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ab8055cfdf6b312cc7e65d2ef80e119c58dbc150))
- **specs:** Add spec, tests and examples for globalprotect-portal ([#535](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/535)) ([2ba9f0e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/2ba9f0e076fb83922490cdb196b637680c6d1bfd))
- **specs:** Add spec, tests and examples for panos_correlation_log_settings ([#563](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/563)) ([0aeb329](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0aeb329af9e36c41a0b2cfe78ca0606e33c6300d))
- **specs:** Add spec, tests and examples for panos_proxy_settings ([#557](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/557)) ([dba8100](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/dba81006ea109c78c681c2f76626cf67dd4b8808))
- **specs:** Create spec, tests and examples for panos_config_log_settings ([#564](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/564)) ([027bd84](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/027bd84209c9a12c6bd75a0210532b683a66af6d))
- **specs:** Create spec, tests and examples for panos_default_security_policy ([#560](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/560)) ([424641e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/424641e27b2244c0795b910cc788c9c5d04e28bf))
- **specs:** Create spec, tests and examples for panos_globalprotect_log_settings ([#566](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/566)) ([200a7cb](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/200a7cb2afa84f45a2b9d7355f8465a36df084bc))
- **specs:** Create spec, tests and examples for panos_hipmatch_log_settings ([#567](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/567)) ([fc0b0e3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/fc0b0e39c8c6dd0611f83c531ba90bc1ce43910f))
- **specs:** Create spec, tests and examples for panos_iptag_log_settings ([#569](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/569)) ([499fe8a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/499fe8aeda2e15f836f9d63deeabbce54fca55ab))
- **specs:** Create spec, tests and examples for panos_syslog_profile ([#565](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/565)) ([24ad69b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/24ad69b993c008d25d8f34b359184a377e0b0ae3))
- **specs:** Create spec, tests and examples for panos_system_log_settings ([#561](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/561)) ([1545bb4](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1545bb499f5c4e1daa7062ee56dccce2fb82ff8e))
- **specs:** Create spec, tests and examples for panos_userid_log_settings ([#562](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/562)) ([b1d4939](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b1d49390ba3650fc21ea4b250e52dc362eba3dc6))
- **specs:** New spec, tests and examples for panos_bgp_auth_routing_profile ([#621](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/621)) ([aaa0197](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/aaa0197920a97a6aced57aa93325d4f7ee4f4d15))

## [2.0.5](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.4...v2.0.5) (2025-09-03)

### Bug Fixes

- **terraform:** Handle position attribute within lifecycle.ignore_changes ([8351af4](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/8351af4))
- **terraform:** properly handle update of security rules when some entries are missing from the server ([3813a6f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3813a6f))

## [2.0.4](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.3...v2.0.4) (2025-07-24)

### Bug Fixes

- **specs:** update specs that use device-group to mark location filter variable ([#544](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/544)) ([69dfcce](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/69dfcce1f10f3c030bf3ac7333f7c291b89971e5))
- **terraform:** Handle device group hierarchy location when listing entries ([#545](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/545)) ([68a574e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/68a574e55ac30bd3c65fac45a92d5012df1ba1a5))

### Features

- **codegen:** Add MiscAttributes field to gather unprocessed XML attributes ([#543](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/543)) ([4af0e79](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/4af0e79ef2747891745573bdd567f7811e67a854))

## [2.0.3](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.2...v2.0.3) (2025-07-24)

### Bug Fixes

- **specs:** add missing ngfw location to logical router ([#542](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/542)) ([27b249a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/27b249ab8c3a7dfd1729818b536a23a7fd97e12d))
- **specs:** update panos xpath for certificate profiles ([#533](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/533)) ([773edee](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/773edee111337db3e53a3c999dbe8825628df23a))
- **terraform:** change nat policy variant check between floating ip and ip ([#541](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/541)) ([de32b33](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/de32b3339da05c02de3108a25e2abceba4dfc88f))

### Features

- **specs:** Add spec, tests and example for dhcp ([#497](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/497)) ([ee10b98](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ee10b98a1cfc047d92ba8ab82f8c4733405e8d53))

## [2.0.2](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.1...v2.0.2) (2025-06-06)

### Bug Fixes

- **codegen:** Fix location xpath components for panos_xpath without variables ([#502](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/502)) ([f55cc80](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f55cc80eedb69a25b64dd2d43520944858a9333b))
- **codegen:** Improve import logic for datasource-only specs ([#493](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/493)) ([509b8b9](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/509b8b9c62a0ba45f6ae74aeafb97a4da42a56fe))
- **gosdk:** change signature of ImportFile to accept content as []byte and set Content-Type during ExportFile ([#500](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/500)) ([bc8fec1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/bc8fec1be9d6fe37790cdcded965d6f45b17de14))
- rename PANOS_HOST to PANOS_HOSTNAME ([#530](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/530)) ([5c5b9fd](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/5c5b9fd7652d8eff12d9c08d53477061306578c0))
- **specs:** Add missing template-stack location to template variables ([#529](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/529)) ([f334922](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f3349228dec1861e8c87995a82df00a08e2e5679))
- **specs:** Add missing vsys location for external dynamic lists ([#496](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/496)) ([9f9d9e3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9f9d9e37e8ae740e68e84ee496fd2de01c3e3e3f))
- **terraform:** delete config resources by updating them with an empty object ([#507](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/507)) ([dd23445](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/dd234450cc5ec199227dba915e6d68f6b152483c))
- **terraform:** fix type casting panic within ephemeral resources ([#528](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/528)) ([21fe494](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/21fe4944195822dd5cda7f9d2ce228545755d18d))
- **terraform:** Make sure we only import resources that have a valid import location ([#505](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/505)) ([5715136](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/57151366fbd1abbb1f551bc75b9dde202e271185))

### Features

- **codegen:** Add support for child specs ([#485](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/485)) ([812760b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/812760b84f35e029b0d8c8f9d172ab9b3cde7f9a))
- **codegen:** Add support for int64 member lists ([#498](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/498)) ([9b82894](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9b8289446396b4951d177afd9a91137e85ddfb82))
- **codegen:** Add support for more hashing types and hide them within private state ([#520](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/520)) ([62756e3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/62756e38a17f4e7606d7e42ebe14b83c9416fd46))
- **codegen:** Add support for overriding gosdk parameter names ([#521](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/521)) ([b5b63c3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b5b63c32362cb29f0da62135feae419206182d72))
- **codegen:** Parametrize which SDK CRUD methods are to be implemented ([#492](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/492)) ([fedcc1b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/fedcc1b87f1995c8f8379eb539d8855bd5c70d05))
- **codegen:** Render intermediate XML structures to support empty variant lists ([#486](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/486)) ([7ec4f0a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/7ec4f0a9286b56445f477f02135219120529ac2f)), closes [golang/go#7233](https://github.com/golang/go/issues/7233)
- **codegen:** Rewrite SDK codegen for structures, normalizers and specifiers ([#483](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/483)) ([30a16e9](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/30a16e9a9702a10a19a95b1b8234858b7f9bfc02))
- **gosdk:** Add global locks manager to serialize some operations ([#494](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/494)) ([b7ba00b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b7ba00b91efa0f7eb297a94a40ab3e6c657e769c))
- **specs:** Add codegen spec and acceptance tests for ethernet layer3 subinterface ([#488](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/488)) ([3d33c12](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3d33c12b8e61d27b584de0d3e97550d917db5d67))
- **specs:** Add codegen specs and acceptance tests for virtual router static routes resources ([#487](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/487)) ([0f58efb](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0f58efbfa4aabaa19a23ff2f8450564445bb77b4))
- **specs:** add spec for read-only device certificates ([#501](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/501)) ([09d538c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/09d538c62776a8291119f2e866c465cc1b918b3a))
- **specs:** Add spec, tests and examples for aggregate layer3 subinterfaces ([#491](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/491)) ([b0f1b9a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b0f1b9ad8e10e97ecd855fd713cef6293fafaf1b))
- **specs:** add spec, tests and examples for device general settings ([#504](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/504)) ([37c290d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/37c290d3ca833d6eda6ee60121551f5004218ff1))
- **specs:** Unify locations for dns and ntp specs ([#503](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/503)) ([44d4337](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/44d4337f56a934371b5c3ab9e7a11c1e36243a28))

## [2.0.1](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/v2.0.0...v2.0.1) (2025-05-20)

### Bug Fixes

- **ike-gateway:** Support local_address with only interface reference ([#467](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/467)) ([d858c2a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d858c2a7d0c532b625e9cf7dd2784bd41a4f5058))
- **specs,codegen:** Add template import location for tunnel interfaces ([#474](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/474)) ([66c4e06](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/66c4e0687ab14448893abfefbe711ee31764292b))
- **specs:** Add missing shared location for ldap profile ([#489](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/489)) ([50b791e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/50b791e6708cef64f5aa73cce77fbf964ee8ef73))
- **specs:** Add missing xpath_prefix component for antivirus profile ([#482](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/482)) ([6530e75](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6530e75972765550f9129c58aaea41b9aa65b81a))
- **terraform:** Improve handling of movement edge cases around non-exhaustive resources ([#470](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/470)) ([48b7825](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/48b7825776bfbef48d11165cebb8e82bcc31dc15)), closes [PaloAltoNetworks/terraform-provider-panos#465](https://github.com/PaloAltoNetworks/terraform-provider-panos/issues/465)
- **terraform:** Only validate position when attribute is known ([#472](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/472)) ([2fac253](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/2fac25347f3d2df5fcef71a1aa90d84ce692b41c))
- **terraform:** Use terraform types for position and location struct members ([#469](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/469)) ([777f477](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/777f47700a0255db060f001ea023c8fb23c89b70))

### Features

- **codegen:** Add support for multiple variant groups within single spec ([#468](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/468)) ([35f01f6](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/35f01f65b6285914f2ba77599d3be384324d9577))
- **specs:** Add initial spec for panos_ldap_profile resource ([#480](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/480)) ([9aacaf3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9aacaf372b28c6a8a8f2c029ac5ed400292cd2fb))
- **specs:** Initial network/logical-router yaml spec ([#471](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/471)) ([d1780cd](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d1780cda29b12c20ed3cbdc2abf7479a8ff4b42d))
- **tests:** Add more tests for position as variable ([#473](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/473)) ([d580d65](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d580d65b29722953031b3d586b480fa7a4a185ae))

# [2.0.0](https://github.com/PaloAltoNetworks/pan-os-codegen/compare/f19e57b806bc816bbfede26e85e9e45a83d52f09...v2.0.0) (2025-04-10)

### Bug Fixes

- **assets/pango/filtering:** Update tests for empty filter case ([#184](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/184)) ([09c17b1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/09c17b175133a883bbf3f1d292c1c579e1482c97))
- **client:** Renames previous occurencies of plugin field of pango client ([#175](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/175)) ([ade707c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ade707ca8ef99140293e837700cf2b616f97c706))
- Default log level handling in the client ([#117](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/117)) ([cdcfe25](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/cdcfe256da1b83c73979316a89442b0307948681))
- Fix provider overwriting parts of the resource that is not managed by it ([#145](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/145)) ([30148fc](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/30148fc1dbd3d57dd1f7e9ccb71ca58743b0718c))
- Fixed env var name for api key authentication ([#163](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/163)) ([6e9d071](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6e9d071f3578405cf1d1ae62bdfea569636b43aa))
- generate usable pango example ([#114](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/114)) ([c7f6cb3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/c7f6cb349711442a113ba0bc52395e46e88dab81))
- Make generated code compile with go 1.22 ([#139](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/139)) ([f6fdfe0](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f6fdfe019b7e82ac3b1b5cc333f097149dc579f9))
- Make sure that entries created during CreateMany() call are in correct order ([#279](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/279)) ([3fcc8a6](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3fcc8a63332c16387190e2bb7eb4f0d6299b5e1b))
- properly delete and update resources during UpdateMany() ([#290](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/290)) ([e7fd331](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/e7fd331952e1d23752b17a3c9762dbc36341fdbc))
- Properly handle empty plural resources ([#291](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/291)) ([9c75c87](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9c75c8715b700f2d468a5a17cbb8d065d37e9117))
- Properly initialize an expected order of entries when generating movement actions ([#449](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/449)) ([bef98af](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/bef98af043ffae9621dc9aa1748802ffd25d3134))
- Properly pass updated position from Terraform to SDK ([#446](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/446)) ([ec01122](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ec011225e30b114c000735589d1e37daf3e2c16f))
- **specs/schema:** Add missing name property in terraform overrides ([#181](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/181)) ([1f4e1d5](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1f4e1d5065f2bd44c6d8cdac4d75193d8706d79a))
- **specs/security-policy-rule:** Drop unused locations variable ([#174](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/174)) ([3a69e53](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3a69e530c7d410c9252fdc387483959d46babce9))
- **specs:** Fix panos_external_dynamic_policy xpath and mark some attributes as encrypted ([#432](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/432)) ([875cf48](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/875cf4892b49d8c3b57920367035d4bae297b271))
- **specs:** Make security policy destination zones a list ([#417](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/417)) ([cdd5af3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/cdd5af3a0f58215c0e34fd9f5be6b98015226ea3))
- **specs:** Remove default value from url_filtering_sec_profile variants ([#419](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/419)) ([f28ddf3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f28ddf3e2f05ea638a78bbc2474414b4775033b8))
- **specs:** Update antivirus profile path ([#285](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/285)) ([d470c67](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d470c67663e0d589758392063a3bcb55e3c1b5e8))
- **specs:** Update file blocking security profile ([#284](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/284)) ([128617a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/128617aff327d613c836e25cd57e0e451559fa80))
- Support rendering of spec lists as unordered sets ([#431](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/431)) ([9ba0d5a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9ba0d5acf4d4c1868949576acbd3414b94909e7e))
- **tests:** Establish terraform dependencies between resources ([#281](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/281)) ([fb89184](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/fb89184d8e61f6342527af16238bf86eda0b7ecd))
- **tests:** Fix panos_ipsec_crypto_profile acceptance tests ([#282](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/282)) ([e2afe57](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/e2afe57ff0a42d7cb15fed98811e9dfdc0fe1c3e))
- **tests:** Fix panos_tunnel_interface acceptance tests ([#283](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/283)) ([2b2ab8e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/2b2ab8ed75e26ff28a61870a3dd9a5b4339c5146))
- Update pango example to match with current SDK ([#178](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/178)) ([8af3034](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/8af303465787cc068ce047c5ffafda4ccb636ed8))
- Update resource names ([#162](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/162)) ([ad3ec23](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ad3ec239584aaf334685ab1804f1ebc7c5250f48))

### Features

- **acc_test:** Adds acceptance testing primitives ([#158](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/158)) ([73cfe3f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/73cfe3faf141e1aa9d43622f07ea892b6462c465))
- Add full rendering for terraform resources and data sources ([#106](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/106)) ([70807b2](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/70807b22fd1c4ff1176895da2cd60108fe276a61)), closes [#118](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/118) [#120](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/120)
- add golangci-lint configuration with pre-commit hooks to lint code ([#116](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/116)) ([96509aa](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/96509aad902f37a42e11f68cd27f3e19481e9112))
- Add import support for ethernet and add additional specs ([#130](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/130)) ([dfa889a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/dfa889a82156db9c95f8d2b4e6c0a13a7253a45a)), closes [#135](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/135) [#136](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/136)
- Add NAT policy specification ([#173](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/173)) ([dc9e7af](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/dc9e7affe5b79673ec1683bccd794144b87431af))
- Add new functions for entries with UUID, fix problem with `log-end` in security policy rule, remove custom MarshalXML for config ([#63](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/63)) ([9cae942](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9cae942386eedcfc1dc6bcd133b3d5f82f9f7ab2))
- Add schema validation for variants ([#152](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/152)) ([8dc803f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/8dc803f088f3de8a60ed417fd4feba22c657bff2))
- Add support for ephemeral resources and implement panos_api_key ([#422](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/422)) ([eb0a4c9](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/eb0a4c9746d498eabe503ba54a86718550669e02))
- Add support for importing plural resources ([#444](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/444)) ([696d894](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/696d894118040d0ea6eca7d7f261b5ab78af0605))
- Add support for importing resources with positions ([#458](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/458)) ([6577dbe](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6577dbe4b661ab081880e00b699ae8fda138800b))
- Adjustment of types and utilising default values in specs ([#108](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/108)) ([2fe98de](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/2fe98debd49a8d326b096f8222fb100cd51b3b5c))
- Basic VR and Ethernet interface ([#96](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/96)) ([cc09bc2](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/cc09bc21ab7eca400bbd2c114acbcc6a561760e0))
- **certificate-profile:** Initial panos_certificate_profile codegen spec ([#423](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/423)) ([3840ccf](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3840ccff245b9612d640f15ed1a0c3723dd5eafd))
- **codegen:** Make code generation reproducible ([#453](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/453)) ([8a7ac6d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/8a7ac6d1ef863c2c531f7be51a85b386ecb0dc99))
- Copy static files (assets) ([#29](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/29)) ([ab72553](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ab725538b656a9fd50a81230504f25e2f20948e6))
- create address_value terraform provider function ([#154](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/154)) ([ef492d8](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ef492d8550c93d8ba0ce55d237f7a96424baf4c9))
- Defer retrieval of system info and plugin data ([#169](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/169)) ([744f93d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/744f93d172f2422f59d40b8e65d71302ba50f8b2))
- **docs:** Add more self-contained examples for resources ([#448](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/448)) ([3410ab8](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/3410ab8ef70cdabeae9eb6506dc38fc46e983007))
- E2E example with static files required to execute it ([#46](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/46)) ([f4e4c00](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f4e4c00b3d0a68708d4a650db55b5da46228bc66))
- Extend existing specs and example to configure resources in Panorama templates and device groups ([#100](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/100)) ([357b401](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/357b401da8c7f9e07663f7859202025576de325a))
- Extend locations for specs ([#103](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/103)) ([6c02349](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6c02349c54404e801c7aa354c72d9298b68497d0))
- Extend PR pipeline to generate files in GitHub Actions ([#36](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/36)) ([4e52c36](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/4e52c36e6a0e5242f450c2f140cbebd99a3efd8d))
- Extend service template ([#50](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/50)) ([d5efaa8](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d5efaa8e7bcf760965f64523d7fa6212b52d0516))
- Extend support for device groups specification ([#137](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/137)) ([1b483a1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1b483a1d6a4dddd62006d2454f6b1c10758640a0))
- Generate descriptions for "entry names", lists and maps ([#147](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/147)) ([ed51a4b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ed51a4b756db93a9f7eaadfa2a968255545a97a2))
- Introduce ephemeral_auth_key resource ([#438](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/438)) ([e4a2d8d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/e4a2d8db1b0afeb841cf8f6258beb21a0e1dca0e))
- Introduce log and refactor logic of cmd ([#18](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/18)) ([ec26a4a](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ec26a4aff070a8c8a575d8739f17f58d724ad7a3))
- Introduce slog based logging interface ([#113](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/113)) ([1ea8a89](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1ea8a898ad92305e06937c67972a7f5426b8c362))
- Logging (e.g. raw XML request, XML reply from PAN-OS API) ([#105](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/105)) ([630cb8d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/630cb8df8d06eb619460669fd1510d1178d4e438))
- Make MultiConfig batch size configurable ([#460](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/460)) ([634d161](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/634d16186992378433d5a3bce05cabc18ea1b728))
- More updates to the Ethernet interface Layer3 spec ([#138](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/138)) ([ad8675f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ad8675f4204fb8ae69e306ec9f1225a1a84c211e))
- New movement implementation ([#123](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/123)) ([1a0acf5](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1a0acf5bbbbafcbff01b81c7eca90fb35b155083))
- Normalised specifications of configuration settings ([#14](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/14)) ([f19e57b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f19e57b806bc816bbfede26e85e9e45a83d52f09))
- Parse normalised specifications ([#17](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/17)) ([813a4d9](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/813a4d9e7e1658a00119b4dcd779df204c9c4e1a))
- Refactor service template add new functions supporting UUID ([#53](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/53)) ([81cd33f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/81cd33fe56298d0e2fa6f19d1dc7695075ce83a7))
- refactor to apply dry ([#33](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/33)) ([5a18280](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/5a182806638ec4be89842007adf641a0edfe3431))
- Render `config.go` (first stage) and improvements in rendering `entry.go` ([#35](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/35)) ([49457a5](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/49457a5e760b853d73cf1f06a4821a9aef4796b9))
- Render `entry.go` (first stage) ([#31](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/31)) ([cdb02a7](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/cdb02a7f3eb0d14f2da3c99709948f521119869e))
- Render `entry.go` (second stage) ([#34](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/34)) ([25ec702](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/25ec702c225797d36e1ce671b0820f9b4dd0c33f))
- Render `location.go` ([#25](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/25)) ([65727a4](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/65727a4e25155aebbf8b025366224ba5b6b6ea89))
- Render `service.go` for entry and config ([#42](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/42)) ([ba3131c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ba3131c8776180e3ee99fd3147fc1343bd6a22bf))
- Render files from templates (first step) ([#21](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/21)) ([b40e3b4](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b40e3b44bc22f650d2d123f0e6c2f7f9800abed2))
- render provider file skeleton ([#64](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/64)) ([6befe3b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6befe3bb7b80a44d33905129879e5f05b676898c)), closes [#58](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/58)
- render resource first part ([#99](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/99)) ([f43efad](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f43efad250bdd3694897b2c5cd3e10edab0aab10))
- Replace hand-written specs with autogenerated specs ([#186](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/186)) ([761708c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/761708c3f05d4744281c5a55c683d95910bda9d6))
- Shared Terraform Provider CRUD functions ([#140](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/140)) ([b8f07e1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b8f07e18431861612c50a0a4e43f2972b41fc031))
- Spec files (mgmt profile, zone, loopback) and improvements in rendering for params with embedded entry ([#94](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/94)) ([7b9a656](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/7b9a656bb45c099d70d3bc19e8cc6e37764df107))
- **spec:** Adds SSL/TLS service profile codegen specs ([#425](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/425)) ([051e50d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/051e50d09a91bc3fed6a623df2923542528036cb))
- **specs:** Adds SSL decrypt configuration specs ([#421](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/421)) ([b5a288e](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b5a288e575b7d2eee65daae6747e780a386311d0))
- **specs:** batch update yaml specs to pull xml schema parser changes ([#441](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/441)) ([b082b83](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b082b831c8d02a3a776347fd1b63606d4a0cfa1c))
- **specs:** Generate singular variant for panos_address resource ([#426](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/426)) ([4e30007](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/4e300077c35c57782eba1f0d8bd43560d8b522ef))
- **specs:** Initial panos_aggregate_interface codegen spec ([#251](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/251)) ([723ca76](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/723ca76575967ed268caf43370f9690eb17eec79))
- **specs:** Initial panos_anti_spyware_security_profile codegen spec ([#210](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/210)) ([e8a350b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/e8a350be282b3c80c3de498b986df10fb35e260a))
- **specs:** Initial panos_antivirus_profile codegen spec ([#182](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/182)) ([e80c499](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/e80c499eb9e97d0ebfc35b74bab152445eb6f7c9))
- **specs:** Initial panos_application codegen spec ([#377](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/377)) ([c73632c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/c73632c22a4c1302dc6965cf9689d67c6850ed1f))
- **specs:** Initial panos_application_group codegen spec ([#378](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/378)) ([6414d88](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6414d88825184836f38e224cdf15d5857ac3bc8f))
- **specs:** Initial panos_dynamic_updates codegen spec ([#261](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/261)) ([2a7ddc9](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/2a7ddc97342f2351345ce60f34e87bdca3bc4336))
- **specs:** Initial panos_ethernet_layer3_subinterface spec ([#277](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/277)) ([5ccd980](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/5ccd9808bac7e5b7853fa011d9e0f8bd7d539c0f))
- **specs:** Initial panos_file_blocking_profile codegen spec ([#241](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/241)) ([af62ae2](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/af62ae2bf467f92c858442a58477a2e77c38f3b6))
- **specs:** Initial panos_ike_crypto_profile codegen spec ([#255](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/255)) ([04e3578](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/04e3578b98ea236adcb8c900d3323d4d68418a48))
- **specs:** Initial panos_ike_gateway codegen spec ([#259](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/259)) ([b0538d1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b0538d1aeae90389a50784f9e8c2ce94211d5708))
- **specs:** Initial panos_ipsec_crypto_profile codegen spec ([#257](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/257)) ([4993359](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/4993359dc4f50c565913404b049832fbc30e1bac))
- **specs:** Initial panos_log_forwarding codegen spec ([#235](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/235)) ([478edb3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/478edb3a29930546fc818481f8c6b77a1d6e1111))
- **specs:** Initial panos_loopback_interface codegen spec ([#253](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/253)) ([bd45d6d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/bd45d6dfdda748c26001d9f7e78c1d9d694f562f))
- **specs:** Initial panos_security_profile_group codegen spec ([#236](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/236)) ([8c2ad93](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/8c2ad93f53110d8878786f7fca23f3fddaf2b8ba))
- **specs:** Initial panos_tunnel_interface codegen spec ([#249](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/249)) ([61b60b6](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/61b60b6021ed8987eb5114cebffa34b02cfb1be2))
- **specs:** Initial panos_url_filtering_security_profile codegen spec ([#243](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/243)) ([f6770f6](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/f6770f60c5b8fa40390a8aeb8bde6c7ee247e9e8))
- **specs:** Initial panos_vlan codegen spec ([#269](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/269)) ([63e8887](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/63e88878774f8b4e59c0e71ea118dcd54a904b75))
- **specs:** Initial panos_vlan_interface codegen spec ([#247](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/247)) ([8ca07de](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/8ca07dede821a83c73fe3ecc734fa66e02f6c36b))
- **specs:** Initial panos_vulnerability_security_profile codegen spec ([#245](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/245)) ([32f925b](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/32f925b5577b0f51a30e85be545fe8af5a97429e))
- **specs:** Initial panos_wildfire_analysis_profile codegen spec ([#239](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/239)) ([ef48836](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ef488362e9701ae545604c83542e417a305c814c))
- **specs:** Initial resource_admin_role spec ([#234](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/234)) ([dba168c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/dba168ca230359d9acdf5d9e81b8d4c8c20a8655))
- **specs:** IPSec Tunnels - Codegen spec ([#199](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/199)) ([7313292](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/73132926513a7b96d8fc71a131ab0102a358c5e7))
- **specs:** Mark policy resource uuid attributes as private ([#445](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/445)) ([478b674](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/478b674280724aeb1a6af4dad7928518f9d3debc))
- Support for tag colors ([#47](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/47)) ([d7d0d9f](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/d7d0d9f37861cdfe99f1da73f0de778c0e3d65a2))
- Support importable resources ([#104](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/104)) ([02c1fa1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/02c1fa1d841009a81a727ff22a049b947e4cf6b1))
- **tests:** Add panos_application acceptance tests ([#379](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/379)) ([b277bad](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/b277badc53c3492e1eae90b4653ced219c70fb22))
- **tests:** Add panos_security_profile_group terraform tests ([#416](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/416)) ([0fcb612](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/0fcb61252c18b307c12e811bc1e057805b7161e4))
- **tests:** Add panos_url_filtering_security_profile terraform tests ([#418](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/418)) ([6feccd8](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/6feccd8f000e62e6e7195da42a8bf898ff2e8b67))
- **tests:** Add panos_vulnerability_security_profile terraform tests ([#420](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/420)) ([4e11ca0](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/4e11ca05cd7bd32a12990089350051357c6dfd17))
- **tests:** Initial panos_admin_role terraform acceptance tests ([#266](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/266)) ([edf193d](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/edf193dff5d5fadf0792fc64563410430e62aac6))
- **tests:** Initial panos_aggregate_interface acceptance tests ([#270](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/270)) ([1463f85](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/1463f8598b83b8ef4b884d427713f4caf420d032))
- **tests:** Initial panos_application_group terraform tests ([#415](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/415)) ([ef62677](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ef62677b92c9a159c786a4d167db2f22d453a800))
- **tests:** Initial panos_dynamic_update terraform acceptance tests ([#264](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/264)) ([ad237ba](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ad237ba1d513185f8a52739afe43fed3d3cb1ac2))
- **tests:** Initial panos_ike_crypto_profile acceptance tests ([#274](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/274)) ([cf6401c](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/cf6401c87ada8ee78501f3f00eb668ffbafdf405))
- **tests:** Initial panos_ipsec_crypto_profile acceptance tests ([#275](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/275)) ([7f04749](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/7f047492ae09dcdb34de57fe4674a0dfcd142d6f))
- **tests:** Initial panos_log_forwarding_profile terraform acceptance tests ([#265](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/265)) ([9f09f11](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/9f09f11e4e4185d12f99cc5216eaab63c9c6648c))
- **tests:** Initial panos_vlan_interface acceptance tests ([#273](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/273)) ([ac3f8a3](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/ac3f8a328ef21b8f931533ced8a7838cc82cfacd))
- **tests:** Initial panos_zone terraform tests ([#328](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/328)) ([98837c2](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/98837c236528ff3734ec20c3144dcf610372ef5b))
- Unify location types, making a shared location an (optionally) empty object ([#461](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/461)) ([caa16a1](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/caa16a1978e2af8820dec34f6a8174491a98ac77))
- Update golangci-lint used in CI to 1.60 ([#134](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/134)) ([373b796](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/373b796228ee7f0cf86e0f03e0215c495ef8729c))
- update schema specification with JSON schema validation ([#122](https://github.com/PaloAltoNetworks/pan-os-codegen/issues/122)) ([e3f2124](https://github.com/PaloAltoNetworks/pan-os-codegen/commit/e3f2124b5844435bef76887ee015e10387406ef5))
