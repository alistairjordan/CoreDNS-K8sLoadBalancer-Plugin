# Kubernetes LoadBalancer

## Name

*k8s_loadbalancer* - Exports Services of type LoadBalancer to external DNS

## Description

The plugin currently takes services of type LoadBalancer from all namespaces and exports it to the .kube. zone.
For example:
```
  Service Name: test-wordpress
  DNS: test-wordpress.kube
```

This can be configured as shown in the configuration section below.

## Compilation

This package will always be compiled as part of CoreDNS and not in a standalone way. It will require you to use `go get` or as a dependency on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg).

The [manual](https://coredns.io/manual/toc/#what-is-coredns) will have more information about how to configure and extend the server with external plugins.

A simple way to consume this plugin, is by adding the following on [plugin.cfg](https://github.com/coredns/coredns/blob/master/plugin.cfg), and recompile it as [detailed on coredns.io](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/#build-with-compile-time-configuration-file).

~~~
k8s_loadbalancer:k8s_loadbalancer
~~~

Put this early in the plugin list, so that *example* is executed before any of the other plugins.

After this you can compile coredns by:

``` sh
go generate
go build
```

Or you can instead use make:

``` sh
make
```

## Syntax

~~~ txt
example
~~~

## Metrics

If monitoring is enabled (via the *prometheus* directive) the following metric is exported:

* `coredns_example_request_count_total{server}` - query count to the *example* plugin.

The `server` label indicated which server handled the request, see the *metrics* plugin for details.

## Ready

This plugin reports readiness to the ready plugin. It will be immediately ready.

## Examples

In this configuration, we forward all queries to 8.8.8.8 and resolve anything going to the .kube zone.

~~~ corefile
. {
  k8s_loadbalancer
  forward . 8.8.8.8
  debug
}
~~~

Or without any external connectivity:

~~~ corefile
. {
  k8s_loadbalancer
}
~~~

Without any zones being specified, the plugin will default to the .kube. zone. 
Additional zones can be added as below (This will add the test, and test2 zones)

~~~ corefile
. {
  k8s_loadbalancer .test. .test2. {
        enableNSZone true
        enableRootZone true
        kubeConfigPath /path/to/.kube/config
  }
}
~~~

Additonally Namespace Zones can be enabled for example:
```
service.namespace.zone 
```
This can be done using the enableNSZone flag (true/false)

## Also See

See the [manual](https://coredns.io/manual).
