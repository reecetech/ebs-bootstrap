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

@test "format & mount loop with ext4" {
    run sudo $(command -v ebs-bootstrap) -config /tmp/ext4-bootstrap.yaml -mode force

    print_run_info
    [ "$status" -eq 0 ] &&
    [[ "$output" = *"Successfully formatted /dev/loop"*" to ext4"* ]] &&
    [[ "$output" = *"Successfully mounted /dev/loop"*" to /tmp/ext4"* ]]
}
