# Kubernetes operator-demo
WebServers and WebManager Operator Demo with kubebuilder v2


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

      $ make docker-build docker-push IMG=tomotake/cnc-demo

## Deploy Operator

- Deploy Controller

      $ ns=cnc-demo-system ; make deploy IMG=tomotake/cnc-demo

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
	- Starting _WebServer Operator_
	
			$ kubectl apply -f config/samples/sample_webserver01.yaml

	- Starting _WebManager Operator_
	
			$ kubectl apply -f config/samples/sample_webmanager01.yaml


