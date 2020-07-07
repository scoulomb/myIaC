# Terraform 

## Comment on resource

### CNAME record

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


### A record

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
- https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md

Here it creates A record where matching FQDN is : (vm_name+zone)

## A record unused fields

### Observations

The cidr and network view name are present and part of a A record configuration when using the plugin.

Thus we have two distinct concept which are:
- view 
- network view, where network view (and attached networks) are actually a distinct concept from ACL!

<!--
I had confused the network view and AcLs
https://github.com/scoulomb/myDNS/commit/5121958f56111eec6f9075398ce11b4c95e5119d 
https://github.com/scoulomb/myDNS/commit/f02b97ffc92ebc59408c0ba837e6a264f645ec49 
https://github.com/scoulomb/myDNS/commit/aa0bf6c4181c35fd2fa3b7b8c1c0a7ab502bf2ea
-->

Based on this observation I updated the understanding of view, network view and ACL in myDNS repo. Check this before continue reading:
- https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#api-impact-and-wrapup
- https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#how-to-have-a-different-answer-based-on-client-ip-with-the-view-mechanism-as-done-with-bind9

==> conclusion is that a view is attached to a network view, thus network view should not be part of a record definition.

Still we can see go client includes it in a [record definition](https://github.com/infobloxopen/infoblox-go-client/blob/master/object_manager.go).

````shell script
CreateARecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, ea EA) (*RecordA, error)
CreateCNAMERecord(canonical string, recordname string, dnsview string, ea EA) (*RecordCNAME, error)
CreateHostRecord(enabledns bool, recordName string, netview string, dnsview string, cidr string, ipAddr string, macAddress string, ea EA) (*HostRecord, error)
````

We will check the reason behind later.

From this understanding, in the previous example we can say we used default view attached to default network.

And as the `cidr` and `network_view_name` were set to dummy,and we can make the hypothesis they are not used by the plugin

### Protocol

To prove it I will use in this new [Terraform file](infoblox/view_test/main.tf)
- `default.$networkName` view attached to `$networkName` network view (automatically created at network view creation, as describe in [myDNS]([myDNS, Infoblox API overview](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#api-impact-and-wrapup))).

Instead of 

- `Default` view is attached to `default` network view.


And will show an invalid network name is accepted. As `default.scoulomb-nw` is not attached to `dummy` network view.

### Execution

#### Environment preparation

I will create a network:


````shell script
sudo pacman -S jq
export API_ENDPOINT="dns_server_dns_name" # IP or FQDN to DNS
export USERNAME="admin"
export PASSWORD="infoblox"

curl -k -u $USERNAME:$PASSWORD -H 'content-type: application/json' -X POST "https://$API_ENDPOINT/wapi/v2.5/networkview?_return_fields%2B=name&_return_as_object=1" -d '{"name":
"scoulomb-nw"}'

export network_view_id=$(curl -k -u $USERNAME:$PASSWORD \
        -H "Content-Type: application/json" \
        -X GET \
        "https://$API_ENDPOINT/wapi/v2.5/networkview?name=scoulomb-nw" |  jq .[0]._ref |  tr -d '"')
echo $network_view_id

````

This created the view:

````shell script
curl -k -u $USERNAME:$PASSWORD -H 'content-type: application/json' -X GET https://$API_ENDPOINT/wapi/v2.5/view | grep -C 2 "scoulomb-nw"
````

Output is

````shell script
[vagrant@archlinux myIaC]$ echo $network_view_id
networkview/ZG5zLm5ldHdvcmtfdmlldyQxOA:scoulomb-nw/false
[vagrant@archlinux myIaC]$
[vagrant@archlinux myIaC]$ curl -k -u $USERNAME:$PASSWORD -H 'content-type: application/json' -X GET https://$API_ENDPOINT/wapi/v2.5/view | grep -C 2 "scoulomb-nw"
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  1000    0  1000    0     0   2967      0 --:--:-- --:--:-- --:--:--  2976
    },
    {
        "_ref": "view/ZG5zLnZpZXckLjMx:default.scoulomb-nw/false",
        "is_default": false,
        "name": "default.scoulomb-nw"
    }
]
````

