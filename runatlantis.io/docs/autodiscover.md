# Auto Discover
By default, Atlantis will auto discover projects in repository when there are no projects explicitly configured (this is called auto mode).
This feature can be configured to always be disabled (never try to discover projects) or always be enabled (always try to discover projects).

## Configuration
### Default auto discover configuration
The default mode of AutoDiscover is "auto". This is equivalent to the below settings:
1. In the server repo config file:
    ```yaml
    repo:
      autodiscover:
        mode: auto
    ```
1. In the repo `atlantis.yaml` file:
    ```yaml
    version: 3
    autodiscover:
      mode: auto
    ```
1. Server Flags
```
atlantis server --autodiscover-mode=auto
```

### Disabling auto discover unconditionally 
1. In the server repo config file:
    ```yaml
    repo:
      autodiscover:
        mode: disabled
    ```
1. In the repo `atlantis.yaml` file:
    ```yaml
    version: 3
    autodiscover:
      mode: disabled
    ```
1. Server Flags
```
atlantis server --autodiscover-mode=disabled
```

### Enabling auto discover unconditionally 
1. In the server repo config file:
    ```yaml
    repo:
      autodiscover:
        mode: enabled
    ```
1. In the repo `atlantis.yaml` file:
    ```yaml
    version: 3
    autodiscover:
      mode: enabled

1. Server Flags
```
atlantis server --autodiscover-mode=enabled
```