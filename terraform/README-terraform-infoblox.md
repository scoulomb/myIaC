# Terraform 

## Terraform Infoblox plugin

We had done something similar with Ansible in [myDNS](https://github.com/scoulomb/myDNS/tree/master/3-DNS-solution-providers/1-Infoblox/4-Ansible-API).

Examples are given here:
- https://github.com/infobloxopen/terraform-provider-infoblox/tree/master/examples/azurerm (it is not Azure DNS but Infoblox deployed on Azure)
- https://www.terraform.io/docs/providers/infoblox/r/a_record.html

We will create a A record and CName record with Terraform (note HostRecord creation seems not possible with Terraform).

### Configure the provider

To start the doc given on Terraform website for Terraform provider here:
https://www.terraform.io/docs/providers/infoblox/index.html

is not accurate, it should not be:

````shell script
provider "infoblox"{
INFOBLOX_USERNAME=infoblox
INFOBLOX_SERVER=10.0.0.1
INFOBLOX_PASSWORD=infoblox
}
````

but 

````shell script
provider "infoblox" {
username="admin"
server="mydns.loc" # DNS server DNS name
password="xxxx"
wapi_version=2.5 # change version to 2.5 and do not use default to 2.7 (as my Infoblox instance is a 2.5)
}
````
Otherwise we have the following error:

```shell script
[root@archlinux infoblox]# terraform apply
[...]
Error: WAPI request error: 400('400 Bad Request')
Contents:
Version 2.7 not supported


  on <input-prompt> line 1:
  (source code not available)

````

To find the issue, I had to check Infoblox plugin code [here](https://github.com/terraform-providers/terraform-provider-infoblox/blob/328f438127673823fc574c7c9dfe75c82978b8bd/infoblox/provider.go).


Similarly running A record creation with 2.7 lead same error:
https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#post-a
```shell script
curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
       -H "Content-Type: application/json" \
       -X GET \
       "https://$API_ENDPOINT/wapi/v2.7/record:a?name=test-infoblox-api-a1.test.loc
```
Output is 

````shell script
Version 2.7 not supported
````


To fix this we set `wapi_version=2.5`


### Configure NIOS

Then when creating a record we had error like 

````shell script
Error: Error creating CNAME Record : WAPI request error: 400('400 Bad Request')
Contents:
{ "Error": "AdmConProtoError: Unknown extensible attribute: VM Name",
  "code": "Client.Ibap.Proto",
  "text": "Unknown extensible attribute: VM Name"
}


  on main.tf line 13, in resource "infoblox_cname_record" "demo_cname":
  13: resource "infoblox_cname_record" "demo_cname"{

````

The reason of this is that I am not using a CNA (Cloud Network Automation) edition of Infoblox.
So we shoud add EA (extensible attributes) with infoblox UI.
Gui is avaialble at http://$API_SERVER/ui.

From there add EA wiht following steps.

1. Navigate to Administration -> Extensible Attributes in your Infoblox Grid Manager GUI.
2. Click on the + (Add) button.
3. Enter the name for the EA, as displayed in bold in the list above.
4. Set the Type dropdown menu to the required setting (refer to the list above).
5. Optional: Add a comment.
6. Click on the small arrow next to Save & Close and select Save & New to add additional EA’s, or click on Save & Close if don

Here are EA to add:

````shell script
If CNA licenses not present following default EA's should be added in NIOS side.
VM Name :: String Type
VM ID :: String Type
Tenant ID :: String Type
CMP Type :: String Type
Cloud API Owned :: List Type
Network Name :: String Type
````

We can refer to:
- https://github.com/infobloxopen/terraform-provider-infoblox/issues/62
- https://vmguru.com/2017/07/infoblox-vrealize-automation-infoblox-nios-setup/


Note we need to add the EA as the plugin forces us to send this attributes.

### And test the configuration

#### Creation

````shell script
cd /home/vagrant/dev/myIaC/terraform/infoblox
terraform init 
terraform apply
````

Output is 

````shell script
[vagrant@archlinux infoblox]$ terraform apply

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # infoblox_a_record.demo_record will be created
  + resource "infoblox_a_record" "demo_record" {
      + cidr              = "10.0.0.0/24"
      + dns_view          = "default"
      + id                = (known after apply)
      + ip_addr           = "42.42.42.42"
      + network_view_name = "dummy"
      + tenant_id         = "dummy"
      + vm_name           = "scoulomb-terraform"
      + zone              = "test.loc"
    }

  # infoblox_cname_record.demo_cname will be created
  + resource "infoblox_cname_record" "demo_cname" {
      + alias     = "demo1"
      + canonical = "demo.scoulomb.com"
      + dns_view  = "default"
      + id        = (known after apply)
      + tenant_id = "test"
      + zone      = "test.loc"
    }

Plan: 2 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

infoblox_a_record.demo_record: Creating...
infoblox_cname_record.demo_cname: Creating...
infoblox_cname_record.demo_cname: Creation complete after 0s [id=record:cname/<uid-1>:demo1.test.loc/default]
infoblox_a_record.demo_record: Creation complete after 0s [id=record:a/<uid-2>:scoulomb-terraform.test.loc/default]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
````

We can now even retrieve those records as made here in [myDNS repo](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#post-a)

```shell script
curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
       -H "Content-Type: application/json" \
       -X GET \
       "https://$API_ENDPOINT/wapi/v2.5/record:cname/<uid-1>:demo1.test.loc/default"

curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
       -H "Content-Type: application/json" \
       -X GET \
       "https://$API_ENDPOINT/wapi/v2.5/record:a/<uid-2>:scoulomb-terraform.test.loc/default"
```

Output is:

````shell script
[vagrant@archlinux infoblox]$ curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
>        -H "Content-Type: application/json" \
>        -X GET \
>        "https://$API_ENDPOINT/wapi/v2.5/record:cname/<uid-1>:demo1.test.loc/default"
{
    "_ref": "record:cname/<uid-1>:demo1.test.loc/default",
    "canonical": "demo.scoulomb.com",
    "name": "demo1.test.loc",
    "view": "default"
}
[vagrant@archlinux infoblox]$ curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
>        -H "Content-Type: application/json" \
>        -X GET \
>        "https://$API_ENDPOINT/wapi/v2.5/record:a/<uid-2>:scoulomb-terraform.test.loc/default"
{
    "_ref": "record:a/<uid-2>:scoulomb-terraform.test.loc/default",
    "ipv4addr": "42.42.42.42",
    "name": "scoulomb-terraform.test.loc",
    "view": "default"
````

And even with nslookup

````shell script

[vagrant@archlinux infoblox]$ nslookup -type=CNAME demo1.test.loc mydnsserver.loc
Server:         mydnsserver.loc
Address:        <ipv4.addr.of.dns-srv>#53

demo1.test.loc  canonical name = demo.scoulomb.com.

[vagrant@archlinux infoblox]$ nslookup -type=A scoulomb-terraform.test.loc mydnsserver.loc
Server:         mydnsserver.loc
Address:        <ipv4.addr.of.dns-srv>#53

Name:   scoulomb-terraform.test.loc
Address: 42.42.42.42
````

#### Destroy

And we delete the ressouce 

```shell script
[vagrant@archlinux infoblox]$ terraform destroy
infoblox_cname_record.demo_cname: Refreshing state... [id=record:cname/<uid-1>:demo1.test.loc/default]
infoblox_a_record.demo_record: Refreshing state... [id=record:a/<uid-2>:scoulomb-terraform.test.loc/default]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # infoblox_a_record.demo_record will be destroyed
  - resource "infoblox_a_record" "demo_record" {
      - cidr              = "10.0.0.0/24" -> null
      - dns_view          = "default" -> null
      - id                = "record:a/<uid-2>:scoulomb-terraform.test.loc/default" -> null
      - ip_addr           = "42.42.42.42" -> null
      - network_view_name = "dummy" -> null
      - tenant_id         = "dummy" -> null
      - vm_name           = "scoulomb-terraform" -> null
      - zone              = "test.loc" -> null
    }

  # infoblox_cname_record.demo_cname will be destroyed
  - resource "infoblox_cname_record" "demo_cname" {
      - alias     = "demo1" -> null
      - canonical = "demo.scoulomb.com" -> null
      - dns_view  = "default" -> null
      - id        = "record:cname/<uid-1>:demo1.test.loc/default" -> null
      - tenant_id = "test" -> null
      - zone      = "test.loc" -> null
    }

Plan: 0 to add, 0 to change, 2 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

infoblox_cname_record.demo_cname: Destroying... [id=record:cname/<uid-1>:demo1.test.loc/default]
infoblox_a_record.demo_record: Destroying... [id=record:a/<uid-2>:scoulomb-terraform.test.loc/default]
infoblox_cname_record.demo_cname: Destruction complete after 0s
infoblox_a_record.demo_record: Destruction complete after 1s

Destroy complete! Resources: 2 destroyed.
```


And testing well deleted:

```shell script
[vagrant@archlinux infoblox]$ curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
>        -H "Content-Type: application/json" \
>        -X GET \
>        "https://$API_ENDPOINT/wapi/v2.5/record:cname/<uid-1>:demo1.test.loc/default"
{ "Error": "AdmConDataNotFoundError: Reference record:cname/<uid-1>:demo1.test.loc/default not found",
  "code": "Client.Ibap.Data.NotFound",
  "text": "Reference record:cname/<uid-1>:demo1.test.loc/default not found"
}[vagrant@archlinux infoblox]$
[vagrant@archlinux infoblox]$ curl -k -H "Authorization: Basic $(cat ~/admin-credentials | base64)" \
>        -H "Content-Type: application/json" \
>        -X GET \
>        "https://$API_ENDPOINT/wapi/v2.5/record:a/<uid-2>:scoulomb-terraform.test.loc/default"
{ "Error": "AdmConDataNotFoundError: Reference record:a/<uid-2>:scoulomb-terraform.test.loc/default not found",
  "code": "Client.Ibap.Data.NotFound",
  "text": "Reference record:a/<uid-2>:scoulomb-terraform.test.loc/default not found"
}[vagrant@archlinux infoblox]$
[vagrant@archlinux infoblox]$ nslookup -type=CNAME demo1.test.loc mydnsserver.loc
Server:         mydnsserver.loc
Address:        <ipv4.addr.of.dns-srv>#53

** server can't find demo1.test.loc: NXDOMAIN

[vagrant@archlinux infoblox]$ nslookup -type=A scoulomb-terraform.test.loc mydnsserver.loc
Server:         mydnsserver.loc
Address:        <ipv4.addr.of.dns-srv>#53

** server can't find scoulomb-terraform.test.loc: NXDOMAIN

```

Note if changing value terraform will do the diff but plugin does not support it and it leads to an error.


### Comment on resource



#### CNAME record

```shell script
resource "infoblox_cname_record" "demo_cname"{
  canonical="demo.scoulomb.com"
  zone="test.loc"
  alias="demo1"
  tenant_id="test"
}
```

alias is not like alias host record.
It will create alias + zone pointing to canonical.


#### A record

````shell script
resource "infoblox_a_record" "demo_record" {
  vm_name="scoulomb-terraform"
  zone="test.loc"
  dns_view="default"
  ip_addr="42.42.42.42" //use the ip address used in IP allocation
  cidr="10.0.0.0/24"
  tenant_id="dummy"
  network_view_name="dummy"
}
````

vm_name is relative DNS name and we have a zone concept which we do not have in original API.
Cf:
- https://github.com/scoulomb/myDNS/blob/master/4-Analysis/1-comparison-table.md
- https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#Infoblox-View-and-Zone-creation

Here it creates A record where DNS name is: vm_name+zone.

The tenant_id, cidr and network view name are accepted because I added [EA in NIOS](#Configure-NIOS).
But not used IMO.
I think Cloud Edition is using it for extra validation.

[here]
Because the datamodel is the following. 
(to update my dns)
We have view and network view, if we check:
- https://docs.infoblox.com/display/NAG8/Configuring+DNS+Views
- https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#infoblox-view-and-zone-creation
- https://github.com/scoulomb/myDNS/blob/master/4-Analysis/1-comparison-table.md
(check p56)


