# Kubernetes operator-demo
WebServers and WebManager Operator Demo with kubebuilder v2

## How demo is this?
- Demo1 : Keep fixed state defined by manifest of Deployment and Service 
(about Kubebuilder v1 version, reffer to https://github.com/cloudnativejp/webserver-operator/)
- Demo2 : Create WebServer(s) in order to decelerate on Manifest.
Then delete all created resource after some period.
- Demo3 : If you call REFILL request of Controller's REST API, you can refill one WebServer resource.

## Prerequisites
- Kubernetes 1.15.x
- Kubebuilder 2.0.0
- Kustomize

## Preparations
- Execute Kubebuilder init

      $ mkdir $GOPATH/src/cnc-demo
      $ cd $GOPATH/src/cnc-demo
      $ kubebuilder init --owner cnc-demo --domain cnc.demo

- Create templates API and Controller  
The 2 controlers and resouces is made by _'kubebuilder create api'_ commands.
	- WebServer
	
			$ kubebuilder create api --group servers --version v1beta1 --kind WebServer
			Create Resource [y/n]
			y
			Create Controller [y/n]
			y
			Writing scaffold for you to edit...
			api/v1beta1/webserver_types.go
			controllers/webserver_controller.go
			:
	
	- WebManager

			$ kubebuilder create api --group servers --version v1beta1 --kind WebManager
			Create Resource [y/n]
			y
			Create Controller [y/n]
			y
			Writing scaffold for you to edit...
			api/v1beta1/webmanager_types.go
			controllers/webmanager_controller.go
			:


## Edit Operator Codes

- File tree of API and Controllers

			./api
			└── v1beta1
			    ├── groupversion_info.go
			    ├── webmanager_types.go
			    ├── webserver_types.go
			    └── zz_generated.deepcopy.go
						
			./controllers
			├── suite_test.go
			├── webmanager_controller.go
			└── webserver_controller.go


## Build Operator

- Make manifets (e.g. rbac) from controler _marker_ in the code's comments

      $ make manifests

- Build Controller

      $ registry=[YOUR DOCKER REGISTORY]
      $ make docker-build docker-push IMG=${registry}  # Probably you need to login to your registry.

## Deploy Operator

- Deploy Controller

      $ ns=cnc-demo-system ; make deploy IMG=${registry}

You need to delete old operator pod once if have deployed already.

      $ kubectl get po -n $ns -o name | xargs kubectl -n $ns delete
      $ kubectl get po -n $ns -o wide

## Test Operator

- Show log of Controler Pod

      $ ns=cnc-demo-system ; kubectl get po -n $ns -o name | xargs -I {}  kubectl -n $ns logs {} manager -f

- Watch status of Resources

      $ while sleep 2 ; do echo ===WebServer ; kubectl get webservers.servers.cnc.demo ; done

      $ while sleep 2 ; do echo ===Webmanager ; kubectl get webmanagers.servers.cnc.demo ; done

      $ while sleep 2 ; do echo ===PODs ; kubectl get po  ; done


- Apply Resource
	- Demo1
		- Starting _WebServer Operator_
	
				$ kubectl apply -f config/samples/sample_webserver01.yaml

		- Delete _WebServer Operator_ after watching status
	
				$ kubectl apply -f config/samples/sample_webserver01.yaml

	- Demo2
		- Starting _WebManager Operator_
	
				$ kubectl apply -f config/samples/sample_webmanager01.yaml

		- Delete _WebManager Operator_ after watching status
		
				$ kubectl delete -f config/samples/sample_webmanager01.yaml

	- Demo3
		- Starting _WebManager Operator_
		
				$ kubectl apply -f config/samples/sample_webmanager02_no_auto_delete.yaml

		- Get WebManager Resource Information
		
				$ ns=cnc-demo-system
				$ k get po -n $ns -o go-template="{{range .items}}{{.status.podIP}}{{end}}" | xargs -I {} curl http://{}:10080

		- Refill WebServer Resource
		
				$ ns=cnc-demo-system
				$ k get po -n $ns -o go-template="{{range .items}}{{.status.podIP}}{{end}}" | xargs -I {} curl -X REFILL http://{}:10080

		- Delete _WebManager Operator_ after watching status
		
				$ kubectl delete -f config/samples/sample_webmanager02_no_auto_delete.yaml
