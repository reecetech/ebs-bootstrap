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

@test "setup loopback device" {
    run bash -c '
      dd if=/dev/zero of=/tmp/fs bs=4096 count=32768 \
      && sudo losetup -f /tmp/fs \
      && losetup --associated /tmp/fs 2>&1 | tee /tmp/losetup \
      && grep "/tmp/fs" /tmp/losetup | cut -d ':' -f 1 > /tmp/loopdev
    '

    print_run_info
    [ "$status" -eq 0 ] &&
    [[ "$output" = *"/tmp/fs"* ]]
}

@test "setup ext4 config" {
    echo """
---
devices:
  $(cat /tmp/loopdev):
    fs: ext4
    mountPoint: /tmp/ext4
""" > /tmp/ext4-bootstrap.yaml

    run mkdir /tmp/ext4
    [ "$status" -eq 0 ]
}
