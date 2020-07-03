# doc is wrong: https://www.terraform.io/docs/providers/infoblox/index.html
# use: https://github.com/terraform-providers/terraform-provider-infoblox/blob/328f438127673823fc574c7c9dfe75c82978b8bd/infoblox/provider.go
provider "infoblox" {
username="admin"
server="x.y.z.t" # DNS server DNS name or IP address (both tested ok), be careful to space
password="infoblox"
wapi_version=2.5 # change version to 2.5 and do not use default to 2.7 (as my Infoblox instance is a 2.5)
}

resource "infoblox_a_record" "demo_record" {
  vm_name="scoulomb-terraform"
  zone="test.loc"
  dns_view="default"
  ip_addr="42.42.42.42" //use the ip address used in IP allocation
  cidr="10.0.0.0/24"
  tenant_id="dummy"
  network_view_name="dummy"
}
# https://github.com/infobloxopen/terraform-provider-infoblox/issues/62

resource "infoblox_cname_record" "demo_cname"{
  canonical="demo.scoulomb.com"
  zone="test.loc"
  alias="demo1"
  tenant_id="test"
}
