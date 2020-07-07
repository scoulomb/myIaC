
provider "infoblox" {
username="admin"
server="" # DNS server DNS name or IP address (both tested ok), be careful to space
password="infoblox"
wapi_version=2.5 # change version to 2.5 and do not use default to 2.7 (as my Infoblox instance is a 2.5)
}

resource "infoblox_a_record" "demo_record" {
  vm_name="scoulomb-terraform2"
  zone="scoulomb.loc"
  dns_view="default.scoulomb-nw"
  ip_addr="" //use the ip address used in IP allocation
  cidr="10.0.0.0/24"
  tenant_id="dummy"
  network_view_name="scoulomb-nw"
}

resource "infoblox_cname_record" "demo_cname"{
  canonical="demo.scoulomb.com"
  zone="test.loc"
  alias="demo1"
  tenant_id="test"
}
