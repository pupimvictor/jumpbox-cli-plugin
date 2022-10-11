# Tanzu Jumpbox CLI

## Summary

Create and Manage Jumpbox VMs using Tanzu VM Service

## Overview

The Jumpbox plugin enables Platform Operators and Developers to easily create and access VMs in a Tanzu Namespace. 

The Jumpbox is a VM with a persistent volume and a persistent IP.  

## Installation

### Local Installation

- git clone https://github.com/pupimvictor/jumpbox-cli-plugin.git
- cd jumpbox-cli-plugin
- make build-install-local

### From remote repository

//TODO

## Usage

### Dependencies

//Content Library

### Create Jumpbox

```tanzu jumpbox create my-jumpbox  --namespace <vsphere-namespace> 
    --image <vm-image> \
    --class <vm-class> \
    --networkp-type <network-type> --network-name <network-name> \
    --ssh-pub <ssh-public-key> \
    --storage-class <storage-class>
```

- vsphere-namespace: Target Namespace
- vm-image: VM Image from Content Library. run `kubectl get virtualmachineimages` to see available images in the namespace
- vm-class: VM Class. run `kubectl get virtualmachineclasses` to see available vm classes
- network-type: `nsx-t` if Tanzu is deployed on NSX-T, `vsphere-distributed` if not using NSX-T
- network-name: network name for the VM. Required if network-type is vsphere-distributed
- ssh-public-key: Path to the ssh public key to include in VM authorized_keys (default "$HOME/.ssh/id_rsa.pub")
- storage-class: Storage class for VM filesystem and Persistent Volume

### Access Jumpbox

```tanzu jumpbox ssh my-jumpbox --namespace <vsphere-namespace> -i <ssh-private-key>```

- vsphere-namespace: Target Namespace
- ssh-private-key: Private key to access the VM

### Power jumpbox

#### Power On VM

```tanzu jumpbox power-on my-jumpbox --namespace <vsphere-namespace> ```

- vsphere-namespace: Target Namespace

#### Power Off VM

Turn Off VM without deleting data in `/workspace`

```tanzu jumpbox power-off my-jumpbox --namespace <vsphere-namespace> ```

- vsphere-namespace: Target Namespace

### Destroy

Destroy VM. Delete persistent volumes and Load Balancer

```tanzu jumpbox destroy my-jumpbox --namespace <vsphere-namespace> ```

- vsphere-namespace: Target Namespace

## Documentation

[include, or provide links to, additional resources that users or contributors may find useful here]

## Versioning

[describe how this plugin follows, or the degree to which it follows or doesn't follow semver]

## Contribution

[describe whether/how/where issues/PR's should be submitted]

## Development

[describe steps to clone, test, build/install locally, etc..]

## License

[name and link to the project this plugin is licensed under]
