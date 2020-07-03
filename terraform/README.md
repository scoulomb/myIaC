# Terraform 

## Prerequisite 

A VM with docker and minikube setup.
For instance use this [setup](https://github.com/scoulomb/myk8s/blob/master/Setup/ArchDevVM/archlinux-dev-vm-with-minikube.md).

### k8s 

However in my case I start from a fresh machine (start-up issue) and issue in k8s setup, and when installing nanually temporary issue with signature

````shell script
➤ sudo pacman -Syu minikube
[...]
error: minikube: signature from "Christian Rebischke (Arch Linux Security Team-Member) <Chris.Rebischke@archlinux.org>" is unknown trust
:: File /var/cache/pacman/pkg/minikube-1.7.3-1-x86_64.pkg.tar.zst is corrupted (invalid or corrupted package (PGP signature)).
````
Hopefully we can install it manually as doicumented here:
https://kubernetes.io/docs/tasks/tools/install-minikube/#install-minikube-using-a-package

````shell script
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && chmod +x minikube
sudo install minikube /usr/local/bin/ # does not work
sudo mv ~/dev/minikube /usr/local/bin/minikube # do instead (as for Terraform)
sudo minikube start --vm-driver=none
````

Same for kubectl

````shell script
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
````

(in doc use mv here and not install, unlike minikube, pr to doc)

### Docker

Check docker status

````shell script
➤ systemctl status docker                                                                                                                                                     vagrant@archlinux● docker.service - Docker Application Container Engine
     Loaded: loaded (/usr/lib/systemd/system/docker.service; enabled; vendor preset: disabled)
     Active: active (running) since Fri 2020-07-03 11:53:14 UTC; 2h 8min ago
````

## Step 1: Install Terraform and deploy a nginx docker image from Terraform

Following:
https://learn.hashicorp.com/terraform/getting-started/install.html (done)


### Setup

We used manual installation: https://www.terraform.io/downloads.html
Unzip in window and copy it in vm sync folder then

````shell script
 sudo mv ~/dev/terraform_0.12.28_linux_amd64/terraform /usr/local/bin/terraform
````

<!--
Manual also because of package signature corrupted 
-->

### deploy 


````shell script
cd /home/vagrant/dev/myIaC/terraform/terraform-docker-demo
sudo su 
terraform init 
terraform apply
````

Output is

````shell script
docker_image.nginx-image: Still creating... [1m0s elapsed]
docker_image.nginx-image: Creation complete after 1m2s [id=sha256:2622e6cca7ebbb6e310743abce3fc47335393e79171b9d76ba9d4f446ce7b163nginx:latest]
docker_container.nginx-container: Creating...
docker_container.nginx-container: Creation complete after 1s [id=c72c05c08211d5a04177d61956da3f87b4476665caac5ee4e7c55ce6797c7f0a]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
````

and docker ps/images output is

````shell script
[root@archlinux terraform-docker-demo]# docker ps | grep tutorial
c72c05c08211        2622e6cca7eb                              "/docker-entrypoint.…"   About a minute ago   Up About a minute   0.0.0.0:8000->80/tcp   tutorial 
[root@archlinux terraform-docker-demo]# docker images | grep nginx
nginx                                     latest              2622e6cca7eb        3 weeks ago         132MB
````

And then doing destroy

````shell script
terraform destroy
````

Output is 


````shell script
docker_container.nginx-container: Destruction complete after 0s
docker_image.nginx-image: Destroying... [id=sha256:2622e6cca7ebbb6e310743abce3fc47335393e79171b9d76ba9d4f446ce7b163nginx:latest]
docker_image.nginx-image: Destruction complete after 0s

Destroy complete! Resources: 2 destroyed.
````
However image is still present even if 2 resources are destroyed 

````shell script
[root@archlinux terraform-docker-demo]# docker ps | grep tutorial
[root@archlinux terraform-docker-demo]# docker images | grep nginx
nginx                                     latest              2622e6cca7eb        3 weeks ago         132MB
````

To delete also the image change `keep_locally = false`, buy we will have to redeploy fully the app :(

````shell script
[root@archlinux terraform-docker-demo]# terraform destroy
Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes


Destroy complete! Resources: 0 destroyed.
[root@archlinux terraform-docker-demo]# docker images | grep nginx
nginx                                     latest              2622e6cca7eb        3 weeks ago         132MB
[root@archlinux terraform-docker-demo]# terraform apply

An execution plan has been generated and is shown below.
[...]
Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
[root@archlinux terraform-docker-demo]# docker ps | grep tutorial
b9a2773b5e16        2622e6cca7eb                              "/docker-entrypoint.…"   7 seconds ago       Up 6 seconds        0.0.0.0:8000->80/tcp   tutorial
[root@archlinux terraform-docker-demo]# docker images | grep nginx
nginx                                     latest              2622e6cca7eb        3 weeks ago         132MB
[root@archlinux terraform-docker-demo]# terraform destroy
docker_image.nginx-image: Refreshing state... [id=sha256:2622e6cca7ebbb6e310743abce3fc47335393e79171b9d76ba9d4f446ce7b163nginx:latest]
docker_container.nginx-container: Refreshing state... [id=b9a2773b5e16f4d590644a95997a794c9708307b46d906bd1ae27ba304f62135]
[...]
Destroy complete! Resources: 2 destroyed.
[root@archlinux terraform-docker-demo]# docker ps | grep tutorial
[root@archlinux terraform-docker-demo]# docker images | grep nginx
[root@archlinux terraform-docker-demo]#
````


## Step 2: deploy a nginx pod from terraform 

See k8s: https://learn.hashicorp.com/terraform/kubernetes/deploy-nginx-kubernetes (done)

### Cluster check


Check kube-config is targeting Minikube (cf. sudo su) and explained here in:
https://github.com/scoulomb/myk8s/blob/master/Master-Kubectl/kube-config.md#undertsand-setup (added here for the occasion)
And not a production cluster (if fresh vm should be fine). 

Thus why using sudo user

````shell script
[root@archlinux terraform-docker-demo]# cat ~/.kube/config
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /root/.minikube/ca.crt
    server: https://10.0.2.15:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
````

<!--
from here link with myk8s and DNS, no crossref
-->

### Apply

````shell script
cd /home/vagrant/dev/myIaC/terraform/learn-terraform-deploy-nginx-kubernetes
sudo su
terraform init
````

Then check pods

````shell script
[root@archlinux learn-terraform-deploy-nginx-kubernetes]# kubectl get po
No resources found in default namespace.
````

And terraform

````shell script
terraform apply
````


Output is

````shell script
[root@archlinux learn-terraform-deploy-nginx-kubernetes]# terraform apply

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:
[...]
kubernetes_deployment.nginx: Creation complete after 1m16s [id=default/scalable-nginx-example]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
[root@archlinux learn-terraform-deploy-nginx-kubernetes]# kubectl get pods
NAME                                    READY   STATUS    RESTARTS   AGE
scalable-nginx-example-f874c6d4-4whnj   1/1     Running   0          2m6s
scalable-nginx-example-f874c6d4-7x549   1/1     Running   0          2m6s

[root@archlinux learn-terraform-deploy-nginx-kubernetes]# terraform destroy
kubernetes_deployment.nginx: Refreshing state... [id=default/scalable-nginx-example]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # kubernetes_deployment.nginx will be destroyed
  - resource "kubernetes_deployment" "nginx" {
      - id = "default/scalable-nginx-example" -> null
[...]
kubernetes_deployment.nginx: Destroying... [id=default/scalable-nginx-example]
kubernetes_deployment.nginx: Destruction complete after 0s

Destroy complete! Resources: 1 destroyed.
[root@archlinux learn-terraform-deploy-nginx-kubernetes]# kubectl get pods
NAME                                    READY   STATUS        RESTARTS   AGE
scalable-nginx-example-f874c6d4-4whnj   0/1     Terminating   0          2m42s
scalable-nginx-example-f874c6d4-7x549   0/1     Terminating   0          2m42s
[root@archlinux learn-terraform-deploy-nginx-kubernetes]# kubectl get pods
No resources found in default namespace.
[root@archlinux learn-terraform-deploy-nginx-kubernetes]#
````

Does not give back the hand for a create (wait running status), but do it for a destroy.

Note they have their own language it not a link to deployment.
Helm chart terraform plugin points to an Helm chart and it better (similar to a custom helm operator linked to a helm chart)

I guess they find api version like this:
https://stackoverflow.com/questions/52711326/kubernetes-how-to-know-latest-supported-api-version

## Wrap-up Terraform

### Summary

Funny because can deploy cloud vm then the install docker/k8s.
Then can deploy docker images directly or kubernetes pod (like an automation server) as shown here.
So terraform could deploy automation server targeting a device.

But we could also apply a config directly on a device with Terraform (Infoblox plugin, Azure DNS plugin).

### Terraform and Infoblox

Usually we use Terraform to call Ansible: https://www.hashicorp.com/resources/ansible-terraform-better-together/
And also vagrant has an Ansible, Shell, Salt provisionner.
But does not have a Terraform provisionner because in Hashicorp vision: https://www.vagrantup.com/intro/vs/terraform
> The primary usage of Terraform is for managing remote resources in cloud providers such as AWS. Terraform is designed to be able to manage extremely large infrastructures that span multiple cloud providers. Vagrant is designed primarily for local development environments that use only a handful of virtual machines at most.


[Next: Terraform and DNS configuration](README-terraform-infoblox.md)