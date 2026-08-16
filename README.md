# dhsync

Sync day and night profile times on Dahua cameras.

Dahua cameras are capable of having custom video settings for day and night.
The camera can switch between day and night using a fixed daily schedule or time plan on newer cameras.
This program calculate the sunrise and sunset for you location and syncs it with the schedule or time plan.

## Usage

See [example-config.yml](./example-config.yml) for a starter configuration.

### Example

Create the file `config.yml` with the following content.

```yml
---
healthcheck_url: "https://hc-ping.com/d5f22685-6129-4682-8960-2bcb47fadc89"
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

The following command will sync cameras at `192.168.1.108`, `192.168.1.109`, `192.168.1.110`.

```
dhsync sync
```

### Run Daemon

Run as a background service that syncs every 24 hours.

```
dhsync daemon
```

### Read Settings

Shows day and night profile times currently on the cameras.

```
dhsync read
```

### Show Version

```
dhsync --version
```