I will also need to create a `scoulomb.loc` zone in that view:

````shell script
curl -k -u $USERNAME:$PASSWORD -H 'content-type: application/json' -X POST "https://$API_ENDPOINT/wapi/v2.5/zone_auth?_return_fields%2B=fqdn,network_view&_return_as_object=1" -d \
'{"fqdn": "scoulomb.loc","view": "default.scoulomb-nw"}'


export find_zone_res=$(curl -k -u $USERNAME:$PASSWORD \
        -H "Content-Type: application/json" \
        -X GET \
        "https://$API_ENDPOINT/wapi/v2.5/zone_auth?fqdn=scoulomb.loc&view=default.scoulomb-nw")
# if using test.loc, add the view in the query!
echo $find_zone_res
export view_scoulomb_loc=$(echo $find_zone_res | jq .[0]._ref |  tr -d '"')
echo $view_scoulomb_loc

````

#### terraform in action

In my Terraform file, in `view_test` folder.
I will use that view but a dummyNetworkName instead of `scoulomb-nw` which would have been correct.

````shell script
cd /home/vagrant/dev/myIaC/terraform/infoblox/view_test
terraform init 
terraform apply
````

Output is 

````shell script
[vagrant@archlinux view_test]$ terraform apply
infoblox_cname_record.demo_cname: Refreshing state... [id=record:cname/ZG5zLmJpbmRfY25hbWUkLl9kZWZhdWx0LmxvYy50ZXN0LmRlbW8x:demo1.test.loc/default]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # infoblox_a_record.demo_record will be created
  + resource "infoblox_a_record" "demo_record" {
      + cidr              = "10.0.0.0/24"
      + dns_view          = "default.scoulomb-nw"
      + id                = (known after apply)
      + ip_addr           = "42.42.42.42"
      + network_view_name = "dummy"
      + tenant_id         = "dummy"
      + vm_name           = "scoulomb-terraform"
      + zone              = "scoulomb.loc"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

infoblox_a_record.demo_record: Creating...
infoblox_a_record.demo_record: Creation complete after 1s [id=record:a/<a-uid>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
````

I will retrieve the created A record:

````shell script
curl -k -u $USERNAME:$PASSWORD -H "Content-Type: application/json"-X GET "https://$API_ENDPOINT/wapi/v2.5/record:a/<a-uid>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw?_return_fields%2B=name&_return_as_object=1" | jq
````

Output is
````shell script
{
  "result": {
    "_ref": "record:a/<a-uid>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw",
    "ipv4addr": "42.42.42.42",
    "name": "scoulomb-terraform.scoulomb.loc",
    "view": "default.scoulomb-nw"
  }
}
````

We can see the record in the UI by selecting the view when in dropdown list  `scoulomb-nw` is selected.
Cf. [myDNS](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#can-i-decide-to-make-custom-direct-view-creation-and-assign-it-to-a-network-view).
I can not show the `network-view` field here unlike Host `Unknown argument/field: 'network_view'"`
But we know from [myDNS](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#api-impact-and-wrapup).
> - default.$networkName view is attached to $networkName network view (at network view creation).
> - A DNS view can be in one network view only, but a network view can have multiple DNS views.

As a consequence dummy value has not been taken into account. CQFD.


#### Clean up 1

````shell script
terraform destroy
curl -k -u $USERNAME:$PASSWORD \
        -H "Content-Type: application/json" \
        -X DELETE \
        "https://$API_ENDPOINT/wapi/v2.5/$view_scoulomb_loc"
curl -k -u $USERNAME:$PASSWORD \
        -H "Content-Type: application/json" \
        -X DELETE \
        "https://$API_ENDPOINT/wapi/v2.5/$network_view_id"
````

<!--
nslookup not purpose here
Understand CNAME no concept of network even if in view ok
-->


## Plugin Code and analysis

Plugin doc: https://www.terraform.io/docs/providers/infoblox/r/a_record.html

From the Terraform plugin code here:

https://github.com/terraform-providers/terraform-provider-infoblox/blob/master/infoblox/resource_infoblox_a_record.go

````go

func resourceARecordCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] %s: Beginning to create A record from  required network block", resourceARecordIDString(d))

	networkViewName := d.Get("network_view_name").(string)
	//This is for record Name
	recordName := d.Get("vm_name").(string)
	ipAddr := d.Get("ip_addr").(string)
	cidr := d.Get("cidr").(string)
	vmID := d.Get("vm_id").(string)
	//This is for vm name
	vmName := d.Get("vm_name").(string)
	zone := d.Get("zone").(string)
	dnsView := d.Get("dns_view").(string)
	tenantID := d.Get("tenant_id").(string)
	connector := m.(*ibclient.Connector)

	ea := make(ibclient.EA)

	ea["VM Name"] = vmName

	if vmID != "" {
		ea["VM ID"] = vmID
	}

	objMgr := ibclient.NewObjectManager(connector, "Terraform", tenantID)
	// fqdn
	name := recordName + "." + zone
	recordA, err := objMgr.CreateARecord(networkViewName, dnsView, name, cidr, ipAddr, ea)
	if err != nil {
		return fmt.Errorf("Error creating A Record from network block(%s): %s", cidr, err)
	}

	d.Set("recordName", name)
	d.SetId(recordA.Ref)

	log.Printf("[DEBUG] %s: Creation of A Record complete", resourceARecordIDString(d))
	return resourceARecordGet(d, m)
}

````

Which is calling Infoblox go client:
https://github.com/infobloxopen/infoblox-go-client/blob/master/object_manager.go

````go
func NewObjectManager(connector IBConnector, cmpType string, tenantID string) *ObjectManager {
	objMgr := new(ObjectManager)

	objMgr.connector = connector
	objMgr.cmpType = cmpType
	objMgr.tenantID = tenantID

	return objMgr
}


func (objMgr *ObjectManager) CreateARecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, ea EA) (*RecordA, error) {

	eas := objMgr.extendEA(ea)

	recordA := NewRecordA(RecordA{
		View: dnsview,
		Name: recordname,
		Ea:   eas})

	if ipAddr == "" {
		recordA.Ipv4Addr = fmt.Sprintf("func:nextavailableip:%s,%s", cidr, netview)
	} else {
		recordA.Ipv4Addr = ipAddr
	}
	ref, err := objMgr.connector.CreateObject(recordA)
	recordA.Ref = ref
	return recordA, err
}
````

We understand:
- The reason we added added [EA in NIOS](README-terraform-infoblox-part1-basic.md#Configure-NIOS) for VM name and IP.
- Why the tenant can be dummy (it results to an EA)
- And understand clearly in the code that when an IP is provided the network view and CIDR are not used.
They are used for IP automatic allocation.

It is equivalent to this rest call (https://www.infoblox.com/wp-content/uploads/infoblox-deployment-infoblox-rest-api.pdf, p30):

````shell script
# Use the function next_available_ip by specifying it in the option _function. You can use it in the longhand form
# or the shorthand form(func). The following example also covers the different forms func supports in shorthand.
curl -k -u admin:infoblox -H 'content-type: application/json' -X POST "https://gridmaster/wapi/v2.11/record:host?_return_fields%2B=name,ipv4addrs&_return_as_object=1 " -d
'{"name":"wapi.info.com","ipv4addrs":[{"ipv4addr":{"_object_function":"next_available_ip","_parameters":{"exclude":[
"10.10.10.1","10.10.10.2"]},"_result_field":"ips","_object" :
"network","_object_parameters":{"network":"10.10.10.0/24"}}}]}'
````

