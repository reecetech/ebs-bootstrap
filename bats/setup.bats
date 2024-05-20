#!/usr/bin/env bats
# vim: set ft=sh sw=4 :

load helper_print-info

setup() {
    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    # make executables in root of the repo visible to PATH
    PATH="$DIR/../:$PATH"
}

@test "setup loopback device for ext4" {
    run bash -c '
      dd if=/dev/zero of=/tmp/ext4-loop bs=4096 count=32768 \
      && sudo losetup -f /tmp/ext4-loop \
      && losetup --associated /tmp/ext4-loop 2>&1 | tee /tmp/losetup \
      && grep "/tmp/ext4-loop" /tmp/losetup | cut -d ':' -f 1 > /tmp/loopdev-ext4
    '

    print_run_info
    [ "$status" -eq 0 ] &&
    [[ "$output" = *"/tmp/ext4-loop"* ]]
}

@test "setup loopback device for xfs" {
    run bash -c '
      dd if=/dev/zero of=/tmp/xfs-loop bs=4096 count=32768 \
      && sudo losetup -f /tmp/xfs-loop \
      && losetup --associated /tmp/xfs-loop 2>&1 | tee /tmp/losetup \
      && grep "/tmp/xfs-loop" /tmp/losetup | cut -d ':' -f 1 > /tmp/loopdev-xfs
    '

    print_run_info
    [ "$status" -eq 0 ] &&
    [[ "$output" = *"/tmp/xfs-loop"* ]]
}

@test "setup ext4 config" {
    echo """
---
devices:
  $(cat /tmp/loopdev-ext4):
    fs: ext4
    mountPoint: /tmp/ext4
""" > /tmp/ext4-bootstrap.yaml

    run mkdir /tmp/ext4
    [ "$status" -eq 0 ]
}

@test "setup xfs config" {
    echo """
---
devices:
  $(cat /tmp/loopdev-xfs):
    fs: xfs
    mountPoint: /tmp/xfs
""" > /tmp/xfs-bootstrap.yaml

    run mkdir /tmp/xfs
    [ "$status" -eq 0 ]
}
