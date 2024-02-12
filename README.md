# ebs-bootstrap

<p align="center">
  <img src="assets/ebs-bootstrap.gif" alt="Demonstration of ebs-bootstrap" width="70%">
</p>

`ebs-bootstrap` is a tool that provides a **safe** and **as-code** approach for managing block devices on AWS EC2. It supports the following block device operations...

* **Format** a file system
* **Label** a file system
* **Resize** a file system
* **Mount** a block device
* Manage **ownership** and **permissions** of the mount point

Currently, the following file systems are supported for querying and modification...

* `ext4`
* `xfs`

Block device mappings can be [unpredictable](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html#device-name-limits) for AWS Nitro EC2 Instance types. `ebs-bootstrap` is equipped with the tools to recover the originally assigned block device mappings (`/dev/sd[a-z]`) from the dynamically allocated device names (`/dev/nvme[0-26]n1`) produced by **EBS** and **Instance Store** volumes.

## Build

`ebs-bootstrap` is a **statically-compiled** binary that can be built for both `linux/amd64` and `linux/arm64`. This process is facilitated by a multi-architecture Docker build process.

```bash
# Specific Architecture
[~] ./build/docker.sh --architecture arm64
[~] ls -la
ebs-bootstrap-linux-aarch64

# All Architectures
[~] ./build/docker.sh
[~] ls -la
ebs-bootstrap-linux-aarch64
ebs-bootstrap-linux-x86_64
```

## Installation

The latest binary of `ebs-bootstrap` can be downloaded from [GitHub Releases](https://github.com/reecetech/ebs-bootstrap/releases)

```
curl -L \
  -o /tmp/ebs-bootstrap \
  "https://github.com/reecetech/ebs-bootstrap/releases/latest/download/ebs-bootstrap-linux-$(uname -m)"
sudo install -m755 /tmp/ebs-bootstrap /usr/local/sbin/ebs-bootstrap
```

## Documentation

<a href="https://github.com/reecetech/ebs-bootstrap/wiki">
  <img src="assets/badges/github-wiki.svg">
</a>

## Use Cases

### `cloud-init`

The advent of Instance Store provided Nitro-enabled EC2 instances the ability to harness the power of high speed NVMe. For a stateful workload like a database, you might desire a EBS volume for critical data and an Instance Store volume for temporary tables. However, these Instance Store devices were ephemeral and had to be formatted and mounted on each startup cycle. 

From the perspective of a **sceptical** Platforms Engineer, you do not mind delegating the task of formatting and mounting ephemeral block devices to `ebs-bootstrap`. However, you personally draw the line on automation executing modifications to a stateful device, **without** the prior consent of a human. `ebs-bootstrap` empowers this Platform Engineer by allowing them to specify the execution mode, on a **device-by-device** basis: Instance Store (`force`) and EBS Volume (`healthcheck`)

<a href="https://github.com/reecetech/ebs-bootstrap/blob/main/examples/cloudformation.yml">
  <img src="assets/badges/cloudformation.svg">
</a>

> This CloudFormation template demonstrates the installation and configuration of `ebs-bootstrap` on an **Ubuntu Nitro EC2 Instance**. The instance has both an EBS and an Instance Store Volume attached, and the setup is performed using `cloud-init`.

On the first launch, `ebs-bootstrap` would refuse to perform any modifications to the EBS volume as it was assigned the `healthcheck` mode. However, we can temporarily override this behaviour with the `-mode=prompt` option. This allows the Platform Engineer to approve any suggested changes by `ebs-bootstrap`.

```

[~] sudo ebs-bootstrap -mode=prompt
🔵 Nitro NVMe detected: /dev/nvme1n1 -> /dev/sdb
🔵 Nitro NVMe detected: /dev/nvme2n1 -> /dev/sdh
🟠 Formatting larger disks can take several seconds ⌛
🟣 Would you like to format /dev/nvme1n1 to ext4? (y/n): y
⭐ Successfully formatted /dev/nvme1n1 to ext4
🟠 Certain file systems require that devices be unmounted prior to labeling
🟣 Would you like to label device /dev/nvme1n1 to 'stateful'? (y/n): y
⭐ Successfully labelled /dev/nvme1n1 to 'stateful'
...
🟣 Would you like to change ownership (1000:1000) of /mnt/ebs? (y/n): y
⭐ Successfully changed ownership (1000:1000) of /mnt/ebs
🟢 Passed all validation checks
```

By inspecting the output of `lsblk`, we can verify that `ebs-bootstrap` was able to recover the CloudFormation assigned block device mappings (`/dev/sdb` and `/dev/sdh`) from both EBS and Instance Store NVMe devices (`/dev/nvme1n1` and `/dev/nvme2n1`) and format/label/mount the respective devices.

```
[~] lsblk -o NAME,FSTYPE,MOUNTPOINT,LABEL,SIZE
NAME         FSTYPE MOUNTPOINT                  LABEL            SIZE
...
nvme0n1                                                            8G
├─nvme0n1p1  ext4   /                           cloudimg-rootfs  7.9G
├─nvme0n1p14                                                       4M
└─nvme0n1p15 vfat   /boot/efi                   UEFI             106M
nvme1n1      ext4   /mnt/ebs                    stateful          10G
nvme2n1      ext4   /mnt/instance-store         ephemeral         155G

[~] ls -la /mnt
total 16
drwxr-xr-x  4 root   root   4096 Jan  8 04:57 .
drwxr-xr-x 19 root   root   4096 Jan  8 04:36 ..
drwxr-xr-x  3 ubuntu ubuntu 4096 Jan  8 04:57 ebs
drwxr-xr-x  3 ubuntu ubuntu 4096 Jan  8 04:39 instance-store
```

The `mounts` module of `cloud-init` will create an entry in `/etc/fstab` for the EBS volume. The EBS volume, now labelled `stateful`, will be mounted to `/mnt/ebs`, by the operating-system, on **future reboots**. Despite device names being unstable because of the dynamic allocation behaviour of the Nitro NVMe driver, their respective labels remain stable across reboots.
```
[~] cat /etc/fstab
LABEL=cloudimg-rootfs	/	 ext4	defaults,discard	0 1
LABEL=UEFI	/boot/efi	vfat	umask=0077	0 1
LABEL=stateful	/mnt/ebs	ext4	defaults,nofail,x-systemd.device-timeout=5,comment=cloudconfig	0	2
```

### `systemd`

One way to utilise `ebs-bootstrap` is by employing a **oneshot** `systemd` service. This approach allows us to activate `ebs-bootstrap` during system startup, guaranteeing that any EBS or Instance Store volumes are formatted and mounted whenever the system is rebooted. This `systemd` unit file can either be generated during the [cloud-init](https://github.com/reecetech/ebs-bootstrap/blob/main/examples/cloudformation.yml#L97-L112) phase or more preferably baked into a Golden AMI.

```ini
[Unit]
Description=ebs-bootstrap
After=local-fs.target cloud-init.service  # Run after /etc/fstab (local-fs.target) and write_files (cloud-init.service)

[Service]
Type=oneshot
RemainAfterExit=true
StandardInput=null                        # Disables stdin to ensure error when prompted for an input
ExecStart=/usr/local/sbin/ebs-bootstrap
PrivateMounts=no                          # Prevents private mount namespaces
MountFlags=shared                         # Shares mounts to other processes

[Install]
WantedBy=multi-user.target
```

It is then possible to configure another `systemd` service to **only** start if the `ebs-bootstrap` service is successful. Certain databases support the ability to spread database chunks across multiple block devices that need to be mounted to pre-defined directories with the correct ownership and permissions enforced.

In this particular use-case, the database could be configured as a `systemd` service that relies on the `ebs-bootstrap.service` to succeed before attempting to start. This can be achieved by specifying `ebs-boostrap.service` as a dependency in the `Requires=` and `After=` parameters.

```ini
[Unit]
Description=example-database
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
