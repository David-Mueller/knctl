## knctl open

Open web browser pointing at a service domain. Requires open command installed on the system.

### Synopsis

Open web browser pointing at a service domain. Requires open command installed on the system.

```
knctl open [flags]
```

### Examples

```

# Open web browser pointing at service 'svc1' in namespace 'ns1'
knctl open -s svc1 -n ns1
```

### Options

```
  -h, --help               help for open
  -n, --namespace string   Specified namespace
  -s, --service string     Specified service
```

### Options inherited from parent commands

```
      --column strings      Filter to show only given columns
      --json                Output as JSON
      --kubeconfig string   Path to the kubeconfig file (default "/Users/pivotal/.kube/config")
      --no-color            Disable colorized output
      --non-interactive     Don't ask for user input
      --tty                 Force TTY-like output
```

### SEE ALSO

* [knctl](knctl.md)	 - knctl controls Knative resources

