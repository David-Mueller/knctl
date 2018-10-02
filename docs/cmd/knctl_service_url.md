## knctl service url

Print service URL

### Synopsis

Print service URL

```
knctl service url [flags]
```

### Examples

```

  # Print service 'svc1' URL in namespace 'ns1'
  knctl service url -s svc1 -n ns1
```

### Options

```
  -h, --help               help for url
  -n, --namespace string   Specified namespace ($KNCTL_NAMESPACE or default from kubeconfig)
  -p, --port int32         Set port (default 80)
  -s, --service string     Specified service
```

### Options inherited from parent commands

```
      --column strings              Filter to show only given columns
      --json                        Output as JSON
      --kubeconfig string           Path to the kubeconfig file ($KNCTL_KUBECONFIG or $KUBECONFIG)
      --kubeconfig-context string   Kubeconfig context override ($KNCTL_KUBECONFIG_CONTEXT)
      --no-color                    Disable colorized output
      --non-interactive             Don't ask for user input
      --tty                         Force TTY-like output
```

### SEE ALSO

* [knctl service](knctl_service.md)	 - Service management (annotate, delete, list, open, show, url)