As a consequence it no ip address is provided, we can allocate it with the IPAM and use it.
This the reason why network and cidr are defined in function parameters!

## A record IP allocation

### Prepare

I will reuse same instruction as [execution/environment preparation](#environment-preparation).

I will also need to add a network matching the cidr in terraform to that network view.
We already did that in [myDNS](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#add-a-network-within-a-network-view).

````shell script
curl -k -u $USERNAME:$PASSWORD -H 'content-type: application/json' -X POST "https://$API_ENDPOINT/wapi/v2.5/network?_return_fields%2B=network&_return_as_object=1 " -d '{"network":
"10.0.0.0/24","network_view": "scoulomb-nw"}'

export network_id=$(curl -k -u $USERNAME:$PASSWORD \
        -H "Content-Type: application/json" \
        -X GET \
        "https://$API_ENDPOINT/wapi/v2.5/network?network_view=scoulomb-nw" |  jq .[0]._ref |  tr -d '"')
echo $network_id
````

### Apply

````shell script
cd /home/vagrant/dev/myIaC/terraform/infoblox/ip_allocation
terraform init 
terraform apply
````

Output is 


````shell script
vagrant@archlinux ip_allocation]$ terraform apply
infoblox_cname_record.demo_cname: Refreshing state... [id=record:cname/ZG5zLmJpbmRfY25hbWUkLl9kZWZhdWx0LmxvYy50ZXN0LmRlbW8x:demo1.test.loc/default]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # infoblox_a_record.demo_record will be created
  + resource "infoblox_a_record" "demo_record" {
      + cidr              = "10.0.0.0/24"
      + dns_view          = "default.scoulomb-nw"
      + id                = (known after apply)
      + network_view_name = "scoulomb-nw"
      + tenant_id         = "dummy"
      + vm_name           = "scoulomb-terraform"
      + zone              = "scoulomb.loc"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

infoblox_a_record.demo_record: Creating...
infoblox_a_record.demo_record: Creation complete after 0s [id=record:a/<a-uid-0>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
````


I will retrieve the created A record:

````shell script
curl -k -u $USERNAME:$PASSWORD -H "Content-Type: application/json"-X GET "https://$API_ENDPOINT/wapi/v2.5/record:a/<a-uid-0>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw" | jq
````

Output is
````json
{
  "result": {
    "_ref": "record:a/<a-uid>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw",
    "ipv4addr": "42.42.42.42",
    "name": "scoulomb-terraform.scoulomb.loc",
    "view": "default.scoulomb-nw"
  }
}
````

Output is

````json
{
  "_ref": "record:a/<a-uid-0>:scoulomb-terraform.scoulomb.loc/default.scoulomb-nw",
  "ipv4addr": "10.0.0.1",
  "name": "scoulomb-terraform.scoulomb.loc",
  "view": "default.scoulomb-nw"
}
````

If I have Terraform file with a second host (when copying file remove state)

````shell script
cd /home/vagrant/dev/myIaC/terraform/infoblox/ip_allocation_2
terraform init
terraform apply
````

Where output is:
````shell script
infoblox_a_record.demo_record: Creation complete after 0s [id=record:a/<a-uid-1>:scoulomb-terraform-2.scoulomb.loc/default.scoulomb-nw]
````

And retrieve

````shell script
curl -k -u $USERNAME:$PASSWORD -H "Content-Type: application/json"-X GET "https://$API_ENDPOINT/wapi/v2.5/record:a/<a-uid-1>:scoulomb-terraform-2.scoulomb.loc/default.scoulomb-nw" | jq
````

Output is 

````json
{
  "_ref": "record:a/<a-uid-1>:scoulomb-terraform-2.scoulomb.loc/default.scoulomb-nw",
  "ipv4addr": "10.0.0.3",
  "name": "scoulomb-terraform-2.scoulomb.loc",
  "view": "default.scoulomb-nw"
}
````

We can see a different IP was allocated.

**So we include network view and CIDR if we need IP allocation made automatically.**

 
### Clean up 2


````shell script
cd /home/vagrant/dev/myIaC/terraform/infoblox/ip_allocation_2
terraform destroy
cd /home/vagrant/dev/myIaC/terraform/infoblox/ip_allocation
terraform destroy
curl -k -u $USERNAME:$PASSWORD \
        -H "Content-Type: application/json" \
        -X DELETE \
        "https://$API_ENDPOINT/wapi/v2.5/$network_id"
````

Then same as [clean up 1](#clean-up-1) curl request.

## Note on host record

From plugin: https://github.com/terraform-providers/terraform-provider-infoblox/tree/master/infoblox

HostRecord can not be bound.

Actaully to create a host record we need to use IP allocation or association. 

- https://github.com/terraform-providers/terraform-provider-infoblox/blob/v1.0.0/infoblox/resource_infoblox_ip_association.go#L77
- https://github.com/terraform-providers/terraform-provider-infoblox/blob/v1.0.0/infoblox/resource_infoblox_ip_allocation.go

We can see we create host record if we provide zone and dns view in plugn.
Host record creation when not given an IP in go client code will allocate IP as for A record.

Otherwise plugin will just allocate the IP.


````go
# https://github.com/terraform-providers/terraform-provider-infoblox/blob/v1.0.0/infoblox/resource_infoblox_ip_association.go
func Resource(d *schema.ResourceData, m interface{}) error {

	matchClient := "MAC_ADDRESS"
	networkViewName := d.Get("network_view_name").(string)
	Name := d.Get("vm_name").(string)
	ipAddr := d.Get("ip_addr").(string)
	cidr := d.Get("cidr").(string)
	macAddr := d.Get("mac_addr").(string)
	tenantID := d.Get("tenant_id").(string)
	vmID := d.Get("vm_id").(string)
	zone := d.Get("zone").(string)
	dnsView := d.Get("dns_view").(string)

	connector := m.(*ibclient.Connector)

	objMgr := ibclient.NewObjectManager(connector, "Terraform", tenantID)
	//conversion from bit reversed EUI-48 format to hexadecimal EUI-48 format
	macAddr = strings.Replace(macAddr, "-", ":", -1)
	name := Name + "." + zone

	if (zone != "" || len(zone) != 0) && (dnsView != "" || len(dnsView) != 0) {
		hostRecordObj, err := objMgr.GetHostRecord(name, networkViewName, cidr, ipAddr)
		if err != nil {
			return fmt.Errorf("GetHostRecord failed from network block(%s):%s", cidr, err)
		}
		_, err = objMgr.UpdateHostRecord(hostRecordObj.Ref, ipAddr, macAddr, vmID, Name)
		if err != nil {
			return fmt.Errorf("UpdateHost Record error from network block(%s):%s", cidr, err)
		}
		d.SetId(hostRecordObj.Ref)
	} else {
		fixedAddressObj, err := objMgr.GetFixedAddress(networkViewName, cidr, ipAddr, "")
		if err != nil {
			return fmt.Errorf("GetFixedAddress error from network block(%s):%s", cidr, err)
		}

		_, err = objMgr.UpdateFixedAddress(fixedAddressObj.Ref, matchClient, macAddr, vmID, Name)
		if err != nil {
			return fmt.Errorf("UpdateFixedAddress error from network block(%s):%s", cidr, err)
		}
		d.SetId(fixedAddressObj.Ref)
	}
	return nil
}
````

It seems allocation accept empty IP from the code (stop here), but I would expect it to be provided.

This makes us understand the disclaimer: https://github.com/terraform-providers/terraform-provider-infoblox/tree/v1.0.0#disclaimer
