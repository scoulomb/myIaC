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

#### A record unused fields

##### Tests

The tenant_id, cidr and network view name are accepted because I added [EA in NIOS](#Configure-NIOS).
But not used IMO.
I think Cloud Edition is using it for extra validation.
Here we used default view attached to default network.

This matches our understanding of the datamodel made here:
https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#infoblox-network-view-and-view-and-zone-creation

To prove it I will use:
- `default.$networkName` view attached to `$networkName` network view (at network view creation).

Instead of 

- `Default` view is attached to `default` network view.

This was described [myDNS, Infoblox API overview](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#api-impact-and-wrapup)

And will show an invalid network name is accepted.


As done in [myDNS, Infoblox API overview](https://github.com/scoulomb/myDNS/blob/master/3-DNS-solution-providers/1-Infoblox/1-Infoblox-API-overview.md#api-impact-and-wrapup)

<!--
I had updated it for the occasion
-->

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

I will retrieve the created host record:

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


##### Clean-up

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
