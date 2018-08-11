## knctl list pods

List pods

### Synopsis

List all pods for a service

```
knctl list pods [flags]
```

### Options

```
  -h, --help               help for pods
  -n, --namespace string   Specified namespace (can be provided via environment variable KNCTL_NAMESPACE)
  -s, --service string     Specified service
```

### Options inherited from parent commands

```
      --column strings      Filter to show only given columns
      --json                Output as JSON
      --kubeconfig string   Path to the kubeconfig file (can be provided via environment variable KNCTL_KUBECONFIG) (default "/Users/pivotal/.kube/config")
      --no-color            Disable colorized output
      --non-interactive     Don't ask for user input
      --tty                 Force TTY-like output
```

### SEE ALSO

* [knctl list](knctl_list.md)	 - List resources (services, revisions, builds, pods, ingresses)

