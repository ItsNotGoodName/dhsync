# dhsync

Sync day and night profile on Dahua cameras.

## Usage

See [example-config.yml](./example-config.yml) for a starter configuration.

### Example

Create the file `config.yml` with the following content.

```yml
---
latitude: 34.0522
longitude: -118.2437
timezone: America/Los_Angeles
sunrise_offset: 30m
sunset_offset: -1h20m
username: admin
password: password

cameras:
  - ip: 192.168.1.108
  - ip: 192.168.1.109
    sunset_offset: 20m
  - ip: 192.168.1.110
    name: FriendlyNameForLogging
    username: OverideDefaultUser
    password: OverideDefaultPassword123
```

The following command will sync cameras located at `192.168.1.108`, `192.168.1.109`, `192.168.1.110`.

```
dhsync sync
```

### Verify Settings

Shows day and night profiles currently on the cameras.

```
dhsync verify
```


### Show Version

```
dhsync --version
```

