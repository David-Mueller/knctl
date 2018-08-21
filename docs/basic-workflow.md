## Basic Workflow

Install Istio and Knative

```bash
$ knctl install
```

Deploy sample service (to current namespace)

```bash
$ knctl deploy --service hello --image gcr.io/knative-samples/helloworld-go --env TARGET=123
```

List deployed services

```bash
$ knctl list services

Services in namespace 'default'

Name   Domain                     Internal Domain                  Age
hello  hello.default.example.com  hello.default.svc.cluster.local  1d

1 services
```

Curl the deployed service and see that it responds

```bash
$ knctl curl --service hello

Running: curl '-H' 'Host: hello.default.example.com' 'http://0.0.0.0:80'

Hello World: 123!
```

Fetch recent logs from the deployed service

```bash
$ knctl logs -f --service hello
hello-00001 > hello-00001-deployment-7d4b4c5cc-v6jvl | 2018/08/02 17:21:51 Hello world sample started.
hello-00001 > hello-00001-deployment-7d4b4c5cc-v6jvl | 2018/08/02 17:22:04 Hello world received a request.
```

Change environment variable and see changes were applied

```bash
$ knctl deploy --service hello --image gcr.io/knative-samples/helloworld-go --env TARGET=new-value

$ knctl curl --service hello

Running: curl '-H' 'Host: hello.default.example.com' 'http://0.0.0.0:80'

Hello World: new-value!
```

List multiple revisions of the deployed service

```bash
$ knctl --service hello list revisions

Revisions for service 'hello'

Name         Allocated Traffic %  Serving State  Age
hello-00002  100%                 Active         2m
hello-00001  0%                   Reserve        3m

2 revisions
```

See how to:

- [deploy from public Git repo](./deploy-public-git-repo.md)
- [deploy from local source directory](./deploy-source-directory.md)
