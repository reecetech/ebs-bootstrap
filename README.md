# EBS Bootstrap

## Build

`ebs-bootstrap` can be built locally regardless of the architecture of the host machine. This is facilitated by a multi-architecture Docker build process. The currently supported architechtures are `linux/amd64` and `linux/arm64`.

```bash
# Specific Architecture
./build/docker.sh --architecture arm64
ls -la
... ebs-bootstrap-linux-aarch64

# All Architectures
./build/docker.sh
ls -la
... ebs-bootstrap-linux-aarch64
... ebs-bootstrap-linux-x86_64
```

## Recommended Setup

### `systemd`

The ideal way of operating `ebs-bootstrap` is through a `systemd` service. This is so we can configure it as a `oneshot` service type that executes after the file system is ready and after `clout-init.service` writes any config files to disk. The latter is essential as `ebs-bootstrap` consumes a config file that is located at `/etc/ebs-boostrap/config.yml` by default. 

`ExecStopPost=-...` con point torwards a script that is executed when the `ebs-bootstrap` service exits on either success or failure. This is a suitable place to include logic to notify a human operator that the configured devices failed their relevant healthchecks and the underlying application failed to launch in the process.

```ini
[Unit]
Description=EBS Bootstrap
After=local-fs.target cloud-init.service

[Service]
Type=oneshot
RemainAfterExit=true
StandardInput=null
ExecStart=ebs-bootstrap
PrivateMounts=no
MountFlags=shared
ExecStopPost=-/etc/ebs-bootstrap/post-hook.sh

[Install]
WantedBy=multi-user.target
```

```
cat /etc/ebs-bootstrap/post-hook.sh
#!/bin/sh
if [ "${EXIT_STATUS}" = "0" ]; then
    echo "🟢 Post Stop Hook: Success"
else
    echo "🔴 Post Stop Hook: Failure"
fi
```

It is then possible to configure another `systemd` service to only start if the `ebs-bootstrap` service is successful. Certain databases support the ability to spread database chunks across multiple devices that need to be mounted to pre-defined directories with the correct ownership and permissions. In this particular use-case, the database could be configured as a `systemd` service that relies on the `ebs-bootstrap.service` to succeed before attempting to start. This can be achieved by specifiying `ebs-boostrap.service` as a dependency in the `Requires=` and `After=` parameters.

```ini
[Unit]
Description=Example Database
Wants=network-online.target
Requires=ebs-bootstrap.service
After=network.target network-online.target ebs-bootstrap.service

[Service]
Type=forking
User=ec2-user
Group=ec2-user
ExecStart=/usr/bin/database start
ExecStop=/usr/bin/database stop

[Install]
WantedBy=multi-user.target
```

### `cloud-init`

`cloud-init` can be configured through EC2 User Data to write a config file to `/etc/ebs-boostrap/config.yml` through the `write_files` module. 

The NVMe Driver, for Nitro-based EC2 Instances, has an [established behaviour](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nvme-ebs-volumes.html) of dynamically renaming a block device, based on the order in which the device is attached to the EC2 instance. This order is unpredictable, thus it is recommended to label the volumes appropriately so that they can be referenced consistently in the `mounts` module. The `mounts` module is responsible for managing the `/etc/fstab` file, thus establishing a reliable mechanism for mounting external volumes at boot, independent of the block storage driver used by the EC2 instance.

```yaml
Resources:
  Instance:
    Type: AWS::EC2::Instance
  ...
  Volumes:
    - Device: /dev/sdb
      VolumeId: !Ref ExternalVolumeID
  UserData:
    Fn::Base64: !Sub
      - |+
        #cloud-config
        write_files:
          - content: |
              global:
                mode: healthcheck
              devices:
                /dev/sdb:
                  fs: ${FileSystem}
                  mount_point: /mnt/app
                  owner: ec2-user
                  group: ec2-user
                  permissions: 755
                  label: external-vol
            path: /etc/ebs-bootstrap/config.yml
        mounts:
          - [ "LABEL=external-vol", "/mnt/app", "${FileSystem}", "${MountOptions}", "0", "2" ]
      - FileSystem: ext4
        MountOptions: defaults,nofail,x-systemd.device-timeout=5
```

## Config

### `global`

#### `mode`

Specifies the mode that `ebs-bootstrap` operates in
  - `healthcheck`
    - Validate whether the state of a device matches its desired configuration
    - Returns an exit code of `0` 🟢, if no changes are detected
    - Returns an exit code of `1` 🔴, if changes are detected

### `devices[*]`

#### `fs`

The **file system** that the device has been formatted to
  - If an empty string is provided, all other device properties will be ignored

#### `mount_point`

The **mount point** that the device has been mounted to
  - If an empty string is provided, `owner`, `group` and `permissions` will be ignored

#### `owner`

The **user** that has been assigned ownership of the mount point
  - Supports both a user **ID** and the **name** of the user

#### `group`

The **group** that has been assigned ownership of the mount point
  - Supports both a group **ID** and the **name** of the group

#### `permissions`

The **permissions** that has been assigned to the mount point
  - Must be specified as a three digit octal: `755`, `644`, ...

#### `label`

The **label** assigned to the formatted device
  - Labels are constrained to the limitations of the underlying file system.
    - `ext4` file systems have a maximum label size of `16`
    - `xfs` file systems have a maximum label size of `12`
   
#### `mode`

Provide a device-level **override** of a global `mode` property

```yaml
global:
  mode: healthcheck
devices:
  /dev/xvdf:
    fs: "xfs"
    mount_point: "/ifmx/dev/root"
    owner: 1000
    group: 1000
    permissions: 755
    label: "external-vol"
    mode: healthcheck
```
