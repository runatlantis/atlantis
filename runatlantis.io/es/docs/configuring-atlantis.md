# Configuring Atlantis

There are three methods for configuring Atlantis:
1. Passing flags to the `atlantis server` command
1. Creating a server-side repo config file and using the `--repo-config` flag
1. Placing an `atlantis.yaml` file at the root of your Terraform repositories

## Flags
Flags to `atlantis server` are used to configure the global operation of
Atlantis, for example setting credentials for your Git Host
or configuring SSL certs.

See [Server Configuration](server-configuration.html) for more details.

## Server-Side Repo Config
A Server-Side Repo Config file is used to control per-repo behaviour
and what users can do in repo-level `atlantis.yaml` files.

See [Server-Side Repo Config](server-side-repo-config.html) for more details.

## Repo-Level `atlantis.yaml` Files
`atlantis.yaml` files placed at the root of your Terraform repos can be used to
change the default Atlantis behaviour for each repo.

See [Repo-Level atlantis.yaml Files](repo-level-atlantis-yaml.html) for more details.
