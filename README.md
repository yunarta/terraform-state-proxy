# tfstate-proxy

## Overview

`tfstate-proxy` is a project designed to manage Terraform state files using a proxy server. 

## Features

- Terraform backend configuration for HTTP

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html)

## Configuring

To configure `tfstate-proxy`, you need to create a `config.toml` file. This file will contain the necessary
configuration settings for the proxy server. Here is an example of what the `config.toml` file might look like:

```toml
# config.toml
[bitbucket]
server = "https://bitbucket.mobilesolutionworks.com"

[gitea]
server = "https://gitea.mobilesolutionworks.com/"
```

### To enable encryption

```shell
export TF_STATE_ENCRYPTION_KEY=your-encryption-key
terraform-state-proxy
```

### Terraform Proxy for Bitbucket

```hcl
terraform {
  backend "http" {
    address  = "http://localhost:8080/bitbucket/terraform-states/test1/file.json?branch=master"
    username = "yunarta"
    password = "token"
  }
}
```

### Terraform Proxy for Gitea

```hcl
terraform {
  backend "http" {
    address  = "http://localhost:8080/gitea/terraform-states/test1/file.json?branch=main"
    username = "yunarta"
    password = "token"
  }
}
```

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.