// File from: https://www.pulumi.com/docs/get-started/kubernetes/modify-program/ and add print
// And do: pulumi up; pulumi cancel; pulumi config set isMinikube true; pulumi up;
// Output is false and still ip issue (but trying to get lb external ip) as expected since no fix
// Then fix proposal made here: https://github.com/pulumi/docs/issues/3780 in go.mod
// pulumi cancel; pulumi destroy; pulumi stack rm dev; pulumi up ;
// Error
//[root@archlinux quickStartWithFix3780]#  pulumi up
//Please choose a stack, or create a new one: <create a new stack>
//Please enter your desired stack name.
//To create a stack in an organization, use the format <org-name>/<stack-name> (e.g. `acmecorp/dev`): scoulomb/dev
//Created stack 'dev'
//Previewing update (dev):
//     Type                 Name                       Plan     Info
//     pulumi:pulumi:Stack  quickStartWithFix3780-dev           1 error; 2 messages
//
//Diagnostics:
//  pulumi:pulumi:Stack (quickStartWithFix3780-dev):
//    # quickStartWithFix3780
//    ./main.go:54:12: cannot use v.Template (type "github.com/pulumi/pulumi-kubernetes/sdk/v2/go/kubernetes/core/v1".PodTemplateSpec) as type *"github.com/pulumi/pulumi-kubernetes/sdk/v2/go/kubernetes/core/v1".PodTemplateSpec in return argument
//
//    error: an unhandled error occurred: program exited with non-zero exit code: 2
//
//
//[root@archlinux quickStartWithFix3780]#
// Idea was to to do after
//pulumi config set isMinikube true; pulumi up
package main

import (
	"fmt"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v2/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v2/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v2/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		isMinikube := config.GetBool(ctx, "isMinikube")
		appName := "nginx"
		appLabels := pulumi.StringMap{
			"app": pulumi.String(appName),
		}
		deployment, err := appsv1.NewDeployment(ctx, appName, &appsv1.DeploymentArgs{
			Spec: appsv1.DeploymentSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: appLabels,
				},
				Replicas: pulumi.Int(1),
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: appLabels,
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String("nginx"),
								Image: pulumi.String("nginx"),
							}},
					},
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Println(isMinikube)
		feType := "LoadBalancer"
		if isMinikube {
			feType = "ClusterIP"
		}

		template := deployment.Spec.ApplyT(func(v *appsv1.DeploymentSpec) *corev1.PodTemplateSpec {
			return v.Template
		}).(corev1.PodTemplateSpecPtrOutput)

		meta := template.ApplyT(func(v *corev1.PodTemplateSpec) *metav1.ObjectMeta { return v.Metadata }).(metav1.ObjectMetaPtrOutput)

		frontend, err := corev1.NewService(ctx, appName, &corev1.ServiceArgs{
			Metadata: meta,
			Spec: &corev1.ServiceSpecArgs{
				Type: pulumi.String(feType),
				Ports: &corev1.ServicePortArray{
					&corev1.ServicePortArgs{
						Port:       pulumi.Int(80),
						TargetPort: pulumi.Int(80),
						Protocol:   pulumi.String("TCP"),
					},
				},
				Selector: appLabels,
			},
		})

		var ip pulumi.StringOutput

		if isMinikube {
			fmt.Println("it is Minikube")
			ip = frontend.Spec.ApplyString(func(val *corev1.ServiceSpec) string {
				return *val.ClusterIP
			})
		} else {
			ip = frontend.Status.ApplyString(func(val *corev1.ServiceStatus) string {
				if val.LoadBalancer.Ingress[0].Ip != nil {
					return *val.LoadBalancer.Ingress[0].Ip
				}
				return *val.LoadBalancer.Ingress[0].Hostname
			})
		}

		ctx.Export("ip", ip)
		return nil
	})
}