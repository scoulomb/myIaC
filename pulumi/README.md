# Pulumi 

## Setup

Follow:
- https://www.pulumi.com/docs/get-started/install/
- https://www.pulumi.com/docs/get-started/kubernetes/begin/

````
curl -fsSL https://get.pulumi.com | sh
# + Please add /home/vagrant/.pulumi/bin to your $PATH
set PATH /home/vagrant/.pulumi/bin $PATH
````
We do not do [move to bin](https://github.com/scoulomb/myIaC/tree/master/terraform#k8s) as several binaries 


## Deploy resource on k8s (Minikube)

Follow: https://www.pulumi.com/docs/get-started/kubernetes/
Use goland and install go on VM /and windows from goland

Issue with fish shortcut

````shell script
rm ~/.config/fish/functions/go.fish # (open new session)
````

````cassandraql
mkdir quickstart && cd quickstart && pulumi new kubernetes-go
````

We had an issue with Python even if pip was installed

````shell script
sudo pacman -S python-pip
sudo pacman -S python-pipenv
````

From
https://www.pulumi.com/docs/get-started/kubernetes/modify-program/

```
cd  ~/dev/myIaC/pulumi/quickstart
sudo su; fish ; # cf kubectl context as for terrafrom to target minikube
pulumi up
pulumi destroy`shell script
````


## Issues
 
- Python is not working
- `pulumi config set isMinikube false` not taken into account so hardcode it
- Not possible to get cluster IP (probably not updated yet when retrieved ), use port provided instead.


## Notes
Note we need an external token for access

For k8s it seems similar to https://github.com/ksonnet/ksonnet-lib 
